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

# Collection with scalar fields (bool, int, string)
resource "milvus_collection" "products" {
  name                 = "products_collection"
  description          = "Product collection with scalar and vector fields"
  enable_dynamic_field = true
  shard_num            = 2
  consistency_level    = "Strong"

  fields = [
    {
      name           = "product_id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name       = "product_name"
      data_type  = "VarChar"
      max_length = 512
      nullable   = false
    },
    {
      name       = "description"
      data_type  = "VarChar"
      max_length = 512
      nullable   = true
    },
    {
      name      = "is_active"
      data_type = "Bool"
      nullable  = false
    },
    {
      name      = "in_stock"
      data_type = "Bool"
      nullable  = false
    },
    {
      name      = "category_id"
      data_type = "Int32"
      nullable  = true
    },
    {
      name      = "price"
      data_type = "Float"
      nullable  = true
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}

# BITMAP index on boolean field - good for filtering
resource "milvus_index" "is_active_bitmap" {
  collection_name = milvus_collection.products.name
  field_name      = "is_active"
  index_type      = "BITMAP"
  metric_type     = "L2" # Required by schema, but not used for scalar indexes
  index_name      = "is_active_idx"
}

# INVERTED index on boolean field - alternative to BITMAP
resource "milvus_index" "in_stock_inverted" {
  collection_name = milvus_collection.products.name
  field_name      = "in_stock"
  index_type      = "INVERTED"
  metric_type     = "L2" # Required by schema, but not used for scalar indexes
  index_name      = "in_stock_idx"
}

# INVERTED index on string field for full-text search
resource "milvus_index" "product_name_inverted" {
  collection_name = milvus_collection.products.name
  field_name      = "product_name"
  index_type      = "INVERTED"
  metric_type     = "L2" # Required by schema, but not used for scalar indexes
  index_name      = "product_name_idx"
}

# SORTED index on int field for range queries
resource "milvus_index" "category_id_sorted" {
  collection_name = milvus_collection.products.name
  field_name      = "category_id"
  index_type      = "SORTED"
  metric_type     = "L2" # Required by schema, but not used for scalar indexes
  index_name      = "category_id_idx"
}

# Vector index on embedding field for similarity search
resource "milvus_index" "embedding_hnsw" {
  collection_name = milvus_collection.products.name
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "COSINE"
  index_name      = "embedding_hnsw_idx"

  index_params = {
    m               = 8
    ef_construction = 200
  }
}
