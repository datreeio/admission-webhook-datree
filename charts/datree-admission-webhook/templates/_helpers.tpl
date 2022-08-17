{{/* Helm required labels */}}
{{- define "datree.labels" -}}
    {{- if .Values.customLabels }}
        {{ toYaml .Values.customLabels }}
    {{- end }}
{{- end -}}
