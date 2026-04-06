# Terraform Provider Milvus

A Terraform provider for managing Milvus collections and indexes.

> ⚠️This provider is in active development. Please report any issues or feature requests.
> Also, any contributions are welcome.

## Overview

The Terraform Provider for Milvus allows you to manage Milvus resources using Infrastructure as Code (IaC).
With this provider, you can:

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
