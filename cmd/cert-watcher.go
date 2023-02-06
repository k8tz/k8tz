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
