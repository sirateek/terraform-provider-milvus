# Copyright IBM Corp. 2026

terraform {
  required_providers {
    milvus = {
      source = "sirateek/milvus"
    }
  }
}

provider "milvus" {
  address    = "localhost:19530"
  username   = ""
  password   = ""
  db_name    = "default"
  enable_tls = false
}

# Simple collection with basic fields
resource "milvus_collection" "simple" {
  name        = "simple_collection"
  description = "A simple collection with basic fields"
  auto_id     = true

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
      is_auto_id     = true
    },
    {
      name       = "text"
      data_type  = "VarChar"
      max_length = 512
    },
    {
      name        = "vector_a"
      data_type   = "FloatVector"
      dim         = 1024
      description = "Test"
    }
  ]
}