# Milvus Scalar and Boolean Field Indexes

This example demonstrates how to create and use indexes on scalar fields (boolean, integer, string) and vector fields in Milvus.

## Overview

While vector indexes are used for similarity search on embedding fields, scalar indexes are used for:
- Filtering and searching on metadata fields
- Boolean filtering
- String search
- Range queries on numeric fields

## Field Types in the Example

### Boolean Fields
- **is_active**: Indicates if a product is active
- **in_stock**: Indicates if a product is in stock

### String Fields
- **product_name**: The product name (VarChar - fixed max length)
- **description**: Product description (String - variable length)

### Numeric Fields
- **product_id**: Primary key (Int64)
- **category_id**: Category reference (Int32, nullable)
- **price**: Product price (Float, nullable)

### Vector Field
- **embedding**: Product embedding vector for similarity search (FloatVector, 768 dimensions)

## Index Types for Scalar Fields

### 1. BITMAP Index
**Best for:** Boolean fields, categorical data with few distinct values

```hcl
resource "milvus_index" "is_active_bitmap" {
  collection_name = milvus_collection.products.name
  field_name      = "is_active"
  index_type      = "BITMAP"
  metric_type     = "L2"  # Required but not used for scalar indexes
}
```

**Use cases:**
- Filtering products by active/inactive status
- Filtering by boolean flags
- Categorical data with low cardinality

### 2. INVERTED Index
**Best for:** Full-text search on string fields, flexible scalar indexing

```hcl
resource "milvus_index" "product_name_inverted" {
  collection_name = milvus_collection.products.name
  field_name      = "product_name"
  index_type      = "INVERTED"
  metric_type     = "L2"
}
```

**Use cases:**
- Full-text search on product names
- String matching
- Flexible scalar field indexing
- Works on boolean fields as well

### 3. SORTED Index
**Best for:** Range queries on numeric fields

```hcl
resource "milvus_index" "category_id_sorted" {
  collection_name = milvus_collection.products.name
  field_name      = "category_id"
  index_type      = "SORTED"
  metric_type     = "L2"
}
```

**Use cases:**
- Range queries: "Find products with category_id between 1 and 10"
- Ordering by numeric fields
- Numeric comparison operations

### 4. TRIE Index
**Best for:** String prefix searches

```hcl
resource "milvus_index" "product_name_trie" {
  collection_name = milvus_collection.products.name
  field_name      = "product_name"
  index_type      = "TRIE"
  metric_type     = "L2"
}
```

**Use cases:**
- Autocomplete features
- Prefix-based searches
- Efficient string matching with common prefixes

## Vector Index

For similarity search on the embedding field:

```hcl
resource "milvus_index" "embedding_hnsw" {
  collection_name = milvus_collection.products.name
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "COSINE"

  index_params = {
    m               = 8
    ef_construction = 200
  }
}
```

## Complete Workflow Example

### 1. Create Collection with Boolean Fields
```hcl
resource "milvus_collection" "products" {
  name = "products_collection"

  fields = [
    {
      name           = "product_id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name     = "is_active"
      data_type = "Bool"
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]
}
```

### 2. Create Indexes
```hcl
# Index for boolean filtering
resource "milvus_index" "is_active_idx" {
  collection_name = milvus_collection.products.name
  field_name      = "is_active"
  index_type      = "BITMAP"
  metric_type     = "L2"
}

# Vector index for similarity search
resource "milvus_index" "embedding_idx" {
  collection_name = milvus_collection.products.name
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "COSINE"

  index_params = {
    m               = 8
    ef_construction = 200
  }
}
```

### 3. Use in Queries
Once created, these indexes speed up:

**Boolean filtering:**
```
Find active products: WHERE is_active = true
```

**Vector similarity search:**
```
Find 10 most similar products to a given embedding
```

**Combined queries:**
```
Find 10 most similar active products:
  WHERE is_active = true
  ORDER BY similarity DESC
  LIMIT 10
```

## Important Notes

1. **Metric Type for Scalar Indexes**: The `metric_type` parameter is required by Terraform schema even for scalar indexes, but it's not actually used by Milvus for indexing scalar fields. Use "L2" as a placeholder.

2. **One Index Per Field**: Remember that Milvus only allows one index per field. If you need to change an index, you must delete and recreate it.

3. **Boolean Field Behavior**:
   - Boolean fields can use BITMAP (most efficient) or INVERTED indexes
   - BITMAP is optimized for boolean true/false values
   - INVERTED works but is more general-purpose

4. **Index Performance**:
   - BITMAP: Fastest for boolean filtering
   - INVERTED: Good for text search and flexible filtering
   - SORTED: Best for range queries
   - TRIE: Best for prefix searches

## Field Nullability

Some fields are nullable (optional):
```hcl
{
  name      = "category_id"
  data_type = "Int32"
  nullable  = true  # Can be NULL
}
```

Others are required:
```hcl
{
  name      = "is_active"
  data_type = "Bool"
  nullable  = false  # Cannot be NULL
}
```

## Resources

- [Milvus Index Documentation](https://milvus.io/docs/index.md)
- [Milvus Scalar Field Indexing](https://milvus.io/docs/scalar-index.md)
- [Milvus Boolean Field Guide](https://milvus.io/docs/bool.md)
