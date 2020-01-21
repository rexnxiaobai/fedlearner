# Copyright 2020 The FedLearner Authors. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# coding: utf-8

import tensorflow.compat.v1 as tf
import logging
from tensorflow.compat.v1.train import Optimizer
from tensorflow.compat.v1.estimator import ModeKeys

class FLModel(object):
    def __init__(self, role, bridge, example_ids, exporting=False):
        self._role = role
        self._bridge = bridge
        self._example_ids = example_ids
        self._exporting = exporting

        self._train_ops = []
        self._recvs = []
        self._sends = []
        self._outputs = []

    @property
    def train_ops(self):
        return self._train_ops

    @property
    def sends(self):
        return [(n, t) for n, t, _ in self._sends]

    @property
    def recvs(self):
        return [(n, t) for n, t, _ in self._recvs]

    def verify_example_ids(self):
        tensor = tf.strings.to_hash_bucket_fast(
            self._example_ids, 2**31-1)
        if self._role == 'leader':
            self.send('_verify_example_ids', tensor)
        else:
            recv_tensor = self.recv('_verify_example_ids', tensor.dtype)
            op = tf.assert_equal(tensor, recv_tensor)
            self._train_ops.append(op)

    def send(self, name, tensor, require_grad=False):
        with tf.control_dependencies([self._example_ids]):
            op = self._bridge.send_op(name, tensor)
        self._train_ops.append(op)
        self._sends.append((name, tensor, require_grad))
        if require_grad:
            return self.recv(name+'_grad', tensor.dtype)
        return None

    def recv(self, name, dtype=tf.float32, require_grad=False):
        with tf.control_dependencies([self._example_ids]):
            tensor = self._bridge.receive_op(name, dtype)
        self._recvs.append((name, tensor, require_grad))
        return tensor

    def minimize(self, optimizer, loss, global_step=None, var_list=None,
                 gate_gradients=Optimizer.GATE_OP,
                 aggregation_method=None,
                 colocate_gradients_with_ops=False, name=None,
                 grad_loss=None):
        recv_grads = [i for i in self._recvs if i[2]]

        if var_list is None:
            var_list = \
                tf.get_collection(tf.GraphKeys.TRAINABLE_VARIABLES) + \
                tf.get_collection(tf.GraphKeys.TRAINABLE_RESOURCE_VARIABLES)
        var_list = [v for _, v, _ in recv_grads] + var_list

        grads_and_vars = optimizer.compute_gradients(
            loss, var_list=var_list, gate_gradients=gate_gradients,
            aggregation_method=aggregation_method,
            colocate_gradients_with_ops=colocate_gradients_with_ops,
            grad_loss=grad_loss)

        send_grads = grads_and_vars[:len(recv_grads)]
        for (n, _, _), (grad, _) in zip(recv_grads, send_grads):
            if grad is not None:
                self.send(n+'_grad', grad)

        train_op = optimizer.apply_gradients(
            grads_and_vars[len(recv_grads):],
            global_step=global_step, name=name)

        return train_op

    def make_spec(self,
                  mode,
                  predictions=None,
                  loss=None,
                  train_op=None,
                  eval_metric_ops=None,
                  export_outputs=None,
                  training_chief_hooks=None,
                  training_hooks=None,
                  scaffold=None,
                  evaluation_hooks=None,
                  prediction_hooks=None):
        if isinstance(predictions, tf.Tensor):
            predictions = {'output': predictions}
        if mode == ModeKeys.TRAIN:
            train_op = tf.group([train_op] + self._train_ops)
        return tf.estimator.EstimatorSpec(
            mode=mode,
            predictions=predictions,
            loss=loss,
            train_op=train_op,
            eval_metric_ops=eval_metric_ops,
            export_outputs=export_outputs,
            training_chief_hooks=training_chief_hooks,
            training_hooks=training_hooks,
            scaffold=scaffold,
            evaluation_hooks=evaluation_hooks,
            prediction_hooks=prediction_hooks)


class FLEstimator(object):
    def __init__(self, model_fn, bridge, trainer_master,
                 role, worker_rank=0, cluster_spec=None):
        self._model_fn = model_fn
        self._bridge = bridge
        self._trainer_master = trainer_master
        self._role = role
        self._worker_rank = worker_rank
        self._cluster_spec = cluster_spec

    def train(self, input_fn,
              checkpoint_path=None,
              save_checkpoint_steps=None):
        if self._cluster_spec is not None:
            device_fn = tf.train.replica_device_setter(
                worker_device="/job:worker/task:%d" % self._worker_rank,
                merge_devices=True, cluster=self._cluster_spec)
            cluster_def = self._cluster_spec.as_cluster_def()
            local_address = self._cluster_spec.job_tasks(
                'worker')[self._worker_rank]
            server = tf.train.Server(tf.train.ClusterSpec(
                {'local': {0: local_address}}),
                job_name='local', task_index=0)
            target = server.target
            logging.debug('local_address %s worker_rank %d target %s',
                          local_address, self._worker_rank, target)
            # 'grpc://' + local_address
        else:
            device_fn = None
            cluster_def = None
            target = None

        config = tf.ConfigProto(cluster_def=cluster_def)
        config.experimental.share_session_state_in_clusterspec_propagation \
            = True
        tf.config.set_soft_device_placement(False)

        with tf.Graph().as_default() as g:
            with tf.device(device_fn):
                features, labels = input_fn(
                    self._bridge, self._trainer_master)
                model = FLModel(self._role, self._bridge,
                                features['example_id'])
                spec = self._model_fn(model, features, labels, ModeKeys.TRAIN)

            self._bridge.connect()
            with tf.train.MonitoredTrainingSession(
                    master=target,
                    config=config,
                    is_chief=(self._worker_rank == 0),
                    checkpoint_dir=checkpoint_path,
                    save_checkpoint_steps=save_checkpoint_steps,
                    hooks=spec.training_hooks) as sess:
                iter_id = 0
                while not sess.should_stop():
                    self._bridge.start(iter_id)
                    sess.run(spec.train_op, feed_dict={})
                    self._bridge.commit()
                    iter_id += 1
            self._bridge.terminate()

    def export_saved_model(self, export_dir_base,
                           serving_input_receiver_fn,
                           checkpoint_path=None):
        with tf.Graph().as_default() as g:
            receiver = serving_input_receiver_fn()
            model = FLModel(self._role, self._bridge,
                        receiver.features.get('example_id', None),
                        exporting=True)
            spec = self._model_fn(model, receiver.features,
                                  None, ModeKeys.PREDICT)
            assert not model.sends, "Exported model cannot send"
            assert not model.recvs, "Exported model cannot receive"

            with tf.Session() as sess:
                saver_for_restore = tf.train.Saver(sharded=True)
                saver_for_restore.restore(
                    sess, tf.train.latest_checkpoint(checkpoint_path))
                tf.saved_model.simple_save(
                    sess, export_dir_base, receiver.receiver_tensors,
                    spec.predictions, None)

            return export_dir_base
