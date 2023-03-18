/*
Copyright © 2023 Andika Ahmad Ramadhan

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

package certwatcher

import (
	"context"
	"os"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCertWatcher_startWatcher(t *testing.T) {
	type fields struct {
		NewTLSCert string
		NewTLSKey  string
	}
	tests := []struct {
		fields fields
	}{
		{
			fields: fields{
				NewTLSCert: "dGVzdDEK",
				NewTLSKey:  "dGVzdDEK",
			},
		},
		{
			fields: fields{
				NewTLSCert: "dGVzdDIK",
				NewTLSKey:  "dGVzdDIK",
			},
		},
		{
			fields: fields{
				NewTLSCert: "dGVzdDMK",
				NewTLSKey:  "dGVzdDMK",
			},
		},
	}

	tempdir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())

	cw := &CertWatcher{
		TLSCertFile:     tempdir + "/tls.crt",
		TLSKeyFile:      tempdir + "/tls.key",
		SecretName:      "k8tz-tls",
		SecretNamespace: "k8tz",
		clientset:       fake.NewSimpleClientset(),
		ctx:             ctx,
		cancel:          cancel,
	}

	t.Run("create mock for secret resource", func(t *testing.T) {
		if _, err := cw.clientset.CoreV1().Secrets(cw.SecretNamespace).Create(cw.ctx, &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      cw.SecretName,
				Namespace: cw.SecretNamespace,
			},
			Data: map[string][]byte{
				"tls.crt": []byte(""),
				"tls.key": []byte(""),
			},
		}, v1.CreateOptions{}); err != nil {
			t.Fatal(err)
		}
	})

	go func() {
		t.Run("simulate secret resource changes", func(t *testing.T) {
			for _, test := range tests {
				// updating k8tz-tls secret
				secret, err := cw.clientset.CoreV1().Secrets(cw.SecretNamespace).Get(cw.ctx, cw.SecretName, v1.GetOptions{})
				if err != nil {
					t.Errorf("error get secret: %v", err)
				}
				secret.Data["tls.crt"] = []byte(test.fields.NewTLSCert)
				secret.Data["tls.key"] = []byte(test.fields.NewTLSKey)
				_, err = cw.clientset.CoreV1().Secrets(cw.SecretNamespace).Update(cw.ctx, secret, v1.UpdateOptions{})
				if err != nil {
					t.Errorf("error update secret: %v", err)
				}
				time.Sleep(1 * time.Second)
				// checking k8tz-tls secret
				if tlsCrt, err := os.ReadFile(cw.TLSCertFile); err != nil {
					t.Errorf("error read tls.crt: %v", err)
				} else if string(tlsCrt) != test.fields.NewTLSCert {
					t.Logf("tls.crt data missmatch with data from %s: %v", cw.TLSCertFile, err)
					t.Logf("expecting: %s", test.fields.NewTLSCert)
					t.Logf("got: %s", string(tlsCrt))
					t.Fail()
				}
				if tlsKey, err := os.ReadFile(cw.TLSKeyFile); err != nil {
					t.Errorf("error read tls.key: %v", err)
				} else if string(tlsKey) != test.fields.NewTLSKey {
					t.Logf("tls.key data missmatch with data from %s: %v", cw.TLSKeyFile, err)
					t.Logf("expecting: %s", test.fields.NewTLSKey)
					t.Logf("got: %s", string(tlsKey))
					t.Fail()
				}
			}
			cw.cancel()
		})
	}()

	t.Run("run cert-watcher with kubernetes api mock", func(t *testing.T) {
		if err := cw.startWatcher(); err != nil {
			t.Errorf("TestCertWatcher_startWatcher: %v", err)
		}
	})
}
