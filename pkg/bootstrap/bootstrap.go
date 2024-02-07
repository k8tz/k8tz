package bootstrap

import (
	k8tz "github.com/k8tz/k8tz/pkg"
	"github.com/k8tz/k8tz/pkg/inject"
	"github.com/k8tz/k8tz/pkg/version"
	"os"
)

type BootstrapOperation struct {
	From      string
	To        string
	Overwrite bool
	Verbose   bool
}

func NewBootstrapOperation() BootstrapOperation {
	return BootstrapOperation{
		From:      inject.DefaultHostPathPrefix,
		To:        "/mnt/zoneinfo",
		Overwrite: true,
		Verbose:   false,
	}
}

func (o *BootstrapOperation) Bootstrap() error {
	if o.Verbose {
		k8tz.VerboseLogger.SetOutput(os.Stderr)
		k8tz.VerboseLogger.Println(version.DisplayVersion())
		k8tz.VerboseLogger.Printf("bootstrap=%+v", *o)
	}
	return copyDirectory(o.From, o.To, o.Overwrite)
}
