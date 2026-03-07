# Kubernetes Timezone Controller

![Build Workflow Status](https://img.shields.io/github/actions/workflow/status/k8tz/k8tz/go.yaml?branch=master)
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://goreportcard.com/report/github.com/k8tz/k8tz)
[![codecov](https://codecov.io/gh/k8tz/k8tz/branch/master/graph/badge.svg?token=3HEoptX1C0)](https://codecov.io/gh/k8tz/k8tz)
[![Go Version](https://img.shields.io/github/go-mod/go-version/k8tz/k8tz)](go.mod)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

![k8tz Logo](assets/k8tz-logo-blue-transparent-medium.png)

`k8tz` is a kubernetes admission controller and a CLI tool to inject timezones into Pods and CronJobs[^1].

Containers do not inherit timezones from host machines and have only accessed to the clock from the kernel. The default timezone for most images is UTC, yet it is not guaranteed and may be different from container to container. With `k8tz` it is easy to standardize selected timezone across pods and namespaces automatically with minimal effort.

## Features

:zap: Coordinate timezone for all pods in clusters and/or namespaces (force UTC by default)

:zap: Standardize [tzdata](https://www.iana.org/time-zones) version across all pods in cluster

:zap: Does not require [tzdata](https://www.iana.org/time-zones) installed in images or nodes

:zap: Easy to configure with [Helm values](charts/k8tz/README.md#values) and [Annotations](#annotations)

:zap: [CLI tools](#cli) for manual timezone injection

:zap: Supports Kubernetes 1.16+ and OpenShift 4.X

Read more: [Timezone in Kubernetes With k8tz](https://medium.com/@yonatankahana/timezone-in-kubernetes-with-k8tz-fdefca785238)

## Install Admission Controller (Helm)

![Short Demo](assets/k8tz-helm-demo.gif)

tl;dr:

```console
helm repo add k8tz https://k8tz.github.io/k8tz/
helm install k8tz k8tz/k8tz --set timezone=Europe/London
```

Read more in the chart [README](charts/k8tz/README.md).

## CLI

`k8tz` can be used as a command-line tool to inject timezone into yaml files or to be integrated inside another deployment script that don't want to use the admission controller automation.

### Examples

You can process the `test-pod.yaml` from file-to-file or directly to `kubectl`:

```console
# to a file
k8tz inject --strategy=hostPath test-pod.yaml > injected-test-pod.yaml

# or directly to kubectl
k8tz inject --timezone=Europe/London test-pod.yaml | kubectl apply -f -
```

Or you can inject to all existing deployments in current namespace:

```console
kubectl get deploy -oyaml | k8tz inject - | kubectl apply -f -
```

NOTE: The injection process is idempotent; you can do it multiple times and/or use the CLI injection alongside the admission controller. Subsequent injections have no effect.

### Download GitHub Release

You can install k8tz binary file by downloading precompiled binary and use it

```console
wget -c https://github.com/k8tz/k8tz/releases/download/v0.18.1/k8tz_0.18.1_linux_amd64.tar.gz -O - | tar xz
chmod +x k8tz
./k8tz version
```

then install it to your `$PATH` with:

```console
sudo install k8tz /usr/local/bin/k8tz
```

### Go Install

If you have `go` installed, you can install `k8tz` with:

```console
go install github.com/k8tz/k8tz@latest
```

### Use Docker

You can use k8tz directly from Docker, here are some examples:

```console
docker run -i quay.io/k8tz/k8tz --help

cat test-pod.yaml | docker run -i quay.io/k8tz/k8tz inject -tPortugal - | kubectl create -f

kubectl get deploy -oyaml | docker run -i quay.io/k8tz/k8tz inject - | kubectl apply -f
```

### From Source

You can build `k8tz` binary from source yourself by simple running:

```console
make compile
```

The created binary will be located at `build/k8tz`, you can then install it to your PATH using:

```console
make TARGET=/usr/local/bin install
```

To uninstall, use `sudo rm -v /usr/local/bin/k8tz`.

## Injection Strategy

Timezone information is defined using Time Zone Information Format files (`TZif`, [RFC-8536](https://datatracker.ietf.org/doc/html/rfc8536)). The Timezone Database contains `TZif` files that represent the local time for many locations around the globe. To set the container's timezone, `/etc/localtime` inside the container should point to a valid `TZif` file which represents the requested timezone. In most images these files do not exist by default, so we need to make them available from inside the container mounted at `/etc/localtime`.

Currently, there are 2 strategies how it can be done:

### Using **hostPath**

If those files (which are located under `/usr/share/zoneinfo`) exist in every node on the cluster (it is the user's responsibility to ensure that), `hostPath` volume can be used to supply the required `TZif` file into the pod. If the required timezone will be missing on the host machine, the pod will be stuck in `PodInitializing` status and will not be started.

### Using bootstrap **initContainer**

Another solution, which is generally safer, is to inject `initContainer` (bootstrap image) to the pod and supply the required `TZif` file using a shared `emptyDir` volume. This is the default method of k8tz.

## Annotations

The behaviour of the controller can be changed using annotations on both `Pod` and/or `Namespace` objects. If the same annotation specified in both, the `Pod`'s annotation value will take place.

| Annotation         | Description                                                            | Default         |
|--------------------|------------------------------------------------------------------------|-----------------|
| `k8tz.io/inject`   | Decide whether k8tz should inject timezone or not                      | `true`          |
| `k8tz.io/timezone` | Decide what timezone should be used, e.g: `Africa/Addis_Ababa`         | `UTC`           |
| `k8tz.io/strategy` | Decide what injection strategy to use, i.e: `hostPath`/`initContainer` | `initContainer` |

## Roadmap

- [X] Support `StatefulSet` injection
- [X] Support `CronJob` injection
- [ ] Better way to lookup pod owner annotations
- [X] Test and document installation on OpenShift
- [X] Implement `make install` for easier installation from source
- [X] Add VERBOSE flag to helm
- [X] Write verbose logs for webhook
- [X] Separate README for Helm chart

[^1]: Timezones for CronJobs are available only from kubernetes >=1.24.0-beta.0 with [`CronJobTimeZone`](https://github.com/kubernetes/enhancements/blob/aad71056d33eccf3845b73670106f06a9e74fec6/keps/sig-apps/3140-TimeZone-support-in-CronJob/README.md) feature gate enabled.
