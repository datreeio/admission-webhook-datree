

{{/* Expand the name of the chart. */}}
{{- define "datree.name" -}}
{{- default "webhook-server" .Values.nameOverride | trunc 63 }}
{{- end }}


{{/* Selector labels */}}
{{- define "datree.selectorLabels" -}}
app: {{ include "datree.name" . }}
{{- end }}

{{/* Labels */}}
{{- define "datree.labels" -}}
{{ include "datree.selectorLabels" . }}
{{- end }}

