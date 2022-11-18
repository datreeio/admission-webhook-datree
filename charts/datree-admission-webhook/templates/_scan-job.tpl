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

{{- define "datree.scanJob" -}}
spec:
  backoffLimit: 4
  template:
    spec:
      serviceAccountName: cluster-scan-job-service-account
      restartPolicy: Never
      containers:
        - name: scan-job
          env:
            - name: DATREE_TOKEN
              value: {{.Values.datree.token}}
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            runAsNonRoot: true
            runAsUser: 25000
            seccompProfile:
              type: RuntimeDefault
          image: "datree/scan-job:0.0.8"
          imagePullPolicy: Always
          resources:
            limits:
              cpu: {{ .Values.scanJob.resources.limits.cpu }}
              memory: {{ .Values.scanJob.resources.limits.memory }}
            requests:
              cpu: {{ .Values.scanJob.resources.requests.cpu }}
              memory: {{ .Values.scanJob.resources.requests.memory }}
{{- end -}}
