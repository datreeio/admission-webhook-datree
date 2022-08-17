{{/* Create chart name and version as used by the chart label. */}}
{{- define "datree.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Helm required labels */}}
{{- define "datree.labels" -}}
app.kubernetes.io/version: "{{ .Chart.Version }}"
helm.sh/chart: {{ template "datree.chart" . }}
    {{- if .Values.customLabels }}
        {{ toYaml .Values.customLabels }}
    {{- end }}
{{- end -}}
