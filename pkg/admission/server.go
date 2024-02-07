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
	"crypto/tls"
	"fmt"
	k8tz "github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/version"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	cliflag "k8s.io/component-base/cli/flag"
)

const (
	jsonContentType = `application/json`
)

var (
	k8sdecode       = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
	podResource     = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	cronJobResource = metav1.GroupVersionResource{Version: "v1", Resource: "cronjobs", Group: "batch"}
)

type Server struct {
	TLSCertFile     string
	TLSKeyFile      string
	TLSCipherSuites []string
	TLSMinVersion   string
	Address         string
	Handler         RequestsHandler
	Verbose         bool
}

func NewAdmissionServer() *Server {
	return &Server{
		TLSCertFile: "/run/secrets/tls/tls.crt",
		TLSKeyFile:  "/run/secrets/tls/tls.key",
		Address:     ":8443",
		Handler:     NewRequestsHandler(),
		Verbose:     false,
	}
}

func (h *Server) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Server) Start(kubeconfigFlag string) error {
	k8tz.InfoLogger.Println(version.DisplayVersion())

	if h.Verbose {
		k8tz.VerboseLogger.SetOutput(os.Stderr)
		k8tz.VerboseLogger.Printf("server=%+v", *h)
	}
	minTLSVersion, err := cliflag.TLSVersion(h.TLSMinVersion)
	if err != nil {
		return err
	}
	tlsCipherSuites, err := cliflag.TLSCipherSuites(h.TLSCipherSuites)
	if err != nil {
		return err
	}

	if err = h.Handler.InitializeClientset(kubeconfigFlag); err != nil {
		return fmt.Errorf("failed to setup connection with kubernetes api: %w", err)
	}

	k8tz.InfoLogger.Printf("Listening on %s\n", h.Address)

	mux := http.NewServeMux()

	mux.HandleFunc("/", h.Handler.handleFunc)
	mux.HandleFunc("/health", h.health)

	server := &http.Server{
		Addr:    h.Address,
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
				cert, err := tls.LoadX509KeyPair(h.TLSCertFile, h.TLSKeyFile)
				if err != nil {
					return nil, err
				}
				return &cert, nil
			},
			CipherSuites: tlsCipherSuites,
			MinVersion:   minTLSVersion,
		},
	}

	return server.ListenAndServeTLS("", "")
}
