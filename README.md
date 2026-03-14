# Terraform Provider for Milvus

A Terraform provider for managing Milvus collections and indexes.

## Overview

The Terraform Provider for Milvus allows you to manage Milvus resources using Infrastructure as Code (IaC). With this provider, you can:

- Create and manage Milvus collections with custom schemas
- Create and manage indexes on collection fields
- Define collection properties and configurations
- Support both vector and scalar field indexing

## Features

✅ **Collection Management**
- Create collections with custom field schemas
- Define field types (Int64, VarChar, FloatVector, BinaryVector, etc.)
- Configure collection properties (TTL, consistency level, dynamic fields, etc.)
- Support for auto-ID generation

✅ **Index Management**
- Vector indexes: FLAT, IVF_FLAT, IVF_SQ8, IVF_PQ, HNSW, DISKANN, SCANN, AUTOINDEX
- Scalar indexes: BITMAP, INVERTED, SORTED, TRIE
- Sparse indexes: SPARSE_INVERTED, SPARSE_WAND
- Index parameter configuration per index type

✅ **Collection Properties**
- Control TTL (Time-To-Live)
- Enable/disable dynamic fields
- Configure consistency levels
- Set shard numbers
- Enable memory-mapped (mmap) storage
- Control auto-ID insertion/update

## Requirements

- Terraform >= 1.0
- Go >= 1.21 (for building from source)
- Milvus >= 2.0

## Installation

### Using Terraform Registry

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    milvus = {
      source  = "sirateek/milvus"
      version = "~> 1.0"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/sirateek/terraform-provider-milvus.git
cd terraform-provider-milvus
make install
```

## Provider Configuration

### Basic Configuration

```hcl
provider "milvus" {
  address  = "localhost:19530"
  username = "default"
  password = "password"
  db_name  = "default"
}
```

### Configuration with TLS

```hcl
provider "milvus" {
  address    = "milvus.example.com:19530"
  username   = "admin"
  password   = "secure_password"
  db_name    = "production"
  enable_tls = true
}
```

### Configuration with Environment Variables

```bash
export MILVUS_ADDRESS="localhost:19530"
export MILVUS_USERNAME="default"
export MILVUS_PASSWORD="password"
export MILVUS_DB_NAME="default"
```

### Provider Arguments

| Name | Description | Type | Required | Default |
|------|-------------|------|----------|---------|
| `address` | Milvus server address (host:port) | string | Yes | - |
| `username` | Username for authentication | string | No | - |
| `password` | Password for authentication | string | No | - |
| `db_name` | Database name to manage | string | No | "default" |
| `enable_tls` | Enable TLS for connection | bool | No | false |
| `api_key` | API key for authentication | string | No | - |

## Resources

### milvus_collection

Create and manage Milvus collections.

#### Example

```hcl
resource "milvus_collection" "example" {
  name                 = "my_collection"
  description          = "Example collection"
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
      name       = "text"
      data_type  = "VarChar"
      max_length = 512
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 768
    }
  ]

  properties = {
    mmap_enabled              = true
    collection_ttl_seconds    = 3600
    allow_insert_auto_id      = true
    partition_key_isolation   = false
    dynamic_field_enabled     = true
  }
}
```

### milvus_index

Create and manage indexes on collection fields.

#### Vector Index Example

```hcl
resource "milvus_index" "embedding" {
  collection_name = milvus_collection.example.name
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "COSINE"
  index_name      = "embedding_idx"

  index_params = {
    m               = 8
    ef_construction = 200
  }
}
```

#### Scalar Index Example

```hcl
resource "milvus_index" "text_idx" {
  collection_name = milvus_collection.example.name
  field_name      = "text"
  index_type      = "INVERTED"
  metric_type     = "L2"
  index_name      = "text_search_idx"
}
```

## Supported Field Types

### Numeric
- Int8, Int16, Int32, Int64
- Float, Double

### String
- VarChar (variable length, max 65535 chars)
- String (long text)

### Boolean
- Bool

### Vector
- FloatVector (32-bit floats)
- BinaryVector (binary data)
- Float16Vector, BFloat16Vector

### Complex
- Array, JSON

## Index Types

### Vector Indexes
- **FLAT** - Exact search
- **IVF_FLAT** - Inverted File clustering
- **IVF_SQ8** - IVF with scalar quantization
- **IVF_PQ** - IVF with product quantization
- **HNSW** - Hierarchical Navigable Small World (recommended)
- **DISKANN** - Disk-based index
- **SCANN** - Scalable Clustered ANN
- **AUTOINDEX** - Auto-selected index

### Scalar Indexes
- **BITMAP** - For boolean/categorical data
- **INVERTED** - For full-text search
- **SORTED** - For range queries
- **TRIE** - For prefix searches

### Sparse Indexes
- **SPARSE_INVERTED**
- **SPARSE_WAND**

## Complete Example

See `examples/` directory for:
- [Vector Index Example](./examples/index/) - Creating indexes on vector fields
- [Scalar Index Example](./examples/scalar-index/) - Boolean fields and scalar indexes
- [Collection Example](./examples/collection/) - Creating collections

## Consistency Levels

- **Strong** - All reads from latest committed version
- **Bounded** - Reads from version no older than specified time
- **Session** - Reads from latest version at write time
- **Eventually** - Reads from any available version

## Quick Start

```hcl
# Create a simple collection
resource "milvus_collection" "quickstart" {
  name = "quickstart"

  fields = [
    {
      name           = "id"
      data_type      = "Int64"
      is_primary_key = true
    },
    {
      name      = "embedding"
      data_type = "FloatVector"
      dim       = 384
    }
  ]
}

# Add index for similarity search
resource "milvus_index" "quickstart_idx" {
  collection_name = milvus_collection.quickstart.name
  field_name      = "embedding"
  index_type      = "HNSW"
  metric_type     = "COSINE"
}
```

## Troubleshooting

### Connection Issues
- Verify Milvus server is running and accessible
- Check address and port are correct
- Verify credentials if authentication is enabled

### Index Creation
- Only one index per field allowed
- Vector indexes require vector fields
- Scalar indexes for boolean, string, numeric fields

### Common Errors
- "at most one distinct index is allowed per field" - Delete existing index first
- "data type cannot build with this index type" - Wrong index type for field type

## Contributing

Contributions welcome! See [CONTRIBUTING.md](./CONTRIBUTING.md)

### Development

```bash
make build    # Build provider
make test     # Run tests
make testacc  # Run acceptance tests
make generate # Generate docs
```

## License

Mozilla Public License 2.0 (MPL-2.0) - See LICENSE file

## Support

- Documentation: [Milvus Docs](https://milvus.io/docs)
- Issues: [GitHub Issues](https://github.com/sirateek/terraform-provider-milvus/issues)
- Examples: See `examples/` directory

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
