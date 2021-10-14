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

package admission

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/inject"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

var updateGoldens = false

func TestAdmissionRequestsHandler_handleFunc(t *testing.T) {
	type fields struct {
		DefaultTimezone          string
		BootstrapImage           string
		DefaultInjectionStrategy inject.InjectionStrategy
		InjectByDefault          bool
		HostPathPrefix           string
		LocalTimePath            string
		ContentType              string
		Method                   string
		ReviewFile               string
		GoldenFile               string
		FakeObjects              []runtime.Object
		WantCode                 int
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "simple valid request",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/review-pod-golden.json",
				FakeObjects:              []runtime.Object{&corev1.Namespace{ObjectMeta: v1.ObjectMeta{Name: "default"}}},
				WantCode:                 http.StatusOK,
			},
		},
		{
			name: "request with wrong method: get",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "GET",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "",
				WantCode:                 http.StatusMethodNotAllowed,
			},
		},
		{
			name: "request with wrong content type: text/plain",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "text/plain",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "",
				WantCode:                 http.StatusBadRequest,
			},
		},
		{
			name: "unsupported object as namespace should be ignored",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-namespace.json",
				GoldenFile:               "testdata/review-namespace-ignored.json",
				WantCode:                 http.StatusOK,
			},
		},
		{
			name: "unparsable review should be considered bad request",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/unparsable.json",
				GoldenFile:               "",
				WantCode:                 http.StatusBadRequest,
			},
		},
		{
			name: "review without request",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-without-request.json",
				GoldenFile:               "",
				WantCode:                 http.StatusBadRequest,
			},
		},
		{
			name: "request with invalid namespace should raise an error",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/invalid-namespace-rejection.json",
				WantCode:                 http.StatusOK,
			},
		},
		{
			name: "valid request will skip injection because of namespace annotation",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/review-pod-skipped-namespace-annotation.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
							Annotations: map[string]string{
								"k8tz.io/inject": "false",
							},
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "valid request will skip injection because of pod annotation",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-injected-pod.json",
				GoldenFile:               "testdata/review-pod-skipped-pod-annotation.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "valid request will skip injection when default is false and no annotations found",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          false,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/review-pod-skipped-default.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "valid request will be injected when default is false and pod annotation found",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          false,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-explicit-true-pod.json",
				GoldenFile:               "testdata/review-explicit-true-pod-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "valid request will be injected when default is false and namespace annotation found",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          false,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/review-explicit-true-namespace-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
							Annotations: map[string]string{
								"k8tz.io/inject": "true",
							},
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "valid request will be injected when default is true but pod annotation is explicitly false",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-explicit-false-pod.json",
				GoldenFile:               "testdata/review-explicit-false-pod-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "explicit timezone annotation on pod",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-explicit-timezone-pod.json",
				GoldenFile:               "testdata/review-explicit-timezone-pod-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "explicit timezone annotation on namespace",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/review-pod-timezone-namespace-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
							Annotations: map[string]string{
								"k8tz.io/timezone": "Israel",
							},
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "explicit strategy annotation on pod",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-explicit-strategy-pod.json",
				GoldenFile:               "testdata/review-explicit-strategy-pod-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "explicit strategy annotation on namespace",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-pod.json",
				GoldenFile:               "testdata/review-pod-strategy-namespace-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
							Annotations: map[string]string{
								"k8tz.io/strategy": "hostPath",
							},
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
		{
			name: "explicit strategy annotation on namespace",
			fields: fields{
				DefaultTimezone:          pkg.UTCTimezone,
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-unparsable-pod.json",
				GoldenFile:               "testdata/review-unparsable-pod-response.json",
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
				WantCode: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &RequestsHandler{
				DefaultTimezone:          tt.fields.DefaultTimezone,
				BootstrapImage:           tt.fields.BootstrapImage,
				DefaultInjectionStrategy: tt.fields.DefaultInjectionStrategy,
				InjectByDefault:          tt.fields.InjectByDefault,
				HostPathPrefix:           tt.fields.HostPathPrefix,
				LocalTimePath:            tt.fields.LocalTimePath,
				clientset:                fake.NewSimpleClientset(tt.fields.FakeObjects...),
			}

			inputFile, err := os.Open(tt.fields.ReviewFile)
			if err != nil {
				t.Fatal(err)
			}

			req, err := http.NewRequest(tt.fields.Method, "/", inputFile)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("Content-Type", tt.fields.ContentType)

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(h.handleFunc)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.fields.WantCode {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.fields.WantCode)
				return
			}

			if tt.fields.GoldenFile != "" {
				if err := compareReviews(rr.Body, tt.fields.GoldenFile); err != nil {
					t.Errorf("TestAdmissionRequestsHandler_handleFunc: %v", err)
				}
			}
		})
	}
}

func compareReviews(got *bytes.Buffer, goldenFile string) error {
	golden, exists, err := readGolden(goldenFile)
	if err != nil {
		return fmt.Errorf("golden file: %s, exists: %t, err: %v", goldenFile, exists, err)
	}

	if !exists || updateGoldens {
		f, err := os.Create(goldenFile)
		if err != nil {
			return fmt.Errorf("failed to create golden file: %s, error: %v", goldenFile, err)
		}

		defer f.Close()

		_, err = f.Write(got.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write golden file: %s, error: %v", goldenFile, err)
		}
	} else {
		hyps := got.String()
		refs := *golden
		if refs != hyps {
			return fmt.Errorf("actual: %v, want: %v", hyps, refs)
		}
	}

	return nil
}

func readGolden(file string) (*string, bool, error) {
	if _, err := os.Stat(file); err == nil {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, true, err
		}

		content := string(data)
		return &content, true, nil
	} else if os.IsNotExist(err) {
		return nil, false, nil
	} else {
		return nil, false, err
	}
}
