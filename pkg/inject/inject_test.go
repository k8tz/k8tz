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

package inject

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	k8tz "github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var updateGoldens = false

func Test_isObjectInjected(t *testing.T) {
	type args struct {
		obj *metav1.ObjectMeta
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "empty object",
			args: args{
				obj: &metav1.ObjectMeta{},
			},
			want: false,
		},
		{
			name: "non-related annotation",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						"not-related-annotation": "not-related-value",
					},
				},
			},
			want: false,
		},
		{
			name: "annotation value is false",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "false",
					},
				},
			},
			want: false,
		},
		{
			name: "annotation value is empty",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "",
					},
				},
			},
			want: false,
		},
		{
			name: "annotation value is not related",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "k8tz is fun",
					},
				},
			},
			want: false,
		},
		{
			name: "annotation value is true",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "true",
					},
				},
			},
			want: true,
		},
		{
			name: "annotation value is true with capital letter",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "True",
					},
				},
			},
			want: true,
		},
		{
			name: "annotation value is 1",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "1",
					},
				},
			},
			want: true,
		},
		{
			name: "annotation value is T",
			args: args{
				obj: &metav1.ObjectMeta{
					Annotations: map[string]string{
						k8tz.InjectedAnnotation: "T",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isObjectInjected(tt.args.obj); got != tt.want {
				t.Errorf("isObjectInjected() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatchGenerator_Generate(t *testing.T) {
	type fields struct {
		Strategy           InjectionStrategy
		Timezone           string
		InitContainerImage string
		HostPathPrefix     string
	}
	type args struct {
		object interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		golden  string
		wantErr bool
	}{
		{
			name: "nil object should raise exception",
			fields: fields{
				Strategy:           InitContainerInjectionStrategy,
				Timezone:           "UTC",
				InitContainerImage: "",
				HostPathPrefix:     "/",
			},
			args: args{
				object: nil,
			},
			wantErr: true,
		},
		{
			name: "unsupported object type should raise exception",
			fields: fields{
				Strategy:           InitContainerInjectionStrategy,
				Timezone:           "UTC",
				InitContainerImage: "",
				HostPathPrefix:     "/",
			},
			args: args{
				object: &corev1.Namespace{},
			},
			wantErr: true,
		},
		{
			name: "unsupported injection strategy type should raise exception",
			fields: fields{
				Strategy:           "moo",
				Timezone:           "UTC",
				InitContainerImage: "",
				HostPathPrefix:     "/",
			},
			args: args{
				object: &corev1.Pod{},
			},
			wantErr: true,
		},
		{
			name: "Pod object should not raise exception",
			fields: fields{
				Strategy:           HostPathInjectionStrategy,
				Timezone:           "UTC",
				InitContainerImage: "",
				HostPathPrefix:     "/",
			},
			args: args{
				object: &corev1.Pod{},
			},
			wantErr: false,
		},
		{
			name: "Deployment object should not raise exception",
			fields: fields{
				Strategy:           InitContainerInjectionStrategy,
				Timezone:           "UTC",
				InitContainerImage: "",
				HostPathPrefix:     "/",
			},
			args: args{
				object: &appsv1.Deployment{},
			},
			wantErr: false,
		},
		{
			name: "List object should not raise exception",
			fields: fields{
				Strategy:           InitContainerInjectionStrategy,
				Timezone:           "UTC",
				InitContainerImage: "",
				HostPathPrefix:     "/",
			},
			args: args{
				object: &corev1.List{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &PatchGenerator{
				Strategy:           tt.fields.Strategy,
				Timezone:           tt.fields.Timezone,
				InitContainerImage: tt.fields.InitContainerImage,
				HostPathPrefix:     tt.fields.HostPathPrefix,
			}
			got, err := g.Generate(tt.args.object, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchGenerator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.golden != "" {
				if err := comparePatches(&got, tt.golden); err != nil {
					t.Errorf("PatchGenerator.Generate(): %v", err)
				}
			}
		})
	}
}

// func TestLists(t *testing.T) {
// 	type fields struct {
// 		Strategy           InjectionStrategy
// 		Timezone           string
// 		InitContainerImage string
// 		HostPathPrefix     string
// 	}
// 	tests := []struct {
// 		name     string
// 		fields   fields
// 		listFile string
// 		golden   string
// 		wantErr  bool
// 	}{
// 		{
// 			name: "Simple list with one pod",
// 			fields: fields{
// 				Strategy:           InitContainerInjectionStrategy,
// 				Timezone:           "UTC",
// 				InitContainerImage: "",
// 				HostPathPrefix:     "/",
// 			},
// 			listFile: "testdata/list-of-uninjected-deployments.yaml",
// 			golden:   "testdata/list-of-deployments-patch.json",
// 			wantErr:  false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			g := &PatchGenerator{
// 				Strategy:           tt.fields.Strategy,
// 				Timezone:           tt.fields.Timezone,
// 				InitContainerImage: tt.fields.InitContainerImage,
// 				HostPathPrefix:     tt.fields.HostPathPrefix,
// 			}

// 			list := corev1.List{}
// 			bytes, err := ioutil.ReadFile(tt.listFile)
// 			if err != nil {
// 				t.Fatalf("TestLists failed to read list file: %s, error: %v", tt.listFile, err)
// 				return
// 			}

// 			err = yaml.Unmarshal(bytes, &list)
// 			if err != nil {
// 				t.Fatalf("TestLists failed to unmarshal list file: %s, error: %v", tt.listFile, err)
// 				return
// 			}

// 			got, err := g.Generate(&list, "")
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("TestLists error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}

// 			if tt.golden != "" {
// 				if err := comparePatches(t, &got, tt.golden); err != nil {
// 					t.Errorf("TestLists: %v", err)
// 				}
// 			}
// 		})
// 	}
// }

func TestPatchGenerator_createEnvironmentVariablePatches(t *testing.T) {
	type fields struct {
		Strategy           InjectionStrategy
		Timezone           string
		InitContainerImage string
		HostPathPrefix     string
	}
	type args struct {
		meta       *metav1.ObjectMeta
		spec       *corev1.PodSpec
		pathprefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		golden string
	}{
		{
			name: "test TZ environment variable with initContainer for 2 containers pod",
			fields: fields{
				Strategy: InitContainerInjectionStrategy,
				Timezone: "Asia/Chita",
			},
			args: args{
				meta: &metav1.ObjectMeta{Name: "myPod"},
				spec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "firstContainer",
							Image: "nginx:latest",
						},
						{
							Name:  "secondContainer",
							Image: "nginx:latest",
						},
					},
				},
				pathprefix: "/spec",
			},
			golden: "testdata/env-2-containers-pod.yaml",
		},
		{
			name: "test TZ environment variable with hostPath for 2 containers pod",
			fields: fields{
				Strategy: HostPathInjectionStrategy,
				Timezone: "Canada",
			},
			args: args{
				meta: &metav1.ObjectMeta{Name: "myPod"},
				spec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "firstContainer",
							Image: "nginx:latest",
						},
						{
							Name:  "secondContainer",
							Image: "nginx:latest",
						},
					},
				},
				pathprefix: "/spec",
			},
			golden: "testdata/env-2-containers-hostPath-pod.yaml",
		},
		{
			name: "test TZ environment variable without containers should return empty array",
			fields: fields{
				Strategy: InitContainerInjectionStrategy,
				Timezone: "Europe/Oslo",
			},
			args: args{
				meta: &metav1.ObjectMeta{Name: "myEmptyPod"},
				spec: &corev1.PodSpec{
					Containers: []corev1.Container{},
				},
				pathprefix: "/spec",
			},
			golden: "testdata/env-without-containers.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &PatchGenerator{
				Strategy:           tt.fields.Strategy,
				Timezone:           tt.fields.Timezone,
				InitContainerImage: version.Image(),
				HostPathPrefix:     "/usr/share/zoneinfo",
			}

			got := g.createEnvironmentVariablePatches(tt.args.spec, tt.args.pathprefix)
			if err := comparePatches(&got, tt.golden); err != nil {
				t.Errorf("PatchGenerator.createEnvironmentVariablePatches(): %v", err)
			}
		})
	}
}

func TestPatchGenerator_createInitContainerPatches(t *testing.T) {
	type fields struct {
		Strategy           InjectionStrategy
		Timezone           string
		InitContainerImage string
		HostPathPrefix     string
	}
	type args struct {
		metadata   *metav1.ObjectMeta
		spec       *corev1.PodSpec
		pathprefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		golden string
	}{
		{
			name: "test initContainer patch without containers should return empty array",
			fields: fields{
				Strategy:           InitContainerInjectionStrategy,
				Timezone:           "Pacific/Fiji",
				InitContainerImage: version.Image(),
			},
			args: args{
				metadata:   &metav1.ObjectMeta{Name: "myPod"},
				spec:       &corev1.PodSpec{},
				pathprefix: "/spec",
			},
			golden: "testdata/initcontainerstrategy-without-containers.json",
		},
		{
			name: "test initContainer patch with for two containers",
			fields: fields{
				Strategy:           InitContainerInjectionStrategy,
				Timezone:           "America/Panama",
				InitContainerImage: "custom.registry.local:5000/repository/k8tz:1.0.0-beta1",
			},
			args: args{
				metadata: &metav1.ObjectMeta{Name: "myPod"},
				spec: &corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "container1",
							Image: "container:1",
						},
						{
							Name:  "container2",
							Image: "container:2",
						},
					},
				},
				pathprefix: "/spec",
			},
			golden: "testdata/initcontainerstrategy-2-containers.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &PatchGenerator{
				Strategy:           tt.fields.Strategy,
				Timezone:           tt.fields.Timezone,
				InitContainerImage: tt.fields.InitContainerImage,
				HostPathPrefix:     "/usr/share/zoneinfo",
				LocalTimePath:      "/etc/localtime",
			}

			got := g.createInitContainerPatches(tt.args.spec, tt.args.pathprefix)
			if err := comparePatches(&got, tt.golden); err != nil {
				t.Errorf("PatchGenerator.createInitContainerPatches(): %v", err)
			}
		})
	}
}

func TestPatchGenerator_createHostPathPatches(t *testing.T) {
	type fields struct {
		Strategy           InjectionStrategy
		Timezone           string
		InitContainerImage string
		HostPathPrefix     string
	}
	type args struct {
		metadata   *metav1.ObjectMeta
		spec       *corev1.PodSpec
		pathprefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		golden string
	}{
		{
			name: "test hostPath patches without containers should return empty array",
			fields: fields{
				Strategy:       HostPathInjectionStrategy,
				Timezone:       "Europe/Copenhagen",
				HostPathPrefix: "/usr/share/zoneinfo",
			},
			args: args{
				metadata:   &metav1.ObjectMeta{Name: "myPod"},
				spec:       &corev1.PodSpec{},
				pathprefix: "/spec",
			},
			golden: "testdata/hostpatchstrategy-without-containers.json",
		},
		{
			name: "test hostPath patches with 2 containers",
			fields: fields{
				Strategy:       HostPathInjectionStrategy,
				Timezone:       "Europe/Vatican",
				HostPathPrefix: "/usr/share/zoneinfo",
			},
			args: args{
				metadata: &metav1.ObjectMeta{Name: "myPod"},
				spec: &corev1.PodSpec{Containers: []corev1.Container{
					{
						Name:  "container1",
						Image: "container:1",
					},
					{
						Name:  "container2",
						Image: "container:2",
					},
				}},
				pathprefix: "/spec",
			},
			golden: "testdata/hostpatchstrategy-2-containers.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &PatchGenerator{
				Strategy:       tt.fields.Strategy,
				Timezone:       tt.fields.Timezone,
				HostPathPrefix: tt.fields.HostPathPrefix,
				LocalTimePath:  "/etc/localtime",
			}
			got := g.createHostPathPatches(tt.args.spec, tt.args.pathprefix)
			if err := comparePatches(&got, tt.golden); err != nil {
				t.Errorf("PatchGenerator.createHostPathPatches(): %v", err)
			}
		})
	}
}

func TestPatchGenerator_createPostInjectionAnnotations(t *testing.T) {
	type fields struct {
		Strategy           InjectionStrategy
		Timezone           string
		InitContainerImage string
		HostPathPrefix     string
	}
	type args struct {
		meta       *metav1.ObjectMeta
		pathprefix string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		golden string
	}{
		{
			name: "test post injection annotations",
			fields: fields{
				Strategy:       InitContainerInjectionStrategy,
				Timezone:       "America/Anguilla",
				HostPathPrefix: "/usr/share/zoneinfo",
			},
			args: args{
				pathprefix: "/spec",
				meta:       &metav1.ObjectMeta{Name: "k8tz"},
			},
			golden: "testdata/postinjectionannotations-patch.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &PatchGenerator{
				Strategy:           tt.fields.Strategy,
				Timezone:           tt.fields.Timezone,
				InitContainerImage: tt.fields.InitContainerImage,
				HostPathPrefix:     tt.fields.HostPathPrefix,
			}

			got := g.createPostInjectionAnnotations(tt.args.meta, tt.args.pathprefix)
			if err := comparePatches(&got, tt.golden); err != nil {
				t.Errorf("TestPatchGenerator_createPostInjectionAnnotations: %v", err)
			}
		})
	}
}

func comparePatches(got *k8tz.Patches, goldenFile string) error {
	hyp, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert reference to json, reference: %v, err: %v", got, err)
	}

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

		_, err = f.Write(hyp)
		if err != nil {
			return fmt.Errorf("failed to write golden file: %s, error: %v", goldenFile, err)
		}
	} else {
		hyps := string(hyp)
		refs := *golden
		if refs != hyps {
			return fmt.Errorf("actual: %v, want: %v", hyps, refs)
		}
	}

	return nil
}

func readGolden(file string) (*string, bool, error) {
	if _, err := os.Stat(file); err == nil {
		bytes, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, true, err
		}

		content := string(bytes)
		return &content, true, nil
	} else if os.IsNotExist(err) {
		return nil, false, nil
	} else {
		return nil, false, err
	}
}

func Test_escapeJsonPointer(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty pointer",
			args: args{
				p: "",
			},
			want: "",
		},
		{
			name: "escape slash",
			args: args{
				p: "k8tz.io/timezone",
			},
			want: "k8tz.io~1timezone",
		},
		{
			name: "escape tilde",
			args: args{
				p: "~test~",
			},
			want: "~0test~0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeJsonPointer(tt.args.p); got != tt.want {
				t.Errorf("escapeJsonPointer() = %v, want %v", got, tt.want)
			}
		})
	}
}
