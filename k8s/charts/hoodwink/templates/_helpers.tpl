{{/*
Nome base. Mantém os nomes curtos e previsíveis (hoodwink-backend etc.),
que são contrato com o nginx do host em k8s/scripts/edge-nginx.sh.
*/}}
{{- define "hoodwink.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "hoodwink.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "hoodwink.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/name: {{ include "hoodwink.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "hoodwink.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hoodwink.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}


{{- define "hoodwink.secretName" -}}
{{- required "secrets.existingSecret é obrigatório" .Values.secrets.existingSecret -}}
{{- end -}}
