package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/k8tz/k8tz/pkg/watcher"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	client     *kubernetes.Clientset
	kubeConfig *rest.Config
	err        error

	infoLogger  *log.Logger
	errorLogger *log.Logger

	secretWatcher = watcher.NewWatcher()
)

var watcherCmd = &cobra.Command{
	Use:    "watcher",
	Hidden: true,
	Short:  "Starts Kubernetes Secret Watcher",
	Long:   ``,
	Run: func(cmd *cobra.Command, args []string) {
		// get kubeconfig, either from kubeconfig arg or incluster kubeconfig
		if kubeConfigFile == "" {
			kubeConfig, err = rest.InClusterConfig()
			if err != nil {
				errorLogger.Fatal(err.Error())
			}
		} else {
			kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
			if err != nil {
				errorLogger.Fatal(err.Error())
			}
		}

		// cteate new kubernetes client
		client, err = kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			errorLogger.Fatal(err.Error())
		}

		// stop signal for the informer
		stopper := make(chan struct{})
		defer close(stopper)

		// setup shared informers
		factory := informers.NewSharedInformerFactoryWithOptions(client, 0, informers.WithNamespace(secretWatcher.SecretNamespace))
		secretInformer := factory.Core().V1().Secrets().Informer()

		// handle runtime crash
		defer runtime.HandleCrash()

		// start informer ->
		go factory.Start(stopper)

		// start to sync and call list
		if !cache.WaitForCacheSync(stopper, secretInformer.HasSynced) {
			runtime.HandleError(fmt.Errorf("timed out waiting for secretInformer caches to sync"))
			return
		}

		// add handler for secret event
		secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				processSecret(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				processSecret(newObj)
			},
		})

		// handle stop signal
		<-stopper
	},
}

func init() {
	rootCmd.AddCommand(watcherCmd)

	watcherCmd.Flags().StringVar(&secretWatcher.TLSCertFile, "tls-crt", secretWatcher.TLSCertFile, "TLS Certificate file")
	watcherCmd.Flags().StringVar(&secretWatcher.TLSKeyFile, "tls-key", secretWatcher.TLSKeyFile, "TLS Key file")
	watcherCmd.Flags().StringVar(&secretWatcher.SecretName, "secret-name", secretWatcher.SecretName, "Kubernetes secret containing TLS Certificate and TLS Key")
	watcherCmd.Flags().StringVar(&secretWatcher.SecretNamespace, "secret-namespace", secretWatcher.SecretNamespace, "Kubernetes secret namespace containing TLS Certificate and TLS Key")
	watcherCmd.Flags().BoolVar(&secretWatcher.Verbose, "verbose", secretWatcher.Verbose, "Print more verbose logs for debugging")

	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func processSecret(Obj interface{}) {
	secret := Obj.(*corev1.Secret)
	if secret.Namespace == secretWatcher.SecretNamespace && secret.Name == secretWatcher.SecretName {
		infoLogger.Printf("processing secret %s/%s ", secret.Namespace, secret.Name)

		infoLogger.Printf("writing secret tls.crt to %s", secretWatcher.TLSCertFile)
		fileCrt, err := os.Create(secretWatcher.TLSCertFile)
		if err != nil {
			return
		}
		defer fileCrt.Close()
		fileCrt.Write(secret.Data["tls.crt"])

		infoLogger.Printf("writing secret tls.key to %s", secretWatcher.TLSKeyFile)
		fileKey, err := os.Create(secretWatcher.TLSKeyFile)
		if err != nil {
			return
		}
		defer fileKey.Close()
		fileKey.Write(secret.Data["tls.key"])
	}
}
