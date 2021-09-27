package bootstrap

import "github.com/k8tz/k8tz/pkg/inject"

type BootstrapOperation struct {
	From      string
	To        string
	Overwrite bool
}

func NewBootstrapOperation() BootstrapOperation {
	return BootstrapOperation{
		From:      inject.DefaultHostPathPrefix,
		To:        "/mnt/zoneinfo",
		Overwrite: true,
	}
}

func (o *BootstrapOperation) Bootstrap() error {
	return copyDirectory(o.From, o.To, o.Overwrite)
}
