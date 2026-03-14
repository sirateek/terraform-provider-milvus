# Milvus Index Resource Examples

This directory contains examples for creating and managing Milvus indexes using Terraform.

## Overview

Milvus indexes optimize search performance by organizing data for faster retrieval. Different index types are suitable for different use cases and performance requirements.

## Index Types Supported

### Vector Indexes (for FloatVector, BinaryVector, etc.)

| Index Type | Best For | Parameters | Notes |
|---|---|---|---|
| **FLAT** | Exact search, small datasets | - | No compression, guarantees exact results |
| **IVF_FLAT** | General purpose | `nlist` | Inverted file with clustering |
| **IVF_SQ8** | Memory efficiency | `nlist` | Scalar quantization (4x memory reduction) |
| **IVF_PQ** | Extreme memory constraint | `nlist`, `m`, `nbits` | Product quantization |
| **HNSW** | Fast approximate search | `m`, `ef_construction` | Hierarchical graph-based |
| **DISKANN** | Very large datasets | - | Disk-based, saves memory |
| **SCANN** | Balanced performance | `nlist`, `with_raw_data` | Efficient quantization |
| **AUTOINDEX** | Let Milvus decide | - | Automatic index selection |

### Scalar Indexes (for String, Int64, etc.)

| Index Type | Best For | Parameters |
|---|---|---|
| **TRIE** | String prefix search | - |
| **SORTED** | Range queries | - |
| **INVERTED** | Full-text search | - |
| **BITMAP** | Boolean/categorical data | - |

### Sparse Indexes

| Index Type | Best For | Parameters |
|---|---|---|
| **SPARSE_INVERTED** | Sparse vectors | `drop_ratio` |
| **SPARSE_WAND** | Fast sparse search | `drop_ratio` |

## Metric Types

For vector indexes, you must specify a metric type:

- **L2** - Euclidean distance (best for general purpose)
- **COSINE** - Cosine similarity (for normalized embeddings)
- **IP** - Inner product (for dense vectors)
- **HAMMING** - Hamming distance (for binary vectors)
- **JACCARD** - Jaccard similarity (for binary vectors)

## Usage Examples

### Simple FLAT Index
```hcl
resource "milvus_index" "simple" {
  collection_name = "my_collection"
  field_name      = "embedding"
  index_type      = "FLAT"
  metric_type     = "L2"
}
```

### IVF_FLAT with Custom Parameters
```hcl
resource "milvus_index" "ivf" {
  collection_name = "my_collection"
  field_name      = "embedding"
  index_type      = "IVF_FLAT"
  metric_type     = "COSINE"

  index_params = {
    nlist = 128  # Number of clusters
  }
}
```

### HNSW Index (Recommended for most use cases)
```hcl
resource "milvus_index" "hnsw" {
  collection_name = "my_collection"
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "IP"

  index_params = {
    m               = 16              # Max connections per node (5-48, default 16)
    ef_construction = 200             # Construction search width (topk-4096)
  }
}
```

### IVF_PQ Index (Maximum compression)
```hcl
resource "milvus_index" "ivf_pq" {
  collection_name = "my_collection"
  field_name      = "embedding"
  index_type      = "IVF_PQ"
  metric_type     = "L2"

  index_params = {
    nlist = 128  # Number of clusters
    m     = 8    # Number of subquantizers
    nbits = 8    # Bits per subvector
  }
}
```

### Custom Index Name
```hcl
resource "milvus_index" "named" {
  collection_name = "my_collection"
  field_name      = "embedding"
  index_name      = "my_custom_index_name"
  index_type      = "FLAT"
  metric_type     = "COSINE"
}
```

## Parameter Recommendations

### For IVF Indexes (IVF_FLAT, IVF_SQ8, IVF_PQ)
- **Small datasets (< 1M vectors)**: `nlist = 64-256`
- **Medium datasets (1M-10M)**: `nlist = 256-1024`
- **Large datasets (> 10M)**: `nlist = 1024-4096`

### For HNSW Index
- **General use**: `m = 16, ef_construction = 200`
- **High recall needed**: `m = 24, ef_construction = 300`
- **Fast indexing**: `m = 8, ef_construction = 100`

## Index Creation Time

- **FLAT**: Instant (no indexing needed)
- **IVF_FLAT/SQ8/PQ**: 1-10 minutes for 1M vectors
- **HNSW**: 5-30 minutes for 1M vectors
- **DISKANN**: 10-60 minutes for 1M vectors

## Notes

1. **Indexes are immutable**: To change an index, delete it and create a new one
2. **Index creation is asynchronous**: Terraform waits for the operation to complete
3. **Multiple indexes**: You can create multiple indexes on the same field
4. **Field type matters**: Choose indexes appropriate for your field data type
5. **Memory overhead**: Some indexes (HNSW, DISKANN) use additional memory

## Troubleshooting

- **"Index type not supported"**: Ensure the index type is one of the supported types listed above
- **Invalid metric type**: Verify metric type (L2, COSINE, IP, HAMMING, JACCARD)
- **Parameter validation**: Check that parameters are within valid ranges for the index type
- **Collection not found**: Ensure the collection exists before creating an index on it
- **Field not found**: Verify the field name exists in the collection

## Reference

For more information, see:
- [Milvus Index Documentation](https://milvus.io/docs/index.md)
- [Milvus Go Client Reference](https://pkg.go.dev/github.com/milvus-io/milvus/client/v2)
