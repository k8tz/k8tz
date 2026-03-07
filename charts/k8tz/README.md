# Kubernetes Timezone Controller - Helm Chart

![Lint Chart Workflow Status](https://img.shields.io/github/actions/workflow/status/k8tz/k8tz/helm-lint.yaml?branch=master&label=Lint)
![Build Chart Workflow Status](https://img.shields.io/github/actions/workflow/status/k8tz/k8tz/helm-release.yaml?branch=master&label=Release)
![Helm 3 Compatible](https://img.shields.io/badge/Helm%203-Compatible-blue)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/k8tz)](https://artifacthub.io/packages/helm/k8tz/k8tz)
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
- Helm 3
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

## Namespacing

k8tz ignores pods in the namespace where it is installed. Therefore it is recommended to install k8tz in a namespace that does not require the service at all. By default, k8tz creates its namespace called `k8tz` in which the admission controller is installed. The namespace can be renamed using the `namespace` value (`--set namespace=foo`). It is also possible to install k8tz in an existing namespace by canceling the creation of the namespace, using the value `createNamespace` (`--set namespace=existingNamespace,createNamespace=false`).

If you want to use the helm built-in namespace (`helm --namespace`), you can completely cancel the above behavior by setting the `namespace` value to null (`helm install k8tz k8tz/k8tz --namespace=foo --set namespace=null`).

## Values

| Parameter                            | Description                                                                                                                                                                   | Default           |
|--------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|
| kind                                 | Sets the workload type for the k8tz controller, which can be `Deployment` or `DaemonSet`                                                                                      | Deployment        |
| replicaCount                         | Amount of admission controller webhooks to spin up. For production use it is recommended to have at least 3 replicas. Only effective when `kind` is `Deployment`              | 1                 |
| namespace                            | The namespace where to install the admission controller. Set to `null` to use helm built-in namespace                                                                         | k8tz              |
| createNamespace                      | Whether the helm chart should create and manage the controller namespace. Only effective when the `namespace` is set from values instead of helm built-in namespace           | true              |
| timezone                             | The default timezone to inject                                                                                                                                                | UTC               |
| injectedInitContainerName            | The default name for injected initContainer                                                                                                                                   | k8tz              |
| injectedInitContainerImagePullPolicy | The default imagePullPolicy for injected initContainer                                                                                                                        | Always            |
| injectionStrategy                    | The default injection strategy to use                                                                                                                                         | initContainer     |
| injectAll                            | If true, timezone will be injected to the pod even when there is no annotation with explicit injection request. When false, the `k8tz.io/inject: true` annotation is required | true              |
| cronJobTimeZone                      | Enable injection of `timeZone` field to `CronJob`s[^1]                                                                                                                        | false             |
| verbose                              | Enable more detailed logs from admission controller and initContainers for debug purposes                                                                                     | false             |
| labels                               | Labels to apply to all resources                                                                                                                                              | {}                |
| image.repository                     | The image repository for the admission controller and bootstrap image                                                                                                         | quay.io/k8tz/k8tz |
| image.pullPolicy                     | Admission controller image pull policy                                                                                                                                        | IfNotPresent      |
| image.tag                            | The image tag for the admission controller and bootstrap image. The default is the chart appVersion                                                                           | -                 |
| imagePullSecrets                     | The image pull secrets for the admission controller                                                                                                                           | []                |
| nameOverride                         | Helm application name override                                                                                                                                                | -                 |
| fullnameOverride                     | Helm application full name override                                                                                                                                           | -                 |
| serviceAccount.annotations           | Annotations to add to the admission controller service account                                                                                                                | {}                |
| serviceAccount.name                  | The name of the service account to use. If empty, a name is generated using the fullname template                                                                             | -                 |
| podAnnotations                       | Annotations to add to the admission controller pod                                                                                                                            | {}                |
| podSecurityContext                   | Pod security context for the admission controller pod                                                                                                                         | {}                |
| securityContext                      | Security context for the admission controller pod                                                                                                                             | {}                |
| service.type                         | Admission controller service type                                                                                                                                             | ClusterIP         |
| service.port                         | Admission controller service port                                                                                                                                             | 443               |
| resources                            | Resource requests and limitations for the admission controller                                                                                                                | {}                |
| initContainerResources               | Resource requests and limitations for the bootstrap init container                                                                                                            | {}                |
| nodeSelector                         | Node selector for the admission controller                                                                                                                                    | {}                |
| tolerations                          | Tolerations for the admission controller                                                                                                                                      | {}                |
| topologySpreadConstraints            | TopologySpreadConstraints for the admission controller                                                                                                                        | []                |
| affinity                             | Affinities and anti-affinities for the admission controller                                                                                                                   | {}                |
| webhook.failurePolicy                | Failure policy for the admission webhook. May be `Fail` or `Ignore`                                                                                                           | `Fail`            |
| webhook.tlsMinVersion                | Minimum TLS version supported. Possible values: VersionTLS10, VersionTLS11, VersionTLS12, VersionTLS13, If omitted, the default VersionTLS12 will be used                     | -                 |
| webhook.tlsCipherSuites              | Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used                                                                   | -                 |
| webhook.certManager.enabled          | Use `cert-manager` to manage the webhook certificate by using `Certificate` resource                                                                                          | false             |
| webhook.certManager.secretTemplate   | Add custom labels and annotations to `Secret` that containing certificate generated by cert-manager[^2]                                                                       | {}                |
| webhook.certManager.duration         | The duration of the `Not After` date for the certificate generated by cert-manager[^2]                                                                                        | 2160h             |
| webhook.certManager.renewBefore      | The duration period before the certificate’s expiry when cert-manager should renew the certificate[^2]                                                                        | 720h              |
| webhook.certManager.issuerRef.name   | The name of cert-manager `Issuer` or `ClusterIssuer` used by the certificate[^2]                                                                                              | selfsigned        |
| webhook.certManager.issuerRef.kind   | The kind of cert-manager resource used by the certificate. May be `Issuer` or `ClusterIssuer`[^2]                                                                             | ClusterIssuer     |
| webhook.crtPEM                       | Certificate in PEM format for the admission controller webhook. Will be generated if not specified (Recommended)                                                              | -                 |
| webhook.keyPEM                       | Private key for in PEM format for the admission controller webhook certificate. Will be generated if not specified (Recommended)                                              | -                 |
| webhook.caBundle                     | Certificate Authority Bundle for the admission controller webhook. Will be generated if not specified (Recommended)                                                           | -                 |
| webhook.ignoredNamespaces            | List of namespaces that should be ignored by k8tz                                                                                                                             | ["kube-system"]   |

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
[^2]: Please refer to cert-manager documentation for using this feature.
