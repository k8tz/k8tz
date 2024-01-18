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
	"fmt"
	"github.com/k8tz/k8tz/pkg/bootstrap"
	"github.com/spf13/cobra"
)

var operation = bootstrap.NewBootstrapOperation()

var bootstrapCmd = &cobra.Command{
	Use:    fmt.Sprintf("bootstrap [--from=%s] [--to=%s] [--overwrite=%t]", operation.From, operation.To, operation.Overwrite),
	Hidden: true,
	Example: `k8tz bootstrap --from=/zoneinfo
k8tz bootstrap -t/path/to/target`,
	Short: "Bootstraps zoneinfo directory from a source directory",
	Long: `Bootstraps zoneinfo directory with TZif files from a source directory.

This command is used inside a bootstrap initContainer to extract
files stored in the k8tz image to emptyDir so other containers
will be able to mount the required TZif file from it.`,
	Run: func(cmd *cobra.Command, args []string) {
		cobra.CheckErr(operation.Bootstrap())
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().StringVarP(&operation.From, "from", "f", operation.From, "Path to directory where to take the files from")
	bootstrapCmd.Flags().StringVarP(&operation.To, "to", "t", operation.To, "Path to directory where copy the files to")
	bootstrapCmd.Flags().BoolVarP(&operation.Overwrite, "overwrite", "o", operation.Overwrite, "If true and file already exists in target directory, it will be overwritten. If false it will be skipped.")
	bootstrapCmd.Flags().BoolVarP(&operation.Verbose, "verbose", "v", operation.Verbose, "Print more verbose logs for debugging")
}
