package certwatcher

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/k8tz/k8tz/pkg/version"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	verboseLogger *log.Logger
	warningLogger *log.Logger
	infoLogger    *log.Logger
	errorLogger   *log.Logger
)

type Watcher struct {
	TLSCertFile     string
	TLSKeyFile      string
	SecretName      string
	SecretNamespace string
	Verbose         bool
	clientset       kubernetes.Interface
}

func NewCertWatcher() *Watcher {
	return &Watcher{
		TLSCertFile:     "/run/secrets/tls/tls.crt",
		TLSKeyFile:      "/run/secrets/tls/tls.key",
		SecretName:      "k8tz-tls",
		SecretNamespace: "k8tz",
		Verbose:         false,
	}
}

func (w *Watcher) Start(kubeconfigFlag string) error {
	infoLogger.Println(version.DisplayVersion())

	if w.Verbose {
		verboseLogger.SetOutput(os.Stderr)
		verboseLogger.Printf("server=%+v", *w)
	}

	infoLogger.Printf("Watching kubenetes secret: %s/%s", w.SecretNamespace, w.SecretName)
	infoLogger.Printf("Syncing tls.crt on %s", w.TLSCertFile)
	infoLogger.Printf("Syncing tls.key on %s", w.TLSKeyFile)

	err := w.InitializeClientset(kubeconfigFlag)
	if err != nil {
		errorLogger.Printf("failed to setup connection with kubernetes api: %v", err)
		return fmt.Errorf("failed to setup connection with kubernetes api: %w", err)
	}

	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactoryWithOptions(w.clientset, 0, informers.WithNamespace(w.SecretNamespace))
	secretInformer := factory.Core().V1().Secrets().Informer()

	defer runtime.HandleCrash()

	go factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, secretInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for secretInformer caches to sync"))
		return fmt.Errorf("timed out waiting for secretInformer caches to sync")
	}

	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			infoLogger.Printf("receiving add event on secret %s/%s", w.SecretNamespace, w.SecretName)
			w.ProcessSecret(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			infoLogger.Printf("receiving update event on secret %s/%s", w.SecretNamespace, w.SecretName)
			w.ProcessSecret(newObj)
		},
	})

	<-stopper

	return nil
}

func getKubeconfig(kubeconfPath string) (*restclient.Config, error) {
	if kubeconfPath == "" {
		verboseLogger.Println("--kubeconfig not specified. Using the inClusterConfig. This might not work.")
		kubeconfig, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}

		warningLogger.Println("error creating inClusterConfig, falling back to default config.")
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
}

func (w *Watcher) InitializeClientset(kubeconfPath string) error {
	config, err := getKubeconfig(kubeconfPath)
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	clientset, _ := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %v", err)
	}

	w.clientset = clientset
	return nil
}

func overwriteFile(filepath string, filecontent []byte) {
	infoLogger.Printf("overwriting file %s", filepath)

	fileCrt, err := os.Create(filepath)
	if err != nil {
		errorLogger.Printf("error creating file: %s, error=%v", filepath, err)
	}

	defer fileCrt.Close()
	_, err = fileCrt.Write(filecontent)
	if err != nil {
		errorLogger.Printf("error writing file: %s, error=%v", filepath, err)
	}
}

func (w *Watcher) ProcessSecret(Obj interface{}) {
	secret := Obj.(*corev1.Secret)
	if secret.Namespace == w.SecretNamespace && secret.Name == w.SecretName {
		infoLogger.Printf("processing secret %s/%s ", secret.Namespace, secret.Name)

		overwriteFile(w.TLSCertFile, secret.Data["tls.crt"])
		overwriteFile(w.TLSKeyFile, secret.Data["tls.key"])
	}
}

func init() {
	verboseLogger = log.New(io.Discard, "VERBOSE: ", log.Ldate|log.Ltime|log.Lshortfile)
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warningLogger = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
