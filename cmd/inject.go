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
	"errors"
	"fmt"
	"os"

	"github.com/k8tz/k8tz/pkg/inject"
	"github.com/k8tz/k8tz/pkg/version"
	"github.com/spf13/cobra"
)

var patchGenerator = inject.NewPatchGenerator()

var injectCmd = &cobra.Command{
	Use:     "inject <input [...]>",
	Aliases: []string{"i"},
	Short:   "Inject timezone to yaml kubernetes resources",
	Long: `Inject timezone to yaml kubernetes resources and print mutated resources back to standard output. 

Input may be '-' for stdin, path to file or http/https url.

Examples:
# Inject Europe/Amsterdam timezone to all the deployments in the current namespace
kubectl get deploy | k8tz inject --timezone=Europe/Amsterdam - | kubectl apply -f -

# Create pod with UTC timezone with hostPath strategy from a yaml file
k8tz i -tUTC --strategy=hostPath examples/test-pod.yaml | kubectl create -f -

# Create pod with New York timezone from URL with custom private registry
k8tz inject --image=registry.example.com/myrepo/k8tz:` + version.Version() + ` -tAmerica/New_York https://github.com/k8tz/k8tz/.../examples/test-pod.yaml | kubectl apply -f -

Injection is applicable on Pods, Deployments and Lists that contains them.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("you must specify at least one input")
		}

		inputs, err := inject.ArgumentsToInputs(args)
		if err != nil {
			return fmt.Errorf("failed to open inputs from arguments: %w", err)
		}

		transformer := &inject.Transformer{
			PatchGenerator: patchGenerator,
			Inputs:         inputs,
			Output:         os.Stdout,
		}

		return transformer.Transform()
	},
}

func init() {
	rootCmd.AddCommand(injectCmd)

	injectCmd.Flags().StringVarP(&patchGenerator.Timezone, "timezone", "t", patchGenerator.Timezone, "Default timezone if not specified explicitly")
	injectCmd.Flags().StringVar(&patchGenerator.InitContainerName, "name", patchGenerator.InitContainerName, "initContainer name")
	injectCmd.Flags().StringVar(&patchGenerator.InitContainerImagePullPolicy, "imagePullPolicy", patchGenerator.InitContainerImagePullPolicy, "initContainer imagePullPolicy")
	injectCmd.Flags().StringVarP(&patchGenerator.InitContainerImage, "image", "i", patchGenerator.InitContainerImage, "initContainer bootstrap image")
	injectCmd.Flags().StringVar(&patchGenerator.InitContainerResources, "resources", patchGenerator.InitContainerResources, "initContainer compute resources in JSON format")
	injectCmd.Flags().StringVarP((*string)(&patchGenerator.Strategy), "strategy", "s", string(patchGenerator.Strategy), "Default injection strategy if not specified explicitly (hostPath/initContainer)")
	injectCmd.Flags().StringVar(&patchGenerator.HostPathPrefix, "hostpath", patchGenerator.HostPathPrefix, "Location of TZif files on host machines")
	injectCmd.Flags().StringVarP(&patchGenerator.LocalTimePath, "mountpath", "m", patchGenerator.LocalTimePath, "Mount path for TZif file on containers")
	injectCmd.Flags().BoolVar(&patchGenerator.CronJobTimeZone, "cronJobTimeZone", patchGenerator.CronJobTimeZone, "Enable CronJob injection. Requires kubernetes >=1.24.0-beta.0 and the 'CronJobTimeZone' feature gate enabled (alpha)")
}
