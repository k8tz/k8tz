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
	"fmt"
	"strconv"
	"strings"

	k8tz "github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/version"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// InjectionStrategy is a method of timezone injection, since there is more
// than one way how it can be done
type InjectionStrategy string

const (
	DefaultHostPathPrefix string = "/usr/share/zoneinfo"
	DefaultLocalTimePath  string = "/etc/localtime"

	// DefaultInjectionStrategy is the default injection strategy of k8tz
	DefaultInjectionStrategy = InitContainerInjectionStrategy
	// InitContainerInjectionStrategy is an injection strategy where we inject
	// k8tz bootstrap initContainer into a pod with a shared emptyDir volume;
	// the bootstrap container then copies TZif files to the emptyDir so the
	// actual container can consume them later
	InitContainerInjectionStrategy InjectionStrategy = "initContainer"
	// HostPathInjectionStrategy is an injection strategy where we assume that
	// TZif files exists on the node machines, and we can just mount them
	// with hostPath volumes
	HostPathInjectionStrategy InjectionStrategy = "hostPath"
)

var (
	jsonPointerEscapeReplacer = strings.NewReplacer("~", "~0", "/", "~1")
)

type PatchGenerator struct {
	Strategy           InjectionStrategy
	Timezone           string
	InitContainerImage string
	HostPathPrefix     string
	LocalTimePath      string
}

func NewPatchGenerator() PatchGenerator {
	return PatchGenerator{
		Strategy:           DefaultInjectionStrategy,
		Timezone:           k8tz.DefaultTimezone,
		InitContainerImage: version.Image(),
		HostPathPrefix:     DefaultHostPathPrefix,
		LocalTimePath:      DefaultLocalTimePath,
	}
}

func isObjectInjected(obj *metav1.ObjectMeta) bool {
	v, e := obj.Annotations[k8tz.InjectedAnnotation]
	if !e {
		return false
	}

	b, _ := strconv.ParseBool(v)
	return b
}

func (g *PatchGenerator) Generate(object interface{}, pathprefix string) (patches k8tz.Patches, err error) {
	switch o := object.(type) {
	case *appsv1.Deployment:
		return g.generate(&o.Spec.Template.Spec, fmt.Sprintf("%s/spec/template/spec", pathprefix), map[string]*metav1.ObjectMeta{
			fmt.Sprintf("%s/metadata", pathprefix):               &o.ObjectMeta,
			fmt.Sprintf("%s/spec/template/metadata", pathprefix): &o.Spec.Template.ObjectMeta,
		})
	case *corev1.Pod:
		return g.generate(&o.Spec, fmt.Sprintf("%s/spec", pathprefix), map[string]*metav1.ObjectMeta{
			fmt.Sprintf("%s/metadata", pathprefix): &o.ObjectMeta,
		})
	case *corev1.List:
		return g.handleList(o, pathprefix)
	}

	return make(k8tz.Patches, 0), fmt.Errorf("not injectable object: %T", object)
}

func (g *PatchGenerator) handleList(list *corev1.List, pathprefix string) (patches k8tz.Patches, err error) {
	patches = k8tz.Patches{}
	if len(list.Items) == 0 {
		return patches, nil
	}

	for i, v := range list.Items {
		obj, err := parseTypeMetaSkeleton(v.Raw)
		if err != nil {
			return patches, err
		}

		if obj == nil {
			continue
		}

		err = yaml.Unmarshal(v.Raw, obj)
		if err != nil {
			return patches, err
		}

		vpatch, err := g.Generate(obj, fmt.Sprintf("%s/items/%d", pathprefix, i))
		if err != nil {
			return patches, err
		}

		patches = append(patches, vpatch...)
	}

	return patches, nil
}

func (g *PatchGenerator) generate(spec *corev1.PodSpec, pathprefix string, postInjectionAnnotations map[string]*metav1.ObjectMeta) (patches k8tz.Patches, err error) {
	if g.Strategy == HostPathInjectionStrategy {
		patches = append(patches, g.createHostPathPatches(spec, pathprefix)...)
	} else if g.Strategy == InitContainerInjectionStrategy {
		patches = append(patches, g.createInitContainerPatches(spec, pathprefix)...)
	} else {
		return nil, fmt.Errorf("unknown injection strategy specified: %s", g.Strategy)
	}

	patches = append(patches, g.createEnvironmentVariablePatches(spec, pathprefix)...)

	for k, v := range postInjectionAnnotations {
		patches = append(patches, g.createPostInjectionAnnotations(v, k)...)
	}

	return patches, nil
}

func (g *PatchGenerator) createEnvironmentVariablePatches(spec *corev1.PodSpec, pathprefix string) k8tz.Patches {
	var patches = k8tz.Patches{}

	for containerId := 0; containerId < len(spec.Containers); containerId++ {
		if len(spec.Containers[containerId].Env) == 0 {
			patches = append(patches, k8tz.Patch{
				Op:    "add",
				Path:  fmt.Sprintf("%s/containers/%d/env", pathprefix, containerId),
				Value: []corev1.EnvVar{},
			})
		}

		patches = append(patches, k8tz.Patch{
			Op:   "add",
			Path: fmt.Sprintf("%s/containers/%d/env/-", pathprefix, containerId),
			Value: corev1.EnvVar{
				Name:  "TZ",
				Value: g.Timezone,
			},
		})
	}

	return patches
}

func (g *PatchGenerator) createInitContainerPatches(spec *corev1.PodSpec, pathprefix string) k8tz.Patches {
	var patches = k8tz.Patches{}

	containers := len(spec.Containers)
	if containers == 0 {
		return patches
	}

	if len(spec.Volumes) == 0 {
		patches = append(patches, k8tz.Patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/volumes", pathprefix),
			Value: []corev1.Volume{},
		})
	}

	patches = append(patches, k8tz.Patch{
		Op:   "add",
		Path: fmt.Sprintf("%s/volumes/-", pathprefix),
		Value: corev1.Volume{
			Name: "k8tz",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	})

	for containerId := 0; containerId < containers; containerId++ {
		if len(spec.Containers[containerId].VolumeMounts) == 0 {
			patches = append(patches, k8tz.Patch{
				Op:    "add",
				Path:  fmt.Sprintf("%s/containers/%d/volumeMounts", pathprefix, containerId),
				Value: []corev1.VolumeMount{},
			})
		}

		patches = append(patches, k8tz.Patch{
			Op:   "add",
			Path: fmt.Sprintf("%s/containers/%d/volumeMounts/-", pathprefix, containerId),
			Value: corev1.VolumeMount{
				Name:      "k8tz",
				ReadOnly:  true,
				MountPath: g.LocalTimePath,
				SubPath:   g.Timezone,
			},
		})

		patches = append(patches, k8tz.Patch{
			Op:   "add",
			Path: fmt.Sprintf("%s/containers/%d/volumeMounts/-", pathprefix, containerId),
			Value: corev1.VolumeMount{
				Name:      "k8tz",
				ReadOnly:  true,
				MountPath: "/usr/share/zoneinfo",
			},
		})
	}

	if len(spec.InitContainers) == 0 {
		patches = append(patches, k8tz.Patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/initContainers", pathprefix),
			Value: []corev1.Container{},
		})
	}

	patches = append(patches, k8tz.Patch{
		Op:   "add",
		Path: fmt.Sprintf("%s/initContainers/-", pathprefix),
		Value: corev1.Container{
			Name:  "k8tz",
			Image: g.InitContainerImage,
			Args:  []string{"bootstrap"},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "k8tz",
					MountPath: "/mnt/zoneinfo",
					ReadOnly:  false,
				},
			},
		},
	})

	return patches
}

func (g *PatchGenerator) createHostPathPatches(spec *corev1.PodSpec, pathprefix string) k8tz.Patches {
	var patches = k8tz.Patches{}
	containers := len(spec.Containers)
	if containers == 0 {
		return patches
	}

	for containerId := 0; containerId < containers; containerId++ {
		if len(spec.Containers[containerId].VolumeMounts) == 0 {
			patches = append(patches, k8tz.Patch{
				Op:    "add",
				Path:  fmt.Sprintf("%s/containers/%d/volumeMounts", pathprefix, containerId),
				Value: []corev1.VolumeMount{},
			})
		}

		patches = append(patches, k8tz.Patch{
			Op:   "add",
			Path: fmt.Sprintf("%s/containers/%d/volumeMounts/-", pathprefix, containerId),
			Value: corev1.VolumeMount{
				Name:      "k8tz",
				ReadOnly:  true,
				MountPath: g.LocalTimePath,
				SubPath:   g.Timezone,
			},
		})

		patches = append(patches, k8tz.Patch{
			Op:   "add",
			Path: fmt.Sprintf("%s/containers/%d/volumeMounts/-", pathprefix, containerId),
			Value: corev1.VolumeMount{
				Name:      "k8tz",
				ReadOnly:  true,
				MountPath: "/usr/share/zoneinfo",
			},
		})
	}

	if len(spec.Volumes) == 0 {
		patches = append(patches, k8tz.Patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/volumes", pathprefix),
			Value: []corev1.Volume{},
		})
	}

	patches = append(patches, k8tz.Patch{
		Op:   "add",
		Path: fmt.Sprintf("%s/volumes/-", pathprefix),
		Value: corev1.Volume{
			Name: "k8tz",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: g.HostPathPrefix,
				},
			},
		},
	})

	return patches
}

func (g *PatchGenerator) createPostInjectionAnnotations(meta *metav1.ObjectMeta, pathprefix string) k8tz.Patches {
	var patches = k8tz.Patches{}
	if len(meta.Annotations) == 0 {
		patches = append(patches, k8tz.Patch{
			Op:    "add",
			Path:  fmt.Sprintf("%s/annotations", pathprefix),
			Value: map[string]string{},
		})
	}

	patches = append(patches, k8tz.Patch{
		Op:    "add",
		Path:  fmt.Sprintf("%s/annotations/%s", pathprefix, escapeJsonPointer(k8tz.InjectedAnnotation)),
		Value: "true",
	})
	patches = append(patches, k8tz.Patch{
		Op:    "add",
		Path:  fmt.Sprintf("%s/annotations/%s", pathprefix, escapeJsonPointer(k8tz.TimezoneAnnotation)),
		Value: g.Timezone,
	})

	return patches
}

// TODO: unit test
func escapeJsonPointer(p string) string {
	return jsonPointerEscapeReplacer.Replace(p)
}
