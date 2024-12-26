{{/*
Expand the name of the chart.
*/}}
{{- define "bk-nodeman.name" -}}
{{- include "common.names.name" . -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "bk-nodeman.fullname" -}}
{{- include "common.names.fullname" . -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "bk-nodeman.chart" -}}
{{- include "common.names.chart" . -}}
{{- end -}}

{{/*
Return the proper bk-nodeman image name
*/}}
{{- define "bk-nodeman.image" -}}
{{ include "common.images.image" (dict "imageRoot" .Values.image "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper bk-nodeman image registry secret names
*/}}
{{- define "bk-nodeman.imagePullSecrets" -}}
{{ include "common.images.pullSecrets" (dict "images" (list .Values.image) "global" .Values.global) }}
{{- end -}}

{{/*
Return the proper bk-nodeman replica count
{{ include "bk-nodeman.replicaCount" ( dict "module" .Values.path.to.module ) }}
*/}}
{{- define "bk-nodeman.replicaCount" -}}
{{- if gt .root.replicaCount 1.0 }}
{{- .root.replicaCount -}}
{{- else -}}
{{- .module.replicaCount -}}
{{- end -}}
{{- end -}}