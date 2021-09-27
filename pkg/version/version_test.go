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

import "testing"

func TestVersion(t *testing.T) {
	tests := []struct {
		name          string
		appVersion    string
		versionSuffix string
		gitCommit     string
		want          string
	}{
		{
			name:          "development version",
			appVersion:    "0.0.0",
			versionSuffix: "-beta1",
			gitCommit:     "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
			want:          "0.0.0-beta1",
		},
		{
			name:          "release version",
			appVersion:    "0.0.0",
			versionSuffix: "",
			gitCommit:     "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
			want:          "0.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppVersion = tt.appVersion
			VersionSuffix = tt.versionSuffix
			GitCommit = tt.gitCommit

			if got := Version(); got != tt.want {
				t.Errorf("Version() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionWithMetadata(t *testing.T) {
	tests := []struct {
		name          string
		appVersion    string
		versionSuffix string
		gitCommit     string
		want          string
	}{
		{
			name:          "development version without git commit",
			appVersion:    "0.6.0",
			versionSuffix: "-beta1",
			gitCommit:     "",
			want:          "0.6.0-beta1",
		},
		{
			name:          "development version with short git commit",
			appVersion:    "0.6.0",
			versionSuffix: "-beta1",
			gitCommit:     "c9b6",
			want:          "0.6.0-beta1+c9b6",
		},
		{
			name:          "development version with git commit to truncate",
			appVersion:    "0.6.0",
			versionSuffix: "-beta1",
			gitCommit:     "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
			want:          "0.6.0-beta1+c9b6d46ae04fec",
		},
		{
			name:          "release version",
			appVersion:    "1.0.0",
			versionSuffix: "",
			gitCommit:     "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
			want:          "1.0.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppVersion = tt.appVersion
			VersionSuffix = tt.versionSuffix
			GitCommit = tt.gitCommit

			if got := VersionWithMetadata(); got != tt.want {
				t.Errorf("VersionWithMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImage(t *testing.T) {
	tests := []struct {
		name            string
		appVersion      string
		versionSuffix   string
		gitCommit       string
		imageRepository string
		want            string
	}{
		{
			name:            "development version",
			appVersion:      "0.0.0",
			versionSuffix:   "-beta1",
			gitCommit:       "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
			imageRepository: "localhost:5000/k8tz",
			want:            "localhost:5000/k8tz:0.0.0-beta1",
		},
		{
			name:            "release version",
			appVersion:      "1.1.0",
			versionSuffix:   "",
			gitCommit:       "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
			imageRepository: "cr.example.com:1000/example/k8tz",
			want:            "cr.example.com:1000/example/k8tz:1.1.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AppVersion = tt.appVersion
			VersionSuffix = tt.versionSuffix
			GitCommit = tt.gitCommit
			ImageRepository = tt.imageRepository

			if got := Image(); got != tt.want {
				t.Errorf("Image() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_truncate(t *testing.T) {
	type args struct {
		text   string
		maxLen int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{
				text:   "",
				maxLen: 42,
			},
			want: "",
		},
		{
			name: "short string",
			args: args{
				text:   "k8tz",
				maxLen: 42,
			},
			want: "k8tz",
		},
		{
			name: "long string first character",
			args: args{
				text:   "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
				maxLen: 1,
			},
			want: "c",
		},
		{
			name: "long string first 10 characters",
			args: args{
				text:   "c9b6d46ae04fecaca6e02b1f7bdef2cb3b3dac4371a692b63e64213b41ce9263",
				maxLen: 10,
			},
			want: "c9b6d46ae0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncate(tt.args.text, tt.args.maxLen); got != tt.want {
				t.Errorf("truncate() = %v, want %v", got, tt.want)
			}
		})
	}
}
