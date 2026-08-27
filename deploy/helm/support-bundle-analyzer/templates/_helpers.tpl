{{- define "sba.name" -}}support-bundle-analyzer{{- end }}
{{- define "sba.labels" -}}
app.kubernetes.io/name: {{ include "sba.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}
