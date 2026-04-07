// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package testtemplate

import (
	"bytes"
	"os"
	"text/template"
)

type TerraformTemplate struct {
	MilvusAddress string
	Collections   []CollectionTemplate
	Indexes       []IndexTemplate
	Aliases       []AliasTemplate
}

type AliasTemplate struct {
	TerraformResourceName string
	AliasName             string
	CollectionName        string
}

type IndexTemplate struct {
	TerraformResourceName string
	CollectionName        string
	FieldName             string
	IndexName             string
	IndexType             string
	MetricType            string
	IndexParams           map[string]any
}

type CollectionTemplate struct {
	Name                  string
	TerraformResourceName string
	Description           string
	AutoID                bool
	DeleteProtection      bool
	ShardNum              int
	ConsistencyLevel      *string
	Properties            *CollectionPropertiesTemplate
	Fields                []FieldTemplate
}

type CollectionPropertiesTemplate struct {
	TTLSeconds            *int64
	MmapEnabled           *bool
	PartitionKeyIsolation *bool
	DynamicFieldEnabled   *bool
	AllowInsertAutoID     *bool
	AllowUpdateAutoID     *bool
	Timezone              *string
}

type FieldTemplate struct {
	Name         string
	DataType     string
	IsPrimaryKey *bool
	Nullable     *bool
	Dim          *int
	MaxLength    *int
	MaxCapacity  *int
	ElementType  *string
}

// Render executes the collection HCL template with the given data.
func (c TerraformTemplate) Render() string {
	funcMap := template.FuncMap{
		"deref": func(v interface{}) interface{} {
			switch p := v.(type) {
			case *bool:
				return *p
			case *int:
				return *p
			case *int64:
				return *p
			case *string:
				return *p
			}
			return v
		},
	}

	address := os.Getenv("MILVUS_ADDRESS")
	if c.MilvusAddress == "" {
		c.MilvusAddress = address

		// Default Value for MilvusAddress
		if address == "" {
			c.MilvusAddress = "localhost:19530"
		}
	}

	tmpl := template.Must(template.New("collection").Funcs(funcMap).Parse(Template))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		panic("testtemplate: failed to render collection template: " + err.Error())
	}
	return buf.String()
}

// BoolPtr is a helper to get a pointer to a bool literal.
func BoolPtr(v bool) *bool { return &v }

// IntPtr is a helper to get a pointer to an int literal.
func IntPtr(v int) *int { return &v }

// Int64Ptr is a helper to get a pointer to an int64 literal.
func Int64Ptr(v int64) *int64 { return &v }

// StringPtr is a helper to get a pointer to a string literal.
func StringPtr(v string) *string { return &v }
