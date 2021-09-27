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
	"os"

	"github.com/spf13/cobra"
)

var kubeConfigFile = ""

var rootCmd = &cobra.Command{
	Use:   "k8tz",
	Short: "Inject timezones into kubernetes pods",
	Long: `Inject timezones into kubernetes Pods from the command-line interface
or with an admission controller. 

Containers does not inherit timezones from host machines
and have only access to the clock from the kernel-space.
The default timezone for most images is UTC. With k8tz it
is easy to standardize selected timezone across pods and
namespaces automatically without any changes to their
deployments.

For more information: https://k8tz.io`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()

	rootCmd.PersistentFlags().StringVar(&kubeConfigFile, "kube-config", kubeConfigFile, "Path to kubeconfig file")
}
