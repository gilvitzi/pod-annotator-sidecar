package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"k8s.io/klog"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	podNameEnvVarKey   = "POD_NAME"
	namespaceEnvVarKey = "POD_NAMESPACE"
)

var (
	podName            string
	namespace          string
	annotationsDirPath string
	annotationsPrefix  string
)

func loadConfigLocal() (*rest.Config, error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	klog.V(5).Infof("Trying to get local config from %s", kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.V(5).Infof("Failed to get local config from %s", kubeconfig)
	}
	return config, err
}

func loadConfig() (*rest.Config, error) {
	// Try getting cluster config first, if not available get local config
	klog.V(5).Info("Trying to load InCluster config")

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.V(5).Info("Failed to get InCluster config")
		config, err = loadConfigLocal()
	} else {
		klog.V(5).Info("Got InCluster config successfully")
	}

	return config, err
}

func parseFlags() {
	defer klog.Flush()
	klog.InitFlags(nil)
	flag.StringVar(&podName, "pod-name", "", "the name of the current running pod")
	flag.StringVar(&namespace, "namespace", "", "the name of the current running pod")
	flag.StringVar(&annotationsDirPath, "dir-path", "", "The agent will monitor this dir and for each file will create an annotation filename=<file contents>")
	flag.StringVar(&annotationsPrefix, "prefix", "", "If given will prepend this prefix to every annotation key i.e --prefix \"my.app.com/\" will generate annotations in the format: my.app.com/filename=file_contents")

	flag.Set("alsologtostderr", "true")

	flag.Parse()

	if podName == "" {
		podName = os.Getenv(podNameEnvVarKey)

		if podName == "" {
			klog.Fatalf("Failed to retrive podName. Please set the env var %s or supply the argument --pod-name", podNameEnvVarKey)
		}
	}

	if namespace == "" {
		namespace = os.Getenv(namespaceEnvVarKey)

		if namespace == "" {
			klog.Fatalf("Failed to retrive namespace. Please set the env var %s or supply the argument --namespace", namespaceEnvVarKey)
		}
	}

	if annotationsDirPath == "" {
		klog.Fatalf("Missing -dir-path flag")
	}
}

func listenToTerminationSignals() chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	return sig
}

func annotatePodOnFileChanges(wg *sync.WaitGroup, config *rest.Config, annotationsDirPath string, changedFiles chan string) {
	defer wg.Done()
	for range changedFiles {
		collectDataAndAnnotatePod(config, annotationsDirPath)
	}

	klog.V(5).Info("annotatePodOnFileChanges finished")
}

func main() {
	wg := &sync.WaitGroup{}
	termSignals := listenToTerminationSignals()

	parseFlags()

	klog.V(0).Infof("Started with podName: %s namespace %s", podName, namespace)

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config %v", err)
	}

	// Load initial data from directory
	collectDataAndAnnotatePod(config, annotationsDirPath)

	watcher, changedFiles := startWatchingDirForChanges(wg, annotationsDirPath)
	wg.Add(1)
	go annotatePodOnFileChanges(wg, config, annotationsDirPath, changedFiles)

	<-termSignals

	// GracefulShutdown
	klog.V(3).Info("Shutdown started..")
	klog.V(5).Info("Closing watcher..")
	watcher.Close()

	klog.V(5).Info("Closing changedFiles channel..")
	close(changedFiles)
	wg.Wait()
	klog.V(3).Info("Shutdown finsihed")
}
