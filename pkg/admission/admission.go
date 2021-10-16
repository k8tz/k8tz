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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	k8tz "github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/inject"
	"github.com/k8tz/k8tz/pkg/version"
	admission "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type RequestsHandler struct {
	DefaultTimezone          string
	BootstrapImage           string
	DefaultInjectionStrategy inject.InjectionStrategy
	InjectByDefault          bool
	HostPathPrefix           string
	LocalTimePath            string
	clientset                kubernetes.Interface
}

func NewRequestsHandler() RequestsHandler {
	return RequestsHandler{
		DefaultTimezone:          k8tz.DefaultTimezone,
		BootstrapImage:           version.Image(),
		DefaultInjectionStrategy: inject.DefaultInjectionStrategy,
		InjectByDefault:          true,
		HostPathPrefix:           inject.DefaultHostPathPrefix,
		LocalTimePath:            inject.DefaultLocalTimePath,
	}
}

func getKubeconfig(kubeconfPath string) (*restclient.Config, error) {
	if kubeconfPath == "" {
		warningLogger.Println("--kubeconfig not specified. Using the inClusterConfig. This might not work.")
		kubeconfig, err := restclient.InClusterConfig()
		if err == nil {
			return kubeconfig, nil
		}

		warningLogger.Println("error creating inClusterConfig, falling back to default config.")
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfPath},
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
}

func (h *RequestsHandler) InitializeClientset(kubeconfPath string) error {
	config, err := getKubeconfig(kubeconfPath)
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	clientset, _ := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create k8s client: %v", err)
	}

	h.clientset = clientset
	return nil
}

func (h *RequestsHandler) handleFunc(w http.ResponseWriter, r *http.Request) {
	review, header, err := h.readAdmissionReview(r)
	if err != nil {
		warningLogger.Printf("failed to parse review: %v\n", err)
		http.Error(w, fmt.Sprintf("failed to parse admission review from request, error: %s", err.Error()), header)
		return
	}

	reviewResponse := admission.AdmissionReview{
		TypeMeta: review.TypeMeta,
		Response: &admission.AdmissionResponse{
			UID: review.Request.UID,
		},
	}

	patches, err := h.handleAdmissionReview(review)
	if err != nil {
		warningLogger.Printf("rejecting request (%s/%s): %v\n", review.Request.Namespace, review.Request.Name, err)
		reviewResponse.Response.Allowed = false
		reviewResponse.Response.Result = &metav1.Status{
			Message: err.Error(),
		}
	} else {
		patchBytes, err := json.Marshal(patches)
		if err != nil {
			errorLogger.Printf("failed to marshal json patch: %v, error: %v\n", patches, err)
			http.Error(w, fmt.Sprintf("could not marshal JSON patch: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		reviewResponse.Response.Patch = patchBytes
		reviewResponse.Response.PatchType = new(admission.PatchType)
		*reviewResponse.Response.PatchType = admission.PatchTypeJSONPatch
		reviewResponse.Response.Allowed = true
		infoLogger.Printf("accepting request (%s/%s), %d patches generated\n", review.Request.Namespace, review.Request.Name, len(patches))
	}

	bytes, err := json.Marshal(&reviewResponse)
	if err != nil {
		errorLogger.Printf("failed to marshal response review: %v, error: %v\n", reviewResponse, err)
		http.Error(w, fmt.Sprintf("failed to marshal response review: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(bytes)
	if err != nil {
		errorLogger.Printf("failed to write response to output http stream: %v\n", err)
		http.Error(w, fmt.Sprintf("failed to write response: %s", err.Error()), http.StatusInternalServerError)
	}
}

func (h *RequestsHandler) handleAdmissionReview(review *admission.AdmissionReview) (k8tz.Patches, error) {
	if review.Request.Resource != podResource || review.Request.Operation != "CREATE" {
		return nil, nil
	}

	var patches k8tz.Patches
	var err error
	if review.Request.Namespace != metav1.NamespacePublic && review.Request.Namespace != metav1.NamespaceSystem {
		patches, err = h.handleAdmissionRequest(review.Request)
		if err != nil {
			return nil, err
		}
	}

	return patches, nil
}

func (h *RequestsHandler) readAdmissionReview(r *http.Request) (*admission.AdmissionReview, int, error) {
	if r.Method != http.MethodPost {
		return nil, http.StatusMethodNotAllowed, fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("could not read request body, error: %s", err.Error())
	}

	if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		return nil, http.StatusBadRequest, fmt.Errorf("unsupported content type %s, only %s is supported", contentType, jsonContentType)
	}

	review := &admission.AdmissionReview{}
	if _, _, err := k8sdecode.Decode(body, nil, review); err != nil {
		return nil, http.StatusBadRequest, fmt.Errorf("could not deserialize request to review object: %v", err)
	} else if review.Request == nil {
		return nil, http.StatusBadRequest, errors.New("review parsed but request is null")
	}

	return review, http.StatusOK, nil
}

func (h *RequestsHandler) lookupGenerator(namespace string, pod *corev1.Pod) (*inject.PatchGenerator, error) {
	namespaceObj, err := h.clientset.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to lookup pod's namespace (%s/%s): %v", namespace, pod.Name, err)
	}

	if _, ok := pod.Annotations[k8tz.InjectedAnnotation]; ok {
		infoLogger.Printf("skipping pod (%s/%s) because its already injected", namespace, pod.Name)
		return nil, nil
	}

	if val, ok := pod.Annotations[k8tz.InjectAnnotation]; ok {
		if val == "false" {
			infoLogger.Printf("skipping pod (%s/%s) because annotation on pod is explicitly false for injection", namespace, pod.Name)
			return nil, nil
		}
	} else if val, ok := namespaceObj.Annotations[k8tz.InjectAnnotation]; ok {
		if val == "false" {
			infoLogger.Printf("skipping pod (%s/%s) because annotation on namespace is explicitly false for injection", namespace, pod.Name)
			return nil, nil
		}
	} else if !h.InjectByDefault {
		infoLogger.Printf("skipping pod (%s/%s) because no other instruction and injection disabled by default", namespace, pod.Name)
		return nil, nil
	}

	timezone := h.DefaultTimezone
	if val, ok := pod.Annotations[k8tz.TimezoneAnnotation]; ok {
		timezone = val
		infoLogger.Printf("explicit timezone requested on pod's (%s/%s) annotation: %s", namespace, pod.Name, val)
	} else if val, ok := namespaceObj.Annotations[k8tz.TimezoneAnnotation]; ok {
		timezone = val
		infoLogger.Printf("explicit timezone requested on namespace (%s/%s) annotation: %s", namespace, pod.Name, val)
	}

	strategy := h.DefaultInjectionStrategy
	if v, e := pod.Annotations[k8tz.InjectionStrategyAnnotation]; e {
		strategy = inject.InjectionStrategy(v)
		infoLogger.Printf("explicit injection strategy requested on pod's (%s/%s) annotation: %s", namespace, pod.Name, v)
	} else if v, e := namespaceObj.Annotations[k8tz.InjectionStrategyAnnotation]; e {
		strategy = inject.InjectionStrategy(v)
		infoLogger.Printf("explicit injection strategy requested on namespace (%s/%s) annotation: %s", namespace, pod.Name, v)
	}

	return &inject.PatchGenerator{
		Strategy:           strategy,
		Timezone:           timezone,
		InitContainerImage: h.BootstrapImage,
		HostPathPrefix:     h.HostPathPrefix,
		LocalTimePath:      h.LocalTimePath,
	}, nil
}

func (h *RequestsHandler) handleAdmissionRequest(req *admission.AdmissionRequest) (k8tz.Patches, error) {
	raw := req.Object.Raw
	pod := corev1.Pod{}
	if _, _, err := k8sdecode.Decode(raw, nil, &pod); err != nil {
		return nil, fmt.Errorf("could not deserialize pod object: %v", err)
	}

	generator, err := h.lookupGenerator(req.Namespace, &pod)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup generator, error: %w", err)
	}

	var patches k8tz.Patches
	if generator != nil {
		patches, err = generator.Generate(&pod, "")
		if err != nil {
			return nil, fmt.Errorf("failed to generate patches for pod, error: %w", err)
		}
	}

	return patches, err
}
