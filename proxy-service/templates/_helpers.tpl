{{/*
Create a default fully qualified app name.
*/}}
{{- define "proxy-service.fullname" -}}
{{- if .Values.fullnameOverride }}
{{.Values.fullnameOverride}}
{{- else }}
{{.Release.Name}}-{{.Chart.Name}}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "proxy-service.name" -}}
{{.Chart.Name}}
{{- end }}
