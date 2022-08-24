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
meta.helm.sh/release-namespace: "{{ .Release.Namespace |  }}" 
helm.sh/chart: {{ template "datree.chart" . }}
    {{- if .Values.customLabels -}}
        {{ toYaml .Values.customLabels }}
    {{- end -}}
{{- end -}}

{{/* The namespace name. */}}
{{- define "datree.namespace" -}}
{{- default .Release.Namespace .Values.namespace -}}
{{- end -}}
