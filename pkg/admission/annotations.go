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
	"context"

	k8tz "github.com/k8tz/k8tz/pkg"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This file resolves k8tz annotations for pod admission. Stable lookup uses
// Pod -> Namespace. Beta owner lookup, when enabled, uses
// Pod -> controller owner chain -> Namespace. Owner lookup is best-effort so
// parent API errors do not block pod admission.

// maxOwnerAnnotationDepth prevents unexpected owner-reference loops from
// causing unbounded API lookups.
const maxOwnerAnnotationDepth = 8

// annotationSource is one object in the annotation precedence chain.
type annotationSource struct {
	name        string
	annotations map[string]string
}

// lookupAnnotation returns the first value found for an annotation while
// scanning sources from closest to farthest.
func lookupAnnotation(sources []annotationSource, annotation string) (string, string, bool) {
	for _, source := range sources {
		if val, ok := source.annotations[annotation]; ok {
			return val, source.name, true
		}
	}

	return "", "", false
}

// lookupPodAnnotationSources builds the annotation source list for a pod,
// preserving the precedence expected by lookupAnnotation. Owner sources are
// included only when the beta pod owner lookup feature is enabled.
func (h *RequestsHandler) lookupPodAnnotationSources(namespace string, pod *corev1.Pod, namespaceObj *corev1.Namespace, includeOwners bool) []annotationSource {
	sources := []annotationSource{
		{
			name:        "pod",
			annotations: pod.Annotations,
		},
	}

	if includeOwners {
		sources = append(sources, h.lookupOwnerAnnotationSources(namespace, &pod.ObjectMeta, 0)...)
	}

	sources = append(sources, annotationSource{
		name:        "namespace",
		annotations: namespaceObj.Annotations,
	})

	return sources
}

// lookupOwnerAnnotationSources follows only the controller owner reference for
// the object and ignores non-controller owners.
func (h *RequestsHandler) lookupOwnerAnnotationSources(namespace string, objectMeta *metav1.ObjectMeta, depth int) []annotationSource {
	if depth >= maxOwnerAnnotationDepth {
		k8tz.WarningLogger.Printf("stopping pod owner annotation lookup for object (%s) after %d owner levels", formatObjectDetails(*objectMeta), maxOwnerAnnotationDepth)
		return nil
	}

	ownerRef := metav1.GetControllerOf(objectMeta)
	if ownerRef == nil {
		if len(objectMeta.OwnerReferences) > 0 {
			k8tz.VerboseLogger.Printf("ignoring non-controller owner references for object (%s)", formatObjectDetails(*objectMeta))
		}

		return nil
	}

	return h.lookupOwnerReferenceAnnotationSources(namespace, ownerRef, depth)
}

// lookupOwnerReferenceAnnotationSources fetches supported built-in owners and
// returns their annotations before continuing up the owner chain. Lookup errors
// and unsupported owners are logged and treated as missing parents.
func (h *RequestsHandler) lookupOwnerReferenceAnnotationSources(namespace string, ownerRef *metav1.OwnerReference, depth int) []annotationSource {
	switch {
	case ownerRef.APIVersion == "apps/v1" && ownerRef.Kind == "ReplicaSet":
		replicaSet, err := h.clientset.AppsV1().ReplicaSets(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			k8tz.WarningLogger.Printf("failed to lookup pod owner ReplicaSet namespace=%s, name=%s: %v", namespace, ownerRef.Name, err)
			return nil
		}

		sources := []annotationSource{{name: "replicaSet", annotations: replicaSet.Annotations}}
		return append(sources, h.lookupOwnerAnnotationSources(namespace, &replicaSet.ObjectMeta, depth+1)...)

	case ownerRef.APIVersion == "apps/v1" && ownerRef.Kind == "Deployment":
		deployment, err := h.clientset.AppsV1().Deployments(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			k8tz.WarningLogger.Printf("failed to lookup pod owner Deployment namespace=%s, name=%s: %v", namespace, ownerRef.Name, err)
			return nil
		}

		sources := []annotationSource{{name: "deployment", annotations: deployment.Annotations}}
		return append(sources, h.lookupOwnerAnnotationSources(namespace, &deployment.ObjectMeta, depth+1)...)

	case ownerRef.APIVersion == "apps/v1" && ownerRef.Kind == "StatefulSet":
		statefulSet, err := h.clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			k8tz.WarningLogger.Printf("failed to lookup pod owner StatefulSet namespace=%s, name=%s: %v", namespace, ownerRef.Name, err)
			return nil
		}

		sources := []annotationSource{{name: "statefulSet", annotations: statefulSet.Annotations}}
		return append(sources, h.lookupOwnerAnnotationSources(namespace, &statefulSet.ObjectMeta, depth+1)...)

	case ownerRef.APIVersion == "apps/v1" && ownerRef.Kind == "DaemonSet":
		daemonSet, err := h.clientset.AppsV1().DaemonSets(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			k8tz.WarningLogger.Printf("failed to lookup pod owner DaemonSet namespace=%s, name=%s: %v", namespace, ownerRef.Name, err)
			return nil
		}

		sources := []annotationSource{{name: "daemonSet", annotations: daemonSet.Annotations}}
		return append(sources, h.lookupOwnerAnnotationSources(namespace, &daemonSet.ObjectMeta, depth+1)...)

	case ownerRef.APIVersion == "batch/v1" && ownerRef.Kind == "Job":
		job, err := h.clientset.BatchV1().Jobs(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			k8tz.WarningLogger.Printf("failed to lookup pod owner Job namespace=%s, name=%s: %v", namespace, ownerRef.Name, err)
			return nil
		}

		sources := []annotationSource{{name: "job", annotations: job.Annotations}}
		return append(sources, h.lookupOwnerAnnotationSources(namespace, &job.ObjectMeta, depth+1)...)

	case ownerRef.APIVersion == "batch/v1" && ownerRef.Kind == "CronJob":
		cronJob, err := h.clientset.BatchV1().CronJobs(namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
		if err != nil {
			k8tz.WarningLogger.Printf("failed to lookup pod owner CronJob namespace=%s, name=%s: %v", namespace, ownerRef.Name, err)
			return nil
		}

		sources := []annotationSource{{name: "cronJob", annotations: cronJob.Annotations}}
		return append(sources, h.lookupOwnerAnnotationSources(namespace, &cronJob.ObjectMeta, depth+1)...)
	}

	k8tz.WarningLogger.Printf("ignoring unsupported pod controller owner namespace=%s, apiVersion=%s, kind=%s, name=%s", namespace, ownerRef.APIVersion, ownerRef.Kind, ownerRef.Name)
	return nil
}
