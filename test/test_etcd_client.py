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

import os

import unittest
import threading
import time

from fedlearner.common import etcd_client

class TestEtcdClient(unittest.TestCase):
    def test_etcd_op(self):
        cli = etcd_client.EtcdClient('test_cluster', '10.8.163.165:4578', 'data_source_a')
        cli.delete('fl_key')
        cli.set_data('fl_key', 'fl_value')
        self.assertEqual(cli.get_data('fl_key'), b'fl_value')
        self.assertFalse(cli.cas('fl_key', 'fl_value1', 'fl_value2'))
        self.assertTrue(cli.cas('fl_key', 'fl_value', 'fl_value1'))
        self.assertEqual(cli.get_data('fl_key'), b'fl_value1')

        goahead = False
        def thread_routine():
            cli.set_data('fl_key', 'fl_value2')
            self.assertEqual(cli.get_data('fl_key'), b'fl_value2')

        eiter, cancel = cli.watch_key('fl_key')
        other = threading.Thread(target=thread_routine)
        other.start()
        for e in eiter:
            self.assertEqual(e.key, b'/data_source_a/fl_key')
            self.assertEqual(e.value, b'fl_value2')
            cancel()
        other.join()
        cli.destory_client_pool()

if __name__ == '__main__':
        unittest.main()
