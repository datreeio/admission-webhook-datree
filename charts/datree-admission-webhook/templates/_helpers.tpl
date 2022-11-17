{{/* Create chart name and version as used by the chart label. */}}
{{- define "datree.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}


{{/* Helm and Kubernetes required labels */}}
{{- define "datree.labels" -}}
app.kubernetes.io/name: {{.Chart.Name}}
app.kubernetes.io/managed-by: {{ .Release.Service | quote }}
app.kubernetes.io/instance: {{ .Release.Name | quote }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/part-of: "datree"
meta.helm.sh/release-name: "{{ .Chart.Name }}"
meta.helm.sh/release-namespace: "{{ .Release.Namespace}}" 
helm.sh/chart: {{ template "datree.chart" . }}
    {{- if .Values.customLabels -}}
        {{ toYaml .Values.customLabels }}
    {{- end -}}
{{- end -}}

{{/* The namespace name. */}}
{{- define "datree.namespace" -}}
{{- default .Release.Namespace .Values.namespace  -}}
{{- end -}}


{{/*
Get KubeVersion removing pre-release information.
*/}}
{{- define "datree.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version (regexFind "v[0-9]+\\.[0-9]+\\.[0-9]+" .Capabilities.KubeVersion.Version) -}}
{{- end -}}
{{/*
Return the appropriate apiVersion for CronJob.
for kubernetes version > 1.21.x use batch/v1
for 1.19.0 <= kubernetes version < 1.21.0 use batch/v1beta1 
*/}}
{{- define "datree.CronJob.apiVersion" -}}
  {{- if (semverCompare ">= 1.21.x" (include "datree.kubeVersion" .)) -}}
      {{- print "batch/v1" -}}
  {{- else -}}
    {{- print "batch/v1beta1" -}}
  {{- end -}}
{{- end -}}
