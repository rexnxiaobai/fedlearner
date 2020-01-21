/* Copyright 2020 The FedLearner Authors. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	etcdclient "github.com/coreos/etcd/clientv3"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/informers"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"

	crdclientset "github.com/bytedance/fedlearner/deploy/kubernetes_operator/pkg/client/clientset/versioned"
	crdinformers "github.com/bytedance/fedlearner/deploy/kubernetes_operator/pkg/client/informers/externalversions"
	"github.com/bytedance/fedlearner/deploy/kubernetes_operator/pkg/controller"
	"github.com/bytedance/fedlearner/deploy/kubernetes_operator/pkg/server"
	"github.com/bytedance/fedlearner/deploy/kubernetes_operator/pkg/servicediscovery"
)

var (
	master                      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	kubeConfig                  = flag.String("kubeConfig", "", "Path to a kube config. Only required if out-of-cluster.")
	peerURL                     = flag.String("peerURL", "localhost:8081", "The URL from/to which send worker pair request")
	port                        = flag.String("port", "8080", "The http port adapter uses")
	etcdURL                     = flag.String("etcdURL", "localhost:2379", "The URL of etcd backend for service discovery")
	workerNum                   = flag.Int("worker-num", 10, "Number of worker threads used by the fedlearner controller.")
	resyncInterval              = flag.Int("resync-interval", 30, "Informer resync interval in seconds.")
	namespace                   = flag.String("namespace", "default", "The namespace to which kubernetes_operator listen FLapps")
	assignWorkerPort            = flag.Bool("assign-worker-port", true, "Whether or not assign port to workers")
	workerPortRange             = flag.String("worker-port-range", "10000-30000", "The port range that controller use to assign ports")
	enableLeaderElection        = flag.Bool("leader-election", false, "Enable fedlearner kubernetes_operator leader election.")
	leaderElectionLockNamespace = flag.String("leader-election-lock-namespace", "fedlearner-system", "Namespace in which to create the Endpoints for leader election.")
	leaderElectionLockName      = flag.String("leader-election-lock-name", "fedlearner-kubernetes_operator-lock", "Name of the Endpoint for leader election.")
	leaderElectionLeaseDuration = flag.Duration("leader-election-lease-duration", 15*time.Second, "Leader election lease duration.")
	leaderElectionRenewDeadline = flag.Duration("leader-election-renew-deadline", 5*time.Second, "Leader election renew deadline.")
	leaderElectionRetryPeriod   = flag.Duration("leader-election-retry-period", 4*time.Second, "Leader election retry period.")
)

func buildConfig(masterURL string, kubeConfig string) (*rest.Config, error) {
	if kubeConfig != "" {
		return clientcmd.BuildConfigFromFlags(masterURL, kubeConfig)
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	if masterURL != "" {
		config.Host = masterURL
	}
	return config, nil
}

func buildClientset(masterURL string, kubeConfig string) (*clientset.Clientset, *crdclientset.Clientset, error) {
	config, err := buildConfig(masterURL, kubeConfig)
	if err != nil {
		return nil, nil, err
	}

	kubeClient, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	crdClient, err := crdclientset.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return kubeClient, crdClient, err
}

func startLeaderElection(
	kubeClient *clientset.Clientset,
	startCh chan struct{},
	stopCh chan struct{},
) {
	hostName, err := os.Hostname()
	if err != nil {
		klog.Error("failed to get hostname")
		return
	}
	hostName = hostName + "_" + string(uuid.NewUUID())

	resourceLock, err := resourcelock.New(resourcelock.EndpointsResourceLock,
		*leaderElectionLockNamespace,
		*leaderElectionLockName,
		kubeClient.CoreV1(),
		nil,
		resourcelock.ResourceLockConfig{
			Identity:      hostName,
			EventRecorder: &record.FakeRecorder{},
		})
	if err != nil {
		return
	}

	electionCfg := leaderelection.LeaderElectionConfig{
		Lock:          resourceLock,
		LeaseDuration: *leaderElectionLeaseDuration,
		RenewDeadline: *leaderElectionRenewDeadline,
		RetryPeriod:   *leaderElectionRetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				close(startCh)
			},
			OnStoppedLeading: func() {
				close(stopCh)
			},
		},
	}
	elector, err := leaderelection.NewLeaderElector(electionCfg)
	if err != nil {
		klog.Fatal(err)
	}

	go elector.Run(context.Background())
}

func buildETCDClient(etcdURL string) (*etcdclient.Client, error) {
	cfg := etcdclient.Config{
		Endpoints:   strings.Split(etcdURL, ","),
		DialTimeout: 2 * time.Second,
	}
	return etcdclient.New(cfg)
}

func main() {
	flag.Parse()

	kubeClient, crdClient, err := buildClientset(*master, *kubeConfig)
	if err != nil {
		klog.Fatalf("failed to build clientset, err = %v", err)
	}

	etcdClient, err := buildETCDClient(*etcdURL)
	if err != nil {
		klog.Fatalf("failed to build etcdclient, err = %v", err)
	}

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	stopCh := make(chan struct{}, 1)
	startCh := make(chan struct{}, 1)

	if *enableLeaderElection {
		startLeaderElection(kubeClient, startCh, stopCh)
	}

	klog.Info("starting the fedlearner kubernetes_operator")

	kubeInformerFactory := informers.NewSharedInformerFactory(kubeClient, time.Duration(*resyncInterval)*time.Second)
	crdInformerFactory := crdinformers.NewSharedInformerFactory(crdClient, time.Duration(*resyncInterval)*time.Second)

	appEventHandler := controller.NewappEventHandler(*peerURL, *namespace, crdClient)
	flController := controller.NewFLController(*namespace, *assignWorkerPort, *workerPortRange, kubeClient, crdClient, kubeInformerFactory, crdInformerFactory, appEventHandler, stopCh)
	sdController := servicediscovery.NewSDController(*namespace, kubeClient, kubeInformerFactory, etcdClient, stopCh)

	go kubeInformerFactory.Start(stopCh)
	go crdInformerFactory.Start(stopCh)

	if *enableLeaderElection {
		klog.Info("waiting to be elected leader before starting application controller goroutines")
		<-startCh
	}
	klog.Info("starting application controller goroutines")
	if err := flController.Start(*workerNum); err != nil {
		klog.Fatal(err)
	}
	if err := sdController.Start(*workerNum); err != nil {
		klog.Fatal(err)
	}

	go func() {
		klog.Infof("starting adapter listening %v", *port)
		server.ServeGrpc("0.0.0.0", *port, appEventHandler)
	}()

	select {
	case <-signalCh:
		close(stopCh)
	case <-stopCh:
	}

	klog.Info("shutting down the fedlearner kubernetes_operator")
	flController.Stop()
}
