package watcher

type Watcher struct {
	TLSCertFile     string
	TLSKeyFile      string
	SecretName      string
	SecretNamespace string
	Verbose         bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		TLSCertFile:     "/run/secrets/tls/tls.crt",
		TLSKeyFile:      "/run/secrets/tls/tls.key",
		SecretName:      "k8tz-tls",
		SecretNamespace: "k8tz",
		Verbose:         false,
	}
}
