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

package pkg

import (
	"io"
	"log"
	"os"
)

const (
	// DefaultTimezone represents the default timezone for k8tz applications
	DefaultTimezone = UTCTimezone
	// UTCTimezone is TZ database name for UTC timezone
	UTCTimezone = "UTC"

	// InjectedAnnotation is a meta object annotation that indicates whether
	// object is already have k8tz timezone injected or not (output only)
	InjectedAnnotation = "k8tz.io/injected"
	// TimezoneAnnotation TODO
	TimezoneAnnotation = "k8tz.io/timezone"
	// InjectionStrategyAnnotation TODO
	InjectionStrategyAnnotation = "k8tz.io/strategy"
	// InjectAnnotation TODO
	InjectAnnotation = "k8tz.io/inject"
)

var VerboseLogger = log.New(io.Discard, "VERBOSE: ", log.Ldate|log.Ltime|log.Lshortfile)
var InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var WarningLogger = log.New(os.Stderr, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

type Patches []Patch
type Patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}
