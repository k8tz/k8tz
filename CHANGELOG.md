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
