/*
Copyright © 2021 Yonatan Kahana

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
	"strings"

	"github.com/k8tz/k8tz/pkg/admission"

	"github.com/spf13/cobra"
	cliflag "k8s.io/component-base/cli/flag"
)

var webhook = admission.NewAdmissionServer()

var webhookCmd = &cobra.Command{
	Use:    "webhook",
	Hidden: true,
	Short:  "Starts Kubernetes Mutating Admission Webhook Server",
	Long: `Starts k8tz's Kubernetes mutating admission controller webhook server.

The webhook will listen to requests from Kubernetes and reply
with list of json patches for kubernetes to perform in order
to have the timezone injected to containers.

TLS certificate and private key is required to receive requests
from kubernetes controllers. The certificate should have SAN
and DNS that reflects the webhooks service FQDN, e.g:
webhook.k8tz.svc.

Injection defaults can be controlled via flags such as '-t'
to change the default timezone; or '-s' to change the injection
strategy.`,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(webhook.Start(kubeConfigFile))
	},
}

func init() {
	rootCmd.AddCommand(webhookCmd)

	webhookCmd.Flags().StringVar(&webhook.TLSCertFile, "tls-crt", webhook.TLSCertFile, "TLS Certificate file")
	webhookCmd.Flags().StringVar(&webhook.TLSKeyFile, "tls-key", webhook.TLSKeyFile, "TLS Key file")
	tlsCipherPreferredValues := cliflag.PreferredTLSCipherNames()
	tlsCipherInsecureValues := cliflag.InsecureTLSCipherNames()
	webhookCmd.Flags().StringSliceVar(&webhook.TLSCipherSuites, "tls-cipher-suites", webhook.TLSCipherSuites,
		"Comma-separated list of cipher suites for the server. "+
			"If omitted, the default Go cipher suites will be used. \n"+
			"Preferred values: "+strings.Join(tlsCipherPreferredValues, ", ")+". \n"+
			"Insecure values: "+strings.Join(tlsCipherInsecureValues, ", ")+".")
	tlsPossibleVersions := cliflag.TLSPossibleVersions()
	webhookCmd.Flags().StringVar(&webhook.TLSMinVersion, "tls-min-version", webhook.TLSMinVersion,
		"Minimum TLS version supported. "+
			"Possible values: "+strings.Join(tlsPossibleVersions, ", "))
	webhookCmd.Flags().StringVar(&webhook.Address, "addr", webhook.Address, "Webhook bind address")
	webhookCmd.Flags().StringVarP(&webhook.Handler.DefaultTimezone, "timezone", "t", webhook.Handler.DefaultTimezone, "Default timezone if not specified explicitly")
	webhookCmd.Flags().StringVar(&webhook.Handler.ContainerName, "container-name", webhook.Handler.ContainerName, "initContainer name")
	webhookCmd.Flags().StringVar(&webhook.Handler.BootstrapImage, "bootstrap-image", webhook.Handler.BootstrapImage, "initContainer bootstrap image")
	webhookCmd.Flags().StringVar(&webhook.Handler.BootstrapContainerImagePullPolicy, "container-imagepullpolicy", webhook.Handler.BootstrapContainerImagePullPolicy, "initContainer bootstrap imagePullPolicy")
	webhookCmd.Flags().BoolVar(&webhook.Handler.BootstrapVerbose, "bootstrap-verbose", webhook.Handler.BootstrapVerbose, "Print more verbose logs inside the bootstrap initContainer for debugging")
	webhookCmd.Flags().StringVar(&webhook.Handler.BootstrapContainerResources, "bootstrap-resources", webhook.Handler.BootstrapContainerResources, "initContainer compute resources in JSON format")
	webhookCmd.Flags().StringVar(&webhook.Handler.HostPathPrefix, "hostPathPrefix", webhook.Handler.HostPathPrefix, "Location of zoneinfo on host machines")
	webhookCmd.Flags().StringVar(&webhook.Handler.LocalTimePath, "localTimePath", webhook.Handler.LocalTimePath, "Mount path for TZif file on containers")
	webhookCmd.Flags().StringVarP((*string)(&webhook.Handler.DefaultInjectionStrategy), "injection-strategy", "s", string(webhook.Handler.DefaultInjectionStrategy), "Default injection strategy if not specified explicitly (hostPath/initContainer)")
	webhookCmd.Flags().BoolVar(&webhook.Handler.InjectByDefault, "inject", webhook.Handler.InjectByDefault, "Whether injection is enabled by default or should be requested by annotation")
	webhookCmd.Flags().BoolVar(&webhook.Handler.CronJobTimeZone, "cronJobTimeZone", webhook.Handler.CronJobTimeZone, "Enable CronJob injection. Requires kubernetes >=1.24.0-beta.0 and the 'CronJobTimeZone' feature gate enabled (alpha)")
	webhookCmd.Flags().BoolVar(&webhook.Verbose, "verbose", webhook.Verbose, "Print more verbose logs for debugging")
}
