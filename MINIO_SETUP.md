# MinIO Setup Guide

MinIO is now integrated into the Docker Compose setup for local file storage.

## Quick Start

1. **Start all services** (includes MinIO):
   ```bash
   docker-compose up -d
   ```

2. **Access MinIO Console**:
   - URL: http://localhost:9001
   - Username: `minioadmin`
   - Password: `minioadmin`

3. **Create the bucket**:
   - Login to console
   - Click "Buckets" → "Create Bucket"
   - Bucket name: `alieze-erp`
   - Click "Create Bucket"

4. **Verify it's working**:
   ```bash
   curl http://localhost:9000/minio/health/live
   # Should return: OK
   ```

## Configuration

MinIO is configured in:
- **Docker Compose**: `docker-compose.yml`
- **Environment**: `.env` file

### Default Settings

| Setting | Value |
|---------|-------|
| API Endpoint | http://localhost:9000 |
| Console | http://localhost:9001 |
| Access Key | minioadmin |
| Secret Key | minioadmin |
| Bucket Name | alieze-erp |

### Environment Variables

The app automatically connects to MinIO using these env vars:

```bash
STORAGE_PROVIDER=minio
STORAGE_S3_ENDPOINT=http://localhost:9000  # or http://minio:9000 from Docker
STORAGE_S3_BUCKET=alieze-erp
STORAGE_S3_ACCESS_KEY=minioadmin
STORAGE_S3_SECRET_KEY=minioadmin
STORAGE_S3_USE_SSL=false
STORAGE_S3_FORCE_PATH_STYLE=true
```

## Features

- ✅ S3-compatible API
- ✅ Web-based console
- ✅ Automatic bucket creation (manual first time)
- ✅ Persistent storage (Docker volume: `minio_data`)
- ✅ Health checks
- ✅ Works with existing S3 Go SDK code

## MinIO CLI (Optional)

Install MinIO client for advanced operations:

```bash
# macOS
brew install minio/stable/mc

# Configure alias
mc alias set local http://localhost:9000 minioadmin minioadmin

# Create bucket via CLI (alternative to console)
mc mb local/alieze-erp

# List buckets
mc ls local

# Upload test file
mc cp testfile.txt local/alieze-erp/
```

## Switching to AWS S3 (Production)

When ready for production, update `.env`:

```bash
STORAGE_PROVIDER=s3
STORAGE_S3_ENDPOINT=              # Empty for AWS
STORAGE_S3_BUCKET=your-s3-bucket
STORAGE_S3_ACCESS_KEY=your-aws-access-key
STORAGE_S3_SECRET_KEY=your-aws-secret-key
STORAGE_S3_USE_SSL=true
STORAGE_S3_FORCE_PATH_STYLE=false
```

No code changes needed - same S3 SDK works for both!

## Troubleshooting

### Bucket not found error
```bash
# Create bucket manually
mc mb local/alieze-erp
# or via console at http://localhost:9001
```

### Connection refused
```bash
# Check if MinIO is running
docker-compose ps minio

# Check logs
docker-compose logs minio

# Restart MinIO
docker-compose restart minio
```

### Permission denied
```bash
# Make bucket public (optional, for testing)
mc anonymous set download local/alieze-erp
```

## Data Persistence

MinIO data is stored in Docker volume `minio_data`:

```bash
# View volume info
docker volume inspect alieze-erp_minio_data

# Backup data
docker run --rm -v alieze-erp_minio_data:/data -v $(pwd):/backup alpine tar czf /backup/minio-backup.tar.gz /data

# Restore data
docker run --rm -v alieze-erp_minio_data:/data -v $(pwd):/backup alpine tar xzf /backup/minio-backup.tar.gz -C /
```

## Next Steps

Once MinIO is set up, the storage service is ready for:
- Document attachments
- Quote PDFs
- Email attachments
- Any file storage needs

The Go storage service (`pkg/storage`) automatically handles MinIO just like AWS S3!
