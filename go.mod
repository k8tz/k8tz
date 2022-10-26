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

replace (
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83 => golang.org/x/crypto v0.0.0-20220314234659-1baeb1ce4c0b // required for CVE-2022-27191
	golang.org/x/net v0.0.0-20210520170846-37e1c6afe023 => golang.org/x/net v0.0.0-20220906165146-f3363e06e74c // required for CVE-2022-27664, CVE-2021-44716
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 => golang.org/x/net v0.0.0-20220906165146-f3363e06e74c // required for CVE-2022-27664, CVE-2021-44716
	golang.org/x/net v0.0.0-20220722155237-a158d28d115b => golang.org/x/net v0.0.0-20220906165146-f3363e06e74c // required for CVE-2022-27664, CVE-2021-44716
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 => golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // required for CVE-2022-29526
	golang.org/x/text v0.3.6 => golang.org/x/text v0.3.8 // required for CVE-2021-38561, CVE-2022-32149
	golang.org/x/text v0.3.7 => golang.org/x/text v0.3.8 // required for CVE-2022-32149
)

exclude (
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633
	github.com/miekg/dns v1.0.14
)
