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

package cmd

import (
	certwatcher "github.com/k8tz/k8tz/pkg/cert-watcher"
	"github.com/spf13/cobra"
)

var certWatcher = certwatcher.NewCertWatcher()

var certWatcherCmd = &cobra.Command{
	Use:    "cert-watcher",
	Hidden: true,
	Short:  "Starts k8tz's certificate watcher",
	Long: `Starts k8tz's certificate watcher.
	
The watcher will listen to Kubernetes Secret that containing TLS certificate
for k8tz and make sure when changes are occured, it will overwrite the current
TLS certificate deployed on k8tz's Pod.`,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(certWatcher.Start(kubeConfigFile))
	},
}

func init() {
	rootCmd.AddCommand(certWatcherCmd)

	certWatcherCmd.Flags().StringVar(&certWatcher.TLSCertFile, "tls-crt", certWatcher.TLSCertFile, "TLS Certificate file")
	certWatcherCmd.Flags().StringVar(&certWatcher.TLSKeyFile, "tls-key", certWatcher.TLSKeyFile, "TLS Key file")
	certWatcherCmd.Flags().StringVar(&certWatcher.SecretName, "secret-name", certWatcher.SecretName, "Kubernetes secret containing TLS Certificate and TLS Key")
	certWatcherCmd.Flags().StringVar(&certWatcher.SecretNamespace, "secret-namespace", certWatcher.SecretNamespace, "Kubernetes secret namespace containing TLS Certificate and TLS Key")
	certWatcherCmd.Flags().BoolVar(&certWatcher.Verbose, "verbose", certWatcher.Verbose, "Print more verbose logs for debugging")
}
