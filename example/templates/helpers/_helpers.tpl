{{- define "example.fullname" -}}
{{ printf "%s-app" .Release.Name }}
{{- end -}}
