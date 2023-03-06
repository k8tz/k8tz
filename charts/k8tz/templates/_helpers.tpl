{{/*
Copyright © 2021 Yonatan Kahana

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/}}

{{/*
Expand the name of the chart.
*/}}
{{- define "k8tz.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "k8tz.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "k8tz.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "k8tz.labels" -}}
helm.sh/chart: {{ include "k8tz.chart" . }}
{{ include "k8tz.selectorLabels" . }}
{{- with .Values.labels }}
{{ toYaml . | trim }}
{{- end }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Pod labels
*/}}
{{- define "k8tz.podLabels" -}}
{{ include "k8tz.selectorLabels" . }}
{{- with .Values.labels }}
{{ toYaml . | trim }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "k8tz.selectorLabels" -}}
app.kubernetes.io/name: {{ include "k8tz.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "k8tz.serviceAccountName" -}}
{{- default (include "k8tz.fullname" .) .Values.serviceAccount.name }}
{{- end }}

{{/*
Defines the service name for the webhook
*/}}
{{- define "k8tz.serviceName" -}}
{{ .Release.Name }}
{{- end }}

{{/*
The default security context fields to be merged with the user defined settings.
The settings may differ between different versions of Kubernetes.
*/}}
{{- define "k8tz.defaultSecurityContext" -}}
allowPrivilegeEscalation: false
{{- if and (ge .Capabilities.KubeVersion.Major "1") (ge .Capabilities.KubeVersion.Minor "22") }}
capabilities:
  drop:
  - ALL
{{- end }}
runAsNonRoot: true
{{- if and (ge .Capabilities.KubeVersion.Major "1") (ge .Capabilities.KubeVersion.Minor "19") }}
seccompProfile:
  type: RuntimeDefault
{{- end }}
{{- end }}

{{/*
The default pod security context fields to be merged with the user defined settings.
The settings may differ between different versions of Kubernetes.
*/}}
{{- define "k8tz.defaultPodSecurityContext" -}}
runAsNonRoot: true
{{- if and (ge .Capabilities.KubeVersion.Major "1") (ge .Capabilities.KubeVersion.Minor "19") }}
seccompProfile:
  type: RuntimeDefault
{{- end }}
{{- end }}

{{/*
Merges the default security context settings with the user defined ones
*/}}
{{- define "k8tz.securityContext" -}}
{{- mergeOverwrite (include "k8tz.defaultSecurityContext" . | fromYaml) .Values.securityContext | toYaml }}
{{- end }}

{{/*
Merges the default pod security context settings with the user defined ones
*/}}
{{- define "k8tz.podSecurityContext" -}}
{{- mergeOverwrite (include "k8tz.defaultPodSecurityContext" . | fromYaml) .Values.podSecurityContext | toYaml }}
{{- end }}
