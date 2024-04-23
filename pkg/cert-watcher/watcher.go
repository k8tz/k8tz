/*
Copyright © 2023 Andika Ahmad Ramadhan

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package certwatcher

import (
	"context"
	"fmt"
	"os"

	k8tz "github.com/k8tz/k8tz/pkg"
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

type CertWatcher struct {
	TLSCertFile     string
	TLSKeyFile      string
	SecretName      string
	SecretNamespace string
	Verbose         bool
	clientset       kubernetes.Interface

	ctx    context.Context
	cancel context.CancelFunc
}

func NewCertWatcher() *CertWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &CertWatcher{
		TLSCertFile:     "/run/secrets/tls/tls.crt",
		TLSKeyFile:      "/run/secrets/tls/tls.key",
		SecretName:      "k8tz-tls",
		SecretNamespace: "k8tz",
		Verbose:         false,
		clientset:       nil,

		ctx:    ctx,
		cancel: cancel,
	}
}

func (w *CertWatcher) Start(kubeconfigFlag string) error {
	k8tz.InfoLogger.Println(version.DisplayVersion())

	if w.Verbose {
		k8tz.VerboseLogger.SetOutput(os.Stderr)
		k8tz.VerboseLogger.Printf("cert-watcher=%+v", *w)
	}

	k8tz.InfoLogger.Printf("Watching kubenetes secret: %s/%s", w.SecretNamespace, w.SecretName)
	k8tz.InfoLogger.Printf("Syncing tls.crt on %s", w.TLSCertFile)
	k8tz.InfoLogger.Printf("Syncing tls.key on %s", w.TLSKeyFile)

	err := w.initializeClientset(kubeconfigFlag)
	if err != nil {
		k8tz.ErrorLogger.Printf("failed to setup connection with kubernetes api: %v", err)
		return fmt.Errorf("failed to setup connection with kubernetes api: %w", err)
	}

	return w.startWatcher()
}

func (w *CertWatcher) startWatcher() error {
	factory := informers.NewSharedInformerFactoryWithOptions(
		w.clientset, 0,
		informers.WithNamespace(w.SecretNamespace),
	)
	secretInformer := factory.Core().V1().Secrets().Informer()

	defer runtime.HandleCrash()

	go factory.Start(w.ctx.Done())

	if !cache.WaitForCacheSync(w.ctx.Done(), secretInformer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for secretInformer caches to sync"))
		return fmt.Errorf("timed out waiting for secretInformer caches to sync")
	}

	_, err := secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			k8tz.InfoLogger.Printf("receiving add event on secret %s/%s", w.SecretNamespace, w.SecretName)
			w.ProcessSecret(obj.(*corev1.Secret))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			k8tz.InfoLogger.Printf("receiving update event on secret %s/%s", w.SecretNamespace, w.SecretName)
			w.ProcessSecret(newObj.(*corev1.Secret))
		},
	})
	if err != nil {
		return fmt.Errorf("failed to register EventHandler for secretInformer")
	}

	<-w.ctx.Done()

	return nil
}

func getKubeconfig(kubeconfPath string) (*restclient.Config, error) {
	if kubeconfPath == "" {
		k8tz.VerboseLogger.Println("--kubeconfig not specified. Using the inClusterConfig. This might not work.")
		kubeconfig, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}

		k8tz.WarningLogger.Println("error creating inClusterConfig, falling back to default config.")
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
}

func (w *CertWatcher) initializeClientset(kubeconfPath string) error {
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
	k8tz.InfoLogger.Printf("overwriting file %s", filepath)

	fileCrt, err := os.Create(filepath)
	if err != nil {
		k8tz.ErrorLogger.Panicf("error creating file: %s, error=%v", filepath, err)
	}

	defer fileCrt.Close()
	_, err = fileCrt.Write(filecontent)
	if err != nil {
		k8tz.ErrorLogger.Panicf("error writing file: %s, error=%v", filepath, err)
	}
}

func (w *CertWatcher) ProcessSecret(secret *corev1.Secret) {
	if secret.Namespace == w.SecretNamespace && secret.Name == w.SecretName {
		k8tz.InfoLogger.Printf("processing secret %s/%s ", secret.Namespace, secret.Name)

		overwriteFile(w.TLSCertFile, secret.Data["tls.crt"])
		overwriteFile(w.TLSKeyFile, secret.Data["tls.key"])
	}
}
