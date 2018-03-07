package main

import (
	"flag"
	clientset "github.com/arnaudmz/kaos/pkg/client/clientset/versioned"
	informers "github.com/arnaudmz/kaos/pkg/client/informers/externalversions"
	"github.com/arnaudmz/kaos/pkg/metrics"
	"github.com/arnaudmz/kaos/pkg/signals"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"goji.io"
	"goji.io/pat"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"net/http"
	"time"
)

var (
	kuberconfig = flag.String("kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	master      = flag.String("master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	version     = "No version"
	timestamp   = "0.0"
  // KilledPods tracks per-rule killed pods count
	KilledPods  = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "kaos_killed_pods_total",
		Help: "Killed Pods summary",
	},
		[]string{"namespace", "kaosrule"})
	appInfo = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "app_info",
		Help:        "Information about application",
		ConstLabels: prometheus.Labels{"version": version, "build_timestamp": timestamp},
	})
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(KilledPods)
	prometheus.MustRegister(appInfo)
	appInfo.Set(1)
}

func debugHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s", r.Method, r.URL)
			next.ServeHTTP(w, r)
		})
}

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

	mux := goji.NewMux()
	mux.Use(debugHandler)
	mux.HandleFunc(pat.Get("/healthz"), metrics.Healthz)
	mux.Handle(pat.Get("/metrics"), promhttp.Handler())

	go http.ListenAndServe(":8080", mux)

	if err = controller.Run(2, stopCh); err != nil {
		glog.Fatalf("Error running controller: %s", err.Error())
	}

}
