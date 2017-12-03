package main

import (
	"flag"
	"github.com/golang/glog"
	clientset "github.com/arnaudmz/kaos/pkg/client/clientset/versioned"
	informers "github.com/arnaudmz/kaos/pkg/client/informers/externalversions"
	"github.com/arnaudmz/kaos/pkg/signals"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
)

func main() {
	flag.Parse()
	stopCh := signals.SetupSignalHandler()

	cfg, err := clientcmd.BuildConfigFromFlags(*master, *kuberconfig)
	if err != nil {
		glog.Fatalf("Error building kubeconfig: %v", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	kaosruleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		glog.Fatalf("Error building kaosrule clientset: %v", err)
	}

	kaosruleInformerFactory := informers.NewSharedInformerFactory(kaosruleClient, time.Second*30)

	controller := NewController(kubeClient, kaosruleClient, kaosruleInformerFactory)

	go kaosruleInformerFactory.Start(stopCh)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}
}
