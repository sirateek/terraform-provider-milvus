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

# Create a collection first
resource "milvus_collection" "example" {
  name                 = "example_collection"
  description          = "Example collection for indexing"
  enable_dynamic_field = true
  shard_num            = 2
  consistency_level    = "Strong"

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name       = "title"
      data_type  = "VarChar"
      max_length = 256
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}

# Simple FLAT index for exact search
resource "milvus_index" "embedding_flat" {
  collection_name = milvus_collection.example.name
  field_name      = "embedding"
  index_type      = "FLAT"
  metric_type     = "L2"
}

# Example: IVF_FLAT index with custom nlist
# Note: Uncomment to replace FLAT index, but only one index allowed per field
# resource "milvus_index" "embedding_ivf_flat" {
#   collection_name = milvus_collection.example.name
#   field_name      = "embedding"
#   index_type      = "IVF_FLAT"
#   metric_type     = "COSINE"
#   index_name      = "embedding_ivf_flat_custom"
#
#   index_params = {
#     nlist = 128
#   }
# }

# Example: HNSW index for fast approximate search
# Note: Uncomment to replace FLAT index, but only one index allowed per field
# resource "milvus_index" "embedding_hnsw" {
#   collection_name = milvus_collection.example.name
#   field_name      = "embedding"
#   index_type      = "HNSW"
#   metric_type     = "IP"
#
#   index_params = {
#     m               = 8
#     ef_construction = 200
#   }
# }

# Example: IVF_PQ index with product quantization
# Note: Uncomment to replace FLAT index, but only one index allowed per field
# resource "milvus_index" "embedding_ivf_pq" {
#   collection_name = milvus_collection.example.name
#   field_name      = "embedding"
#   index_type      = "IVF_PQ"
#   metric_type     = "COSINE"
#
#   index_params = {
#     nlist = 128
#     m     = 8
#     nbits = 8
#   }
# }

# Example: DISKANN index for disk-based storage
# Note: Uncomment to replace FLAT index, but only one index allowed per field
# resource "milvus_index" "embedding_diskann" {
#   collection_name = milvus_collection.example.name
#   field_name      = "embedding"
#   index_type      = "DISKANN"
#   metric_type     = "L2"
# }
