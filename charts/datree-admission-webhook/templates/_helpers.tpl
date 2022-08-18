{{/* Create chart name and version as used by the chart label. */}}
{{- define "datree.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Helm required labels */}}
{{- define "datree.labels" -}}
app.kubernetes.io/name: datree-admission-webhook
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: "{{ .Chart.Version }}"
app.kubernetes.io/managed-by: "Helm"
app.kubernetes.io/part-of: "datree"
meta.helm.sh/release-name: "{{ .Chart.Name }}"
meta.helm.sh/release-namespace: "{{ .Release.Namespace }}" 
helm.sh/chart: {{ template "datree.chart" . }}
    {{- if .Values.customLabels }}
        {{ toYaml .Values.customLabels }}
    {{- end }}
{{- end -}}
