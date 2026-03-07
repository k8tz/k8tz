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

package version

import (
	"fmt"
	"runtime"
)

var (
	AppVersion      = "0.18.1"
	VersionSuffix   = ""
	GitCommit       = ""
	ImageRepository = "quay.io/k8tz/k8tz"
)

func Version() string {
	version := AppVersion

	if VersionSuffix != "" {
		version += VersionSuffix
	}

	return version
}

func VersionWithMetadata() string {
	version := Version()
	if VersionSuffix != "" && GitCommit != "" {
		version += "+" + truncate(GitCommit, 14)
	}

	return version
}

func Image() string {
	return fmt.Sprintf("%s:%s", ImageRepository, Version())
}

func DisplayVersion() string {
	return fmt.Sprintf("k8tz v%s %s %s/%s", VersionWithMetadata(), runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func truncate(s string, maxLen int) string {
	if len(s) < maxLen {
		return s
	}
	return s[:maxLen]
}
