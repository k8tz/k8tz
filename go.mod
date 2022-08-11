module github.com/k8tz/k8tz

go 1.16

require (
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/spf13/cobra v1.2.1
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/yaml v1.2.0
)

replace golang.org/x/text v0.3.6 => golang.org/x/text v0.3.7

exclude (
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633
	github.com/miekg/dns v1.0.14
)
