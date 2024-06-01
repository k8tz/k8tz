package inject

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTransformer_Transform(t *testing.T) {
	type fields struct {
		PatchGenerator PatchGenerator
		Inputs         []string
	}
	tests := []struct {
		name    string
		fields  fields
		golden  string
		wantErr bool
	}{
		{
			name: "patch pod with initContainer",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "America/Jamaica",
					InitContainerName:  "k8tz",
					InitContainerImage: "quay.io/k8tz/k8tz:0.0.1-beta2",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml"},
			},
			golden:  "testdata/test-pod-initContainer-1.yaml",
			wantErr: false,
		},
		{
			name: "patch pod with hostPath",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:       HostPathInjectionStrategy,
					Timezone:       "Asia/Jerusalem",
					HostPathPrefix: "/usr/share/zoneinfo",
					LocalTimePath:  "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml"},
			},
			golden:  "testdata/test-pod-hostPath-1.yaml",
			wantErr: false,
		},
		{
			name: "invalid yaml file should raise an error",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:       HostPathInjectionStrategy,
					Timezone:       "Asia/Jerusalem",
					HostPathPrefix: "/usr/share/zoneinfo",
					LocalTimePath:  "/etc/localtime",
				},
				Inputs: []string{"testdata/invalid.yaml"},
			},
			wantErr: true,
		},
		{
			name: "patch 2 valid pods from separate inputs with hostPath",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:       HostPathInjectionStrategy,
					Timezone:       "Europe/Astrakhan",
					HostPathPrefix: "/usr/share/zoneinfo",
					LocalTimePath:  "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml", "testdata/simple-pod.yaml"},
			},
			golden:  "testdata/test-pod-multiple-inputs-hostPath.yaml",
			wantErr: false,
		},
		{
			name: "patch 2 valid pods from single input with initContainer",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "Europe/Luxembourg",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/two-pods.yaml"},
			},
			golden:  "testdata/two-pods-initContainer.yaml",
			wantErr: false,
		},
		{
			name: "transform a simple StatefulSet",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/statefulset-simple.yaml"},
			},
			golden:  "testdata/statefulset-simple-injected.yaml",
			wantErr: false,
		},
		{
			name: "transform list of StatefulSets",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/statefulset-list.yaml"},
			},
			golden:  "testdata/statefulset-list-injected.yaml",
			wantErr: false,
		},
		{
			name: "mix single input with multiple pods and another pod from a different input",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml", "testdata/two-pods.yaml"},
			},
			golden:  "testdata/mixed-inputs-1.yaml",
			wantErr: false,
		},
		{
			name: "unrelated object should remain untouched",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/namespace.yaml"},
			},
			golden:  "testdata/namespace.yaml",
			wantErr: false,
		},
		{
			name: "unrelated object should remain untouched even when surrounded by pods",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml", "testdata/namespace.yaml", "testdata/simple-pod.yaml"},
			},
			golden:  "testdata/namespace-surrounded-by-pods.yaml",
			wantErr: false,
		},
		{
			name: "pod with existing initContainer that should preserve",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/pod-with-initContainers.yaml"},
			},
			golden:  "testdata/pod-with-initContainer-injected.yaml",
			wantErr: false,
		},
		{
			name: "simple cronjob injection",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "Europe/Dublin",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
					CronJobTimeZone:    true,
				},
				Inputs: []string{"testdata/simple-cronjob.yaml"},
			},
			golden:  "testdata/simple-cronjob-dublin.yaml",
			wantErr: false,
		},
		{
			name: "list of uninjected deployments",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/list-of-uninjected-deployments.yaml"},
			},
			golden:  "testdata/list-of-uninjected-deployments-injected.yaml",
			wantErr: false,
		},
		{
			name: "test container volumeMounts",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           HostPathInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/test-pod-volumeMounts.yaml"},
			},
			golden:  "testdata/test-pod-volumeMounts-result.yaml",
			wantErr: false,
		},
		{
			name: "test initContainer",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:           InitContainerInjectionStrategy,
					Timezone:           "UTC",
					InitContainerName:  "k8tz",
					InitContainerImage: "testimage:0.0.0",
					HostPathPrefix:     "/usr/share/zoneinfo",
					LocalTimePath:      "/etc/localtime",
				},
				Inputs: []string{"testdata/test-pod-volumeMounts.yaml"},
			},
			golden:  "testdata/test-pod-volumeMounts-initContainer-result.yaml",
			wantErr: false,
		},
		{
			name: "patch pod with valid full compute resources for the initContainer",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:               InitContainerInjectionStrategy,
					Timezone:               "America/Jamaica",
					InitContainerName:      "k8tz",
					InitContainerImage:     "quay.io/k8tz/k8tz:0.0.1-beta2",
					InitContainerResources: "{\"limits\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}",
					HostPathPrefix:         "/usr/share/zoneinfo",
					LocalTimePath:          "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml"},
			},
			golden:  "testdata/test-pod-initContainer-full-computeResources.yaml",
			wantErr: false,
		},
		{
			name: "patch pod with valid partial compute resources for the initContainer",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:               InitContainerInjectionStrategy,
					Timezone:               "America/Jamaica",
					InitContainerName:      "k8tz",
					InitContainerImage:     "quay.io/k8tz/k8tz:0.0.1-beta2",
					InitContainerResources: "{\"limits\":{\"memory\":\"128Mi\"},\"requests\":{\"cpu\":\"100m\",\"memory\":\"128Mi\"}}",
					HostPathPrefix:         "/usr/share/zoneinfo",
					LocalTimePath:          "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml"},
			},
			golden:  "testdata/test-pod-initContainer-partial-computeResources.yaml",
			wantErr: false,
		},
		{
			name: "patch pod with bad compute resources for the initContainer",
			fields: fields{
				PatchGenerator: PatchGenerator{
					Strategy:               InitContainerInjectionStrategy,
					Timezone:               "America/Jamaica",
					InitContainerName:      "k8tz",
					InitContainerImage:     "quay.io/k8tz/k8tz:0.0.1-beta2",
					InitContainerResources: "BAD_RESOURCES",
					HostPathPrefix:         "/usr/share/zoneinfo",
					LocalTimePath:          "/etc/localtime",
				},
				Inputs: []string{"testdata/simple-pod.yaml"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs, err := ArgumentsToInputs(tt.fields.Inputs)
			if err != nil {
				t.Errorf("failed to convert test arguments to inputs: %s", err.Error())
				return
			}

			var buffer bytes.Buffer
			tr := &Transformer{
				PatchGenerator: tt.fields.PatchGenerator,
				Inputs:         inputs,
				Output:         &buffer,
			}

			if err := tr.Transform(); (err != nil) != tt.wantErr {
				t.Errorf("Transformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.golden != "" {
				got := buffer.String()

				if err := compareGolden(got, tt.golden); err != nil {
					t.Errorf("Transformer.Transform(): %v", err)
				}
			}
		})
	}
}

func Test_parseTypeMetaSkeleton(t *testing.T) {
	type args struct {
		object interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "test valid but unknown object type",
			args: args{
				object: corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "test valid Pod type",
			args: args{
				object: metav1.TypeMeta{Kind: "Pod"},
			},
			want:    &corev1.Pod{},
			wantErr: false,
		},
		{
			name: "test valid Deployment type",
			args: args{
				object: metav1.TypeMeta{Kind: "Deployment"},
			},
			want:    &appsv1.Deployment{},
			wantErr: false,
		},
		{
			name: "test valid StatefulSet type",
			args: args{
				object: metav1.TypeMeta{Kind: "StatefulSet"},
			},
			want:    &appsv1.StatefulSet{},
			wantErr: false,
		},
		{
			name: "test valid CronJob type",
			args: args{
				object: metav1.TypeMeta{Kind: "CronJob"},
			},
			want:    &batchv1.CronJob{},
			wantErr: false,
		},
		{
			name: "test valid List type",
			args: args{
				object: metav1.TypeMeta{Kind: "List"},
			},
			want:    &corev1.List{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jb, err := json.Marshal(tt.args.object)
			if err != nil {
				t.Errorf("failed to marshal test argument to json: %v", tt.args.object)
				return
			}

			got, err := parseTypeMetaSkeleton(jb)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTypeMetaSkeleton() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTypeMetaSkeleton() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseTypeMetaSkeletonInvalidJson(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test invalid json in input",
			args: args{
				data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTypeMetaSkeleton(tt.args.data)
			if err == nil {
				t.Errorf("expected error on invalid json in input")
			}
		})
	}
}

func TestArgumentsToInputs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		want    Inputs
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ArgumentsToInputs(tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArgumentsToInputs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ArgumentsToInputs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func compareGolden(hyp string, goldenFile string) error {
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

		_, err = f.WriteString(hyp)
		if err != nil {
			return fmt.Errorf("failed to write golden file: %s, error: %v", goldenFile, err)
		}
	} else {
		refs := *golden
		if refs != hyp {
			return fmt.Errorf("actual: %v, want: %v", hyp, refs)
		}
	}

	return nil
}
