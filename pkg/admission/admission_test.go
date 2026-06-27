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
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	k8tz "github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/inject"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
)

var updateGoldens = false

func TestAdmissionRequestsHandler_handleFunc(t *testing.T) {
	type fields struct {
		DefaultTimezone          string
		ContainerName            string
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
		CronJobTimeZone          bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "simple valid request",
			fields: fields{
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
			name: "cronjob request should be handled when feature enabled",
			fields: fields{
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-cronjob.json",
				GoldenFile:               "testdata/review-cronjob-enabled.json",
				WantCode:                 http.StatusOK,
				CronJobTimeZone:          true,
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
			},
		},
		{
			name: "cronjob request should be ignored when feature enabled but InjectByDefault is false and not overridden by annotation",
			fields: fields{
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          false,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				ContentType:              "application/json",
				Method:                   "POST",
				ReviewFile:               "testdata/review-cronjob.json",
				GoldenFile:               "testdata/review-cronjob-ignored.json",
				WantCode:                 http.StatusOK,
				CronJobTimeZone:          true,
				FakeObjects: []runtime.Object{
					&corev1.Namespace{
						ObjectMeta: v1.ObjectMeta{
							Name: "default",
						},
					},
				},
			},
		},
		{
			name: "unparsable review should be considered bad request",
			fields: fields{
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
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
			// long warnings are expected here and better not be logged
			k8tz.WarningLogger.SetOutput(io.Discard)

			h := &RequestsHandler{
				DefaultTimezone:          tt.fields.DefaultTimezone,
				ContainerName:            tt.fields.ContainerName,
				BootstrapImage:           tt.fields.BootstrapImage,
				DefaultInjectionStrategy: tt.fields.DefaultInjectionStrategy,
				InjectByDefault:          tt.fields.InjectByDefault,
				HostPathPrefix:           tt.fields.HostPathPrefix,
				LocalTimePath:            tt.fields.LocalTimePath,
				CronJobTimeZone:          tt.fields.CronJobTimeZone,
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

func compareReviews(got *bytes.Buffer, goldenFile string) (err error) {
	golden, exists, err := readGolden(goldenFile)
	if err != nil {
		return fmt.Errorf("golden file: %s, exists: %t, err: %v", goldenFile, exists, err)
	}

	if !exists || updateGoldens {
		f, err := os.Create(goldenFile)
		if err != nil {
			return fmt.Errorf("failed to create golden file: %s, error: %v", goldenFile, err)
		}

		defer func() {
			if closeErr := f.Close(); err == nil && closeErr != nil {
				err = fmt.Errorf("failed to close golden file: %s, error: %v", goldenFile, closeErr)
			}
		}()

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
		data, err := os.ReadFile(file)
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

func TestRequestsHandler_lookupPodParentAnnotations(t *testing.T) {
	tests := []struct {
		name            string
		pod             *corev1.Pod
		objects         []runtime.Object
		reactor         func(*fake.Clientset)
		injectByDefault *bool
		wantNil         bool
		wantTimezone    string
		wantStrategy    inject.InjectionStrategy
	}{
		{
			name: "pod annotation wins over all parents",
			pod: testPod(map[string]string{
				k8tz.TimezoneAnnotation:          "Europe/Pod",
				k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
			}, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation:          "Europe/Namespace",
					k8tz.InjectionStrategyAnnotation: string(inject.InitContainerInjectionStrategy),
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.TimezoneAnnotation:          "Europe/ReplicaSet",
					k8tz.InjectionStrategyAnnotation: string(inject.InitContainerInjectionStrategy),
				}, testOwnerReference("apps/v1", "Deployment", "deploy")),
				testDeployment("deploy", map[string]string{
					k8tz.TimezoneAnnotation: "Europe/Deployment",
				}),
			},
			wantTimezone: "Europe/Pod",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "replicaset annotation wins over deployment and namespace",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation:          "Europe/Namespace",
					k8tz.InjectionStrategyAnnotation: string(inject.InitContainerInjectionStrategy),
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.TimezoneAnnotation:          "Europe/ReplicaSet",
					k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
				}, testOwnerReference("apps/v1", "Deployment", "deploy")),
				testDeployment("deploy", map[string]string{
					k8tz.TimezoneAnnotation:          "Europe/Deployment",
					k8tz.InjectionStrategyAnnotation: string(inject.InitContainerInjectionStrategy),
				}),
			},
			wantTimezone: "Europe/ReplicaSet",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "deployment annotation is used when replicaset lacks the key",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Europe/Namespace",
				}),
				testReplicaSet("rs", nil, testOwnerReference("apps/v1", "Deployment", "deploy")),
				testDeployment("deploy", map[string]string{
					k8tz.TimezoneAnnotation: "Europe/Deployment",
				}),
			},
			wantTimezone: "Europe/Deployment",
			wantStrategy: inject.InitContainerInjectionStrategy,
		},
		{
			name: "job and cronjob annotations are inherited",
			pod:  testPod(nil, testOwnerReference("batch/v1", "Job", "job")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Europe/Namespace",
				}),
				testJob("job", map[string]string{
					k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
				}, testOwnerReference("batch/v1", "CronJob", "cron")),
				testCronJob("cron", map[string]string{
					k8tz.TimezoneAnnotation: "Pacific/Honolulu",
				}),
			},
			wantTimezone: "Pacific/Honolulu",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "statefulset annotation is inherited",
			pod:  testPod(nil, testOwnerReference("apps/v1", "StatefulSet", "sts")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Europe/Namespace",
				}),
				testStatefulSet("sts", map[string]string{
					k8tz.TimezoneAnnotation: "Asia/Tokyo",
				}),
			},
			wantTimezone: "Asia/Tokyo",
			wantStrategy: inject.InitContainerInjectionStrategy,
		},
		{
			name: "daemonset annotation is inherited",
			pod:  testPod(nil, testOwnerReference("apps/v1", "DaemonSet", "ds")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Europe/Namespace",
				}),
				testDaemonSet("ds", map[string]string{
					k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
				}),
			},
			wantTimezone: "Europe/Namespace",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "annotation keys merge independently by closest source",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Asia/Jerusalem",
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.InjectAnnotation: "true",
				}, testOwnerReference("apps/v1", "Deployment", "deploy")),
				testDeployment("deploy", map[string]string{
					k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
				}),
			},
			wantTimezone: "Asia/Jerusalem",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "parent inject false skips injection",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.InjectAnnotation: "true",
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.InjectAnnotation: "false",
				}),
			},
			wantNil: true,
		},
		{
			name: "parent inject true enables injection when default is disabled",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(nil),
				testReplicaSet("rs", map[string]string{
					k8tz.InjectAnnotation: "true",
				}),
			},
			injectByDefault: testBool(false),
			wantTimezone:    k8tz.UTCTimezone,
			wantStrategy:    inject.InitContainerInjectionStrategy,
		},
		{
			name: "missing owner lookup falls back to namespace",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "missing-rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Asia/Jerusalem",
				}),
			},
			wantTimezone: "Asia/Jerusalem",
			wantStrategy: inject.InitContainerInjectionStrategy,
		},
		{
			name: "forbidden owner lookup falls back to namespace",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Asia/Jerusalem",
				}),
			},
			reactor: func(clientset *fake.Clientset) {
				clientset.PrependReactor("get", "replicasets", func(action ktesting.Action) (bool, runtime.Object, error) {
					return true, nil, apierrors.NewForbidden(schema.GroupResource{Group: "apps", Resource: "replicasets"}, "rs", fmt.Errorf("denied"))
				})
			},
			wantTimezone: "Asia/Jerusalem",
			wantStrategy: inject.InitContainerInjectionStrategy,
		},
		{
			name: "unsupported custom owner falls back to namespace",
			pod:  testPod(nil, testOwnerReference("example.com/v1", "Widget", "widget")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Asia/Jerusalem",
				}),
			},
			wantTimezone: "Asia/Jerusalem",
			wantStrategy: inject.InitContainerInjectionStrategy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8tz.WarningLogger.SetOutput(io.Discard)

			clientset := fake.NewSimpleClientset(tt.objects...)
			if tt.reactor != nil {
				tt.reactor(clientset)
			}

			injectByDefault := true
			if tt.injectByDefault != nil {
				injectByDefault = *tt.injectByDefault
			}

			h := &RequestsHandler{
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          injectByDefault,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				PodOwnerLookup:           true,
				clientset:                clientset,
			}

			got, err := h.lookupPod("default", tt.pod)
			if err != nil {
				t.Fatalf("lookupPod() error = %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Fatalf("lookupPod() = %+v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("lookupPod() = nil, want generator")
			}
			if got.Timezone != tt.wantTimezone {
				t.Errorf("lookupPod().Timezone = %s, want %s", got.Timezone, tt.wantTimezone)
			}
			if got.Strategy != tt.wantStrategy {
				t.Errorf("lookupPod().Strategy = %s, want %s", got.Strategy, tt.wantStrategy)
			}
		})
	}
}

func TestRequestsHandler_lookupPodParentAnnotationsDisabled(t *testing.T) {
	tests := []struct {
		name         string
		pod          *corev1.Pod
		objects      []runtime.Object
		wantTimezone string
		wantStrategy inject.InjectionStrategy
	}{
		{
			name: "owner annotations are ignored",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation:          "Asia/Jerusalem",
					k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.TimezoneAnnotation:          "Europe/ReplicaSet",
					k8tz.InjectionStrategyAnnotation: string(inject.InitContainerInjectionStrategy),
				}),
			},
			wantTimezone: "Asia/Jerusalem",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "pod annotations still beat namespace annotations",
			pod: testPod(map[string]string{
				k8tz.TimezoneAnnotation:          "Europe/Pod",
				k8tz.InjectionStrategyAnnotation: string(inject.HostPathInjectionStrategy),
			}, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation:          "Asia/Jerusalem",
					k8tz.InjectionStrategyAnnotation: string(inject.InitContainerInjectionStrategy),
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.TimezoneAnnotation: "Europe/ReplicaSet",
				}),
			},
			wantTimezone: "Europe/Pod",
			wantStrategy: inject.HostPathInjectionStrategy,
		},
		{
			name: "namespace annotations are used without pod annotations",
			pod:  testPod(nil, testOwnerReference("apps/v1", "ReplicaSet", "rs")),
			objects: []runtime.Object{
				testNamespace(map[string]string{
					k8tz.TimezoneAnnotation: "Asia/Jerusalem",
				}),
				testReplicaSet("rs", map[string]string{
					k8tz.TimezoneAnnotation: "Europe/ReplicaSet",
				}),
			},
			wantTimezone: "Asia/Jerusalem",
			wantStrategy: inject.InitContainerInjectionStrategy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8tz.WarningLogger.SetOutput(io.Discard)

			ownerLookupCalled := false
			clientset := fake.NewSimpleClientset(tt.objects...)
			clientset.PrependReactor("get", "replicasets", func(action ktesting.Action) (bool, runtime.Object, error) {
				ownerLookupCalled = true
				return true, nil, fmt.Errorf("owner lookup should not be called when pod owner lookup is disabled")
			})

			h := &RequestsHandler{
				DefaultTimezone:          k8tz.UTCTimezone,
				ContainerName:            "k8tz",
				BootstrapImage:           "test:0.0.0",
				DefaultInjectionStrategy: inject.InitContainerInjectionStrategy,
				InjectByDefault:          true,
				HostPathPrefix:           "/usr/share/zoneinfo",
				LocalTimePath:            "/etc/localtime",
				clientset:                clientset,
			}

			got, err := h.lookupPod("default", tt.pod)
			if err != nil {
				t.Fatalf("lookupPod() error = %v", err)
			}
			if ownerLookupCalled {
				t.Fatal("owner lookup was called when pod owner lookup is disabled")
			}
			if got == nil {
				t.Fatal("lookupPod() = nil, want generator")
			}
			if got.Timezone != tt.wantTimezone {
				t.Errorf("lookupPod().Timezone = %s, want %s", got.Timezone, tt.wantTimezone)
			}
			if got.Strategy != tt.wantStrategy {
				t.Errorf("lookupPod().Strategy = %s, want %s", got.Strategy, tt.wantStrategy)
			}
		})
	}
}

func testNamespace(annotations map[string]string) *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name:        "default",
			Annotations: annotations,
		},
	}
}

func testPod(annotations map[string]string, ownerReferences ...v1.OwnerReference) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:            "pod",
			Namespace:       "default",
			Annotations:     annotations,
			OwnerReferences: ownerReferences,
		},
	}
}

func testReplicaSet(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: testObjectMeta(name, annotations, ownerReferences...),
	}
}

func testDeployment(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: testObjectMeta(name, annotations, ownerReferences...),
	}
}

func testStatefulSet(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: testObjectMeta(name, annotations, ownerReferences...),
	}
}

func testDaemonSet(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: testObjectMeta(name, annotations, ownerReferences...),
	}
}

func testJob(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: testObjectMeta(name, annotations, ownerReferences...),
	}
}

func testCronJob(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) *batchv1.CronJob {
	return &batchv1.CronJob{
		ObjectMeta: testObjectMeta(name, annotations, ownerReferences...),
	}
}

func testObjectMeta(name string, annotations map[string]string, ownerReferences ...v1.OwnerReference) v1.ObjectMeta {
	return v1.ObjectMeta{
		Name:            name,
		Namespace:       "default",
		Annotations:     annotations,
		OwnerReferences: ownerReferences,
	}
}

func testOwnerReference(apiVersion, kind, name string) v1.OwnerReference {
	return v1.OwnerReference{
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
		Controller: &inject.True,
	}
}

func testBool(v bool) *bool {
	return &v
}
