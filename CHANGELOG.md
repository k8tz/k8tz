# 0.18.1
- Adds `injectedInitContainerImagePullPolicy` to helm chart and `container-imagepullpolicy` flag to set imagePullPolicy of the injected init container

# 0.18.0

- Update tzdata package to [2025b](https://github.com/k8tz/k8tz/pull/116)
- Bump golang.org/x/net from 0.23.0 to 0.36.0 (https://github.com/k8tz/k8tz/pull/117)

# 0.17.1

- Helm: don't render secret permissions when cert-manager is not used (https://github.com/k8tz/k8tz/pull/110)

# 0.17.0

- Update tzdata package to [2024b](https://github.com/k8tz/k8tz/pull/107)

# 0.16.2

- Fix volumeMounts volume name not found error when using custom initContainer name (https://github.com/k8tz/k8tz/pull/102)

# 0.16.1

- Add possibility to set init container resources (https://github.com/k8tz/k8tz/pull/85)

# 0.16.0

- Update tzdata package to [2024a](https://mm.icann.org/pipermail/tz-announce/2024-February/000081.html) (https://github.com/k8tz/k8tz/pull/96)
- Add verbose flag to bootstrap initContainer (https://github.com/k8tz/k8tz/pull/95)

# 0.15.0

- Update tzdata package to [2023d](https://github.com/k8tz/k8tz/pull/94)

# 0.14.1

- Allow to override injected initContainer name (https://github.com/k8tz/k8tz/pull/84)
- Helm: Allow using helm built-in namespace (https://github.com/k8tz/k8tz/pull/92)
- Helm: Add possibility to set topologySpreadConstraints to k8tz controller (https://github.com/k8tz/k8tz/pull/90)
- Bump golang.org/x/net from 0.7.0 to 0.17.0 (https://github.com/k8tz/k8tz/pull/89)
- Upgrade to go 1.21 (https://github.com/k8tz/k8tz/pull/87)

# 0.14.0

- Add flags `--tls-cipher-suites` and `--tls-min-version` to the admission controller (https://github.com/k8tz/k8tz/pull/75, https://github.com/k8tz/k8tz/pull/78)
- Don't explicitly set `runAsNonRoot` to avoid conflict with pod security context (https://github.com/k8tz/k8tz/pull/82)
- Skip `kube-system` namespace early and add configurable `ignoredNamespaces` option (https://github.com/k8tz/k8tz/pull/79)

# 0.13.1

- Prevent null pointer de-referencing on cronjobs (https://github.com/k8tz/k8tz/pull/73)

# 0.13.0

- Update tzdata package to [2023c](https://mm.icann.org/pipermail/tz-announce/2023-March/000079.html) (https://github.com/k8tz/k8tz/pull/67)

# 0.12.0

- Update tzdata package to [2023b](https://mm.icann.org/pipermail/tz-announce/2023-March/000078.html) (https://github.com/k8tz/k8tz/pull/65)
- Helm: Add possibility to add custom labels to k8tz resources (https://github.com/k8tz/k8tz/pull/59)
- Add restricted SecurityContext to controller and initContainers (https://github.com/k8tz/k8tz/pull/58)
- Automatically update webhook TLS key pair for cert-manager (https://github.com/k8tz/k8tz/pull/55)
- Upgrade go to 1.18 and update dependencies to latest (CVE-2022-41721, CVE-2022-41717, CVE-2022-41723, CVE-2022-28948) (https://github.com/k8tz/k8tz/pull/61)

# 0.11.0

- Update tzdata package to [2022g](https://mm.icann.org/pipermail/tz-announce/2022-November/000076.html) (https://github.com/k8tz/k8tz/pull/48)
- Helm: add option to use daemonset instead of deployment (https://github.com/k8tz/k8tz/pull/44)
- Helm: add support for managing certificate with cert-manager (https://github.com/k8tz/k8tz/pull/45)

# 0.10.0

- Update tzdata package to [2022f](https://mm.icann.org/pipermail/tz-announce/2022-October/000075.html)

# 0.9.0

- Update tzdata package to [2022e](https://mm.icann.org/pipermail/tz-announce/2022-October/000074.html)
- Resolve: CVE-2022-27664, CVE-2021-44716, CVE-2022-32149

# 0.8.0

- Update tzdata package to [2022d](https://mm.icann.org/pipermail/tz-announce/2022-September/000073.html)

# 0.7.0

- Update tzdata package to [2022c](https://mm.icann.org/pipermail/tz-announce/2022-August/000072.html) (https://github.com/k8tz/k8tz/pull/39)
- Resolve: CVE-2022-27191, CVE-2021-44716 and CVE-2022-29526 (https://github.com/k8tz/k8tz/pull/40)

# 0.6.0

- Update tzdata package to [2022b](https://mm.icann.org/pipermail/tz-announce/2022-August/000071.html) (https://github.com/k8tz/k8tz/pull/37)
- Resolve: CVE-2022-1996, CVE-2019-19794, CVE-2021-38561 ([bc315ad8d](https://github.com/k8tz/k8tz/commit/bc315ad8dfcd73463f2d44bff50d56f08f477aec))
- Separate README.md for helm chart (https://github.com/k8tz/k8tz/pull/34)
- Add icon, home, sources and keywords to Chart.yaml (https://github.com/k8tz/k8tz/pull/33)

# v0.5.2

- Fix conflict with existing volumeMounts (https://github.com/k8tz/k8tz/pull/27)
- Improve admission controller logs (https://github.com/k8tz/k8tz/pull/30)
- Make tzdata image as k8tz base image (https://github.com/k8tz/k8tz/pull/29)

# v0.5.1

- Run container as non-root user (https://github.com/k8tz/k8tz/pull/25)

# v0.5.0

- Add CronJob support (for kubernetes >=1.24.0-beta.0) (https://github.com/k8tz/k8tz/pull/17)
- Hide confusing log message about `--kubeconfig` not specified (https://github.com/k8tz/k8tz/pull/19)

# v0.4.0

- Add StatefulSet transformation support (https://github.com/k8tz/k8tz/pull/15)
- Update tzdata package to 2022a (https://github.com/k8tz/k8tz/pull/14)

# v0.3.0

- First public release of k8tz
