{{/*
Get KubeVersion removing pre-release information.
inspired by: https://github.com/prometheus-community/helm-charts/blob/a3b1ba697a0656e2403700e29b9e7d193d5caad3/charts/prometheus/templates/_helpers.tpl#L159
*/}}
{{- define "datree.kubeVersion" -}}
  {{- default .Capabilities.KubeVersion.Version (regexFind "v[0-9]+\\.[0-9]+\\.[0-9]+" .Capabilities.KubeVersion.Version) -}}
{{- end -}}


{{/*
Return the appropriate apiVersion for CronJob.
inspired by: https://github.com/prometheus-community/helm-charts/blob/a3b1ba697a0656e2403700e29b9e7d193d5caad3/charts/prometheus/templates/_helpers.tpl#L202
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

{{- define "datree.scanJob" -}}
spec:
  backoffLimit: 4
  ttlSecondsAfterFinished: {{.Values.scanJob.ttlSecondsAfterFinished | default "1" }}
  template:
    spec:
      {{- if .Values.nodeSelector }}
      nodeSelector: {{- toYaml .Values.nodeSelector | nindent 8 }}
      {{- end }}
      {{- with .Values.affinity }}
      affinity: {{ toYaml . | nindent 8 }}
      {{- end }}
      {{- if .Values.tolerations }}
      tolerations: {{- toYaml .Values.tolerations | nindent 8 }}
      {{- end }}
      serviceAccountName: cluster-scan-job-service-account
      restartPolicy: Never
      containers:
        - name: scan-job
          env:
            - name: DATREE_TOKEN
            {{- if and .Values.datree.existingSecret (ne .Values.datree.existingSecret.name "") (ne .Values.datree.existingSecret.key "") }}
              valueFrom:
                secretKeyRef:
                  name: {{ .Values.datree.existingSecret.name }}
                  key: {{ .Values.datree.existingSecret.key }}
            {{- else }}
              value: "{{ .Values.datree.token }}"
            {{- end }}
            - name: DATREE_POLICY
              value: {{.Values.datree.policy | default "Starter"}}
            - name: CLUSTER_NAME
              value: {{.Values.datree.clusterName}}
            - name: DATREE_NAMESPACE
              value: {{template "datree.namespace" .}}
          securityContext:
            {{- with .Values.securityContext }}
            {{ toYaml . | nindent 12 }}
            {{- end }}
            seccompProfile:
              type: RuntimeDefault
          image: "{{ .Values.scanJob.image.repository }}:{{ .Values.scanJob.image.tag }}"
          imagePullPolicy: IfNotPresent
          resources: {{- toYaml .Values.scanJob.resources | nindent 12 }}
          volumeMounts:
            - name: webhook-config
              mountPath: /config
              readOnly: true
            - name: custom-skip-list
              mountPath: /config/datreeSkipList
              readOnly: true
      volumes:
        - name: webhook-config
          configMap:
            name: webhook-scanning-filters
            optional: true
        - name: custom-skip-list
          configMap:
            name: custom-scanning-filters
            optional: true
{{- end -}}
