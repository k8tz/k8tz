# Kubernetes Timezone Controller - Helm Chart

![Lint Chart Workflow Status](https://img.shields.io/github/workflow/status/k8tz/k8tz/Lint%20Helm%20Charts?label=Lint)
![Build Chart Workflow Status](https://img.shields.io/github/workflow/status/k8tz/k8tz/Release%20Helm%20Charts?label=Release)
![Helm 2 Compatible](https://img.shields.io/badge/Helm%202-Compatible-blue)
![Helm 2 Compatible](https://img.shields.io/badge/Helm%203-Compatible-blue)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/k8tz)](https://artifacthub.io/packages/search?repo=k8tz)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.1-4baaaa.svg)](https://github.com/k8tz/k8tz/blob/master/CODE_OF_CONDUCT.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

![k8tz Logo](https://raw.githubusercontent.com/k8tz/k8tz/master/assets/k8tz-logo-blue-transparent-medium.png)

k8tz is a kubernetes admission controller and a CLI tool to inject timezones into Pods and CronJobs[^1].

## TL;DR

```console
helm repo add k8tz https://k8tz.github.io/k8tz/
helm install k8tz k8tz/k8tz --set timezone=Europe/London
```

## Prerequisites
- Helm 2 or later
- Kubernetes 1.16+ or OpenShift 4.X
- Permissions to use `emptyDir` or `hostPath`

## Installation

First, you need to add k8tz helm repository:

```console
helm repo add k8tz https://k8tz.github.io/k8tz/
```

Then, you can install k8tz admission controller:

```console
helm install k8tz k8tz/k8tz
```

If you want to set some values to configure k8tz, you can use `--set`, e.g:

```console
helm upgrade --install k8tz k8tz/k8tz \
    --set timezone=Europe/Amsterdam \
    --set injectionStrategy=hostPath
```

## Values

| Parameter                  | Description                                                                                                                                                                   | Default           |
|----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|
| replicaCount               | Amount of admission controller webhooks to spin up. For production use it is recommended to have at least 3 replicas                                                          | 1                 |
| namespace                  | The namespace where to install the admission controller                                                                                                                       | k8tz              |
| timezone                   | The default timezone to inject                                                                                                                                                | UTC               |
| injectionStrategy          | The default injection strategy to use                                                                                                                                         | initContainer     |
| injectAll                  | If true, timezone will be injected to the pod even when there is no annotation with explicit injection request. When false, the `k8tz.io/inject: true` annotation is required | true              |
| cronJobTimeZone            | Enable injection of `timeZone` field to `CronJob`s[^1]                                                                                                                        | false             |
| verbose                    | Enable more detailed logs for debug purposes                                                                                                                                  | false             |
| image.repository           | The image repository for the admission controller and bootstrap image                                                                                                         | quay.io/k8tz/k8tz |
| image.pullPolicy           | Admission controller image pull policy                                                                                                                                        | IfNotPresent      |
| image.tag                  | The image tag for the admission controller and bootstrap image. The default is the chart appVersion                                                                           | -                 |
| imagePullSecrets           | The image pull secrets for the admission controller                                                                                                                           | []                |
| nameOverride               | Helm application name override                                                                                                                                                | -                 |
| fullnameOverride           | Helm application full name override                                                                                                                                           | -                 |
| serviceAccount.annotations | Annotations to add to the admission controller service account                                                                                                                | {}                |
| serviceAccount.name        | The name of the service account to use. If empty, a name is generated using the fullname template                                                                             | -                 |
| podAnnotations             | Annotations to add to the admission controller pod                                                                                                                            | {}                |
| podSecurityContext         | Pod security context for the admission controller pod                                                                                                                         | {}                |
| securityContext            | Security context for the admission controller pod                                                                                                                             | {}                |
| service.type               | Admission controller service type                                                                                                                                             | ClusterIP         |
| service.port               | Admission controller service port                                                                                                                                             | 443               |
| resources                  | Resource requests and limitations for the admission controller deployment                                                                                                     | {}                |
| nodeSelector               | Node selector for the admission controller deployment                                                                                                                         | {}                |
| tolerations                | Tolerations for the admission controller deployment                                                                                                                           | {}                |
| affinity                   | Affinities and anti-affinities for the admission controller deployment                                                                                                        | {}                |
| webhook.failurePolicy      | Failure policy for the admission webhook. May be `Fail` or `Ignore`                                                                                                           | `Fail`            |
| webhook.crtPEM             | Certificate in PEM format for the admission controller webhook. Will be generated if not specified (Recommended)                                                              | -                 |
| webhook.keyPEM             | Private key for in PEM format for the admission controller webhook certificate. Will be generated if not specified (Recommended)                                              | -                 |
| webhook.caBundle           | Certificate Authority Bundle for the admission controller webhook. Will be generated if not specified (Recommended)                                                           | -                 |

## Optional: Test Installation

You can use helm to test that installation was successful and the admission controller is running using:

```console
helm test k8tz
```

## Upgrade

Upgrading k8tz does not have any effect on running pods, but only on pods created after the upgrade. 
In production environments, it is recommended to use `--atomic` flag to make the upgrade safer.

```console
helm repo update
helm upgrade k8tz k8tz/k8tz --reuse-values --atomic
```

## Uninstall

To uninstall k8tz with Helm use:

```console
helm delete k8tz
```

[^1]: Timezones for CronJobs are available only from kubernetes >=1.24.0-beta.0 with [`CronJobTimeZone`](https://github.com/kubernetes/enhancements/blob/aad71056d33eccf3845b73670106f06a9e74fec6/keps/sig-apps/3140-TimeZone-support-in-CronJob/README.md) feature gate enabled.
