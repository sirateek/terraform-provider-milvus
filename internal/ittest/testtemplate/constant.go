// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package testtemplate

const Template = `
provider "milvus" {
  address = "{{ .MilvusAddress }}"
}
{{ range .Collections }}
resource "milvus_collection" "{{ .TerraformResourceName }}" {
  name              = "{{ .Name }}"
  description       = "{{ .Description }}"
  auto_id           = {{ .AutoID }}
  delete_protection = {{ .DeleteProtection }}
  shard_num         = {{ .ShardNum }}

  fields = [
    {{- range $i, $f := .Fields }}
    {
      name      = "{{ $f.Name }}"
      data_type = "{{ $f.DataType }}"
      {{- if $f.IsPrimaryKey }}
      is_primary_key = {{ deref $f.IsPrimaryKey }}
      {{- end }}
      {{- if $f.Dim }}
      dim = {{ deref $f.Dim }}
      {{- end }}
      {{- if $f.MaxLength }}
      max_length = {{ deref $f.MaxLength }}
      {{- end }}
      {{- if $f.MaxCapacity }}
      max_capacity = {{ deref $f.MaxCapacity }}
      {{- end }}
      {{- if $f.ElementType }}
      element_type = "{{ deref $f.ElementType }}"
      {{- end }}
    },
    {{- end }}
  ]
}
{{ end }}
{{- range .Indexes }}
resource "milvus_index" "{{ .TerraformResourceName }}" {
  collection_name = {{ .CollectionName }}
  field_name      = "{{ .FieldName }}"
  index_name      = "{{ .IndexName }}"
  index_type      = "{{ .IndexType }}"
  metric_type     = "{{ .MetricType }}"
  {{- if .IndexParams }}

  index_params = {
    {{- range $k, $v := .IndexParams }}
    {{ $k }} = {{ $v }}
    {{- end }}
  }
  {{- end }}
}
{{ end }}`
