![proteus banner](assets/banner.png)
<h3 align="center">High‑load async image processing service with idempotent atomic uploads, rollback support, and background cleanup of orphaned objects.</h3>

##



<br>

## Table of Contents

- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Shutting down](#shutting-down)
- [API](#api)
- [Validation](#validation)
- [Upload & rollback mechanics](#upload--rollback-mechanics)
- [Processing pipeline](#processing-pipeline)
- [Batch deletion & cleanup](#batch-deletion--cleanup)
- [Status values](#status-values)
- [Request examples](#request-examples)

<br>

## Architecture

- **App** — the central orchestrator of the system.  
  Responsible for application bootstrap and lifecycle management. It loads configuration, bootstraps both storages (PostgreSQL with Goose migrations + MinIO with bucket creation), wires all components (service, server, Kafka producer/consumer, meta & image storage) and manages graceful shutdown via a shared context.

- **Broker** — the messaging layer built on Kafka.  
  Publishes image processing tasks after successful upload and consumes them asynchronously. Uses a single topic with consumer groups. The producer and consumer are wrapped with logging, retries and graceful shutdown support.

- **Service** — the core business logic layer.  
  Validates uploads, performs atomic save to both storages, enqueues tasks to Kafka, processes images, handles rollbacks on failure, serves images and performs soft-delete + background cleanup.

- **MetaStorage** — PostgreSQL-backed metadata repository (source of truth).  
  Stores image UUID, object key, status and timestamps. Implements all status transitions, soft-delete with update locking and batch cleanup queries.

- **ImageStorage** — MinIO-backed object storage.  
  Persists original and processed images. Supports upload, download, single delete and batch delete (including temporary "unprocessed" variants).

- **Handler** — HTTP layer (Gin-based).  
  Exposes REST API under /api/v1, serves a simple web UI at root and enforces request/file size limits.

- **Cleaner** — background maintenance goroutine.  
  Periodically removes soft-deleted and stale pending images from both storages using batch operations.


<br>

## Installation
⚠️ Note: This project requires Docker Compose, regardless of how you choose to run it. 

First, clone the repository and enter the project folder:

```bash
git clone https://github.com/Pur1st2EpicONE/Proteus.git
cd Proteus
```

Then you have two options:

#### 1. Run everything in containers
```bash
make
```

This will start the entire project fully containerized using Docker Compose.

#### 2. Run Proteus locally
```bash
make local
```

In this mode, only PostgreSQL, MinIO and Kafka are started in containers, while the Go application runs locally.

⚠️ Note:
Local mode requires Go 1.25.1 installed on your machine.

<br>

## Configuration

### Runtime configuration

Proteus uses two configuration files, depending on the selected run mode:

[config.full.yaml](./configs/config.full.yaml) — used for the fully containerized setup

[config.dev.yaml](./configs/config.dev.yaml) — used for local development

You may optionally review and adjust the corresponding configuration file to match your preferences. The default values are suitable for most use cases.

<br>

## Shutting down

Stopping Proteus depends on how it was started:

- Local setup — press Ctrl+C to send SIGINT to the application. The service will gracefully close connections and finish any in-progress operations.  
- Full Docker setup — containers run by Docker Compose will be stopped automatically.

In both cases, to stop all services and clean up containers, run:

```bash
make down
```

⚠️ Note: In the full Docker setup, the log folder is created by the container as root and will not be removed automatically. To delete it manually, run:
```bash
sudo rm -rf <log-folder>
```

⚠️ Note: Docker Compose also creates persistent volumes for data storage (e.g., postgres_data, minio-data). These volumes are not removed automatically when containers are stopped. To remove them and fully reset the environment, run:
```bash
make reset
```

<br>

## API

All endpoints are mounted under /api/v1. Responses follow a simple wrapper convention:

- **Success**: 200 OK or 202 Accepted with JSON body {"result": <value>}
- **Error**: appropriate status code with JSON body {"error": "<message>"}

<br>

### Upload image

```bash
POST /api/v1/upload
```

Multipart form:

- image (file, required)
- action (string, required) — one of: resize, thumbnail, watermark
- watermark (string) — required for watermark action
- width / height (int) — used for resize/thumbnail

On success returns the image UUID.

### Get image

```bash
GET /api/v1/image/:id
```

- If still processing → 202 Accepted with status pending
- If ready → 200 OK with the processed image (correct Content-Type)
- If deleted or not found → 404 Not Found

### Mark as deleted (soft delete)

```bash
DELETE /api/v1/image/:id
```

Changes status to **deleted**. The image will be permanently removed by the background cleaner.

<br>

## Validation

All validation happens in the service layer:

- Supported image formats: JPEG, PNG, WEBP, GIF
- Maximum dimensions: 12000×12000 px
- File and request size limits are taken from config
- Action-specific rules (watermark text required, resize needs at least one positive dimension, etc.)

<br>

## Upload & rollback mechanics

Upload is atomic:

1. Validate request and image
2. Generate UUID + temporary object key
3. Save metadata to PostgreSQL and upload original to MinIO in parallel
4. If any step fails → rollback:
   - If meta save failed → delete image from MinIO
   - If MinIO upload failed → meta is not saved (no rollback needed)
5. Only after both storages succeed → marshal task and send to Kafka
6. If Kafka produce fails → delete image from MinIO (orphan cleanup)

This guarantees no orphaned images and no lost metadata.

<br>

## Processing pipeline

After successful upload:

- Task is published to Kafka
- Consumer receives ImageProcessTask
- Downloads original from MinIO
- Applies the requested transformation (resize / thumbnail / watermark) using the imaging library
- Uploads processed image to MinIO
- Updates meta record → status = ready + final object_key

All operations are idempotent and safe to retry.

<br>



## Batch deletion & cleanup

- Soft delete (DELETE /image/:id) only changes status to deleted
- Background Cleaner goroutine runs every **[cleanup_interval](https://github.com/Pur1st2EpicONE/Proteus/blob/68e926e6b64cbb7d2720acdf1a39c25ae07d921b/configs/config.full.yaml#L21)**
- It selects images where status = deleted or status = pending longer than **[pending_timeout](https://github.com/Pur1st2EpicONE/Proteus/blob/68e926e6b64cbb7d2720acdf1a39c25ae07d921b/configs/config.full.yaml#L34)**

- Performs batch delete:
  1. imageStorage.DeleteBatch() — removes both original and processed variants from MinIO
  2. metaStorage.DeleteBatch() — permanently removes records from PostgreSQL

This approach keeps the database and object storage clean without blocking user requests.

<br>

## Status values
- **pending** — image uploaded and queued for processing (or currently being processed)
- **ready** — processing finished successfully, image is available for download
- **deleted** — soft-deleted by user (will be cleaned up by background job)

<br>

## Request examples

⚠️ Note: When the service is running, a web-based UI is available at http://localhost:8080. The examples below demonstrate how to interact with the API directly using curl.


### Upload image (resize)

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "image=@photo.jpg" \
  -F "action=resize" \
  -F "width=800" \
  -F "height=600"
```

### Upload image (watermark)

```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -F "image=@photo.jpg" \
  -F "action=watermark" \
  -F "watermark=© MyBrand"
```

### Get image
```bash
curl -O -J "http://localhost:8080/api/v1/image/123e4567-e89b-12d3-a456-426614174000"
```

### Delete image
```bash
curl -X DELETE "http://localhost:8080/api/v1/image/123e4567-e89b-12d3-a456-426614174000"
```
