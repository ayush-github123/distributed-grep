# distGrep - Distributed Grep Using MapReduce

A distributed grep tool built in Go that leverages the MapReduce pattern to efficiently search for patterns across multiple files in a distributed system.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Key Components](#key-components)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [Future Enhancements](#future-enhancements)

## Overview

**distGrep** is a distributed pattern matching system that distributes grep operations across multiple worker nodes. It implements the classic MapReduce algorithm to parallelize searching for patterns in large log files or text datasets.

### Key Features

- **Distributed Processing**: Distributes pattern matching tasks across multiple worker nodes
- **MapReduce Pattern**: Uses the Map (pattern matching) and Reduce (result aggregation) paradigm
- **Scalable Architecture**: Easy to add more workers to handle larger datasets
- **HTTP-based Communication**: Master-Worker communication via REST APIs
- **Directory Support**: Can recursively process directories
- **Regex Pattern Matching**: Uses Go's regexp package for powerful pattern matching

## Architecture

The system follows a **Master-Worker** architecture:

```
          ┌──────────── Map() ────────────┐
File --> Split chunks --> key-value pairs --> Group by key --> Reduce() --> Final result
         
Client / curl
     ↓
MASTER  (POST /run)
  ↓
  Schedules tasks to Workers
  ↓
Workers /map
  ↓
Collects intermediate results
  ↓
Reduce Phase
  ↓
Final result to Client
```

## Project Structure

```
distGrep/
├── cmd/
│   ├── main.go                 # Master server entry point
│   ├── master/
│   │   ├── main.go             # Master startup script
│   │   └── master.go           # Master request handlers
│   └── worker/
│       └── worker.go           # Worker server implementation
├── internals/
│   ├── grep/
│   │   └── grep.go             # Grep mapper and reducer logic
│   ├── mapreduce/
│   │   ├── mapreduce.go        # MapReduce job orchestration
│   │   └── worker_client.go    # Worker client calls
│   └── types/
│       └── types.go            # Type definitions
├── test-logs/                  # Test data and logs
├── go.mod                      # Go module file
├── Readme.md                   # This file
└── notes.md                    # Development notes
```

## Key Components

### 1. Master Node (`cmd/master/master.go`)

The master node coordinates the entire MapReduce job:

- **HandleRegister**: Registers new worker nodes
- **HandleRunJob**: Accepts grep requests and orchestrates the job execution
- **GetWorkers**: Manages the list of available workers
- Handles directory traversal for multi-file searches

### 2. Worker Nodes (`cmd/worker/worker.go`)

Worker nodes perform the actual grep operation:

- **HandleMap**: Processes a single file and returns matching lines
- **RegisterWorker**: Registers itself with the master
- Executes pattern matching on assigned file chunks

### 3. MapReduce Engine (`internals/mapreduce/mapreduce.go`)

Core MapReduce implementation:

- **Job**: Represents a MapReduce job with pattern, reducer, tasks, and workers
- **MapPhase**: Executes the mapping phase across workers
- **Reduce Phase**: Aggregates results from all mappers
- Handles synchronization and error management

### 4. Grep Logic (`internals/grep/grep.go`)

Contains the actual grep implementation:

- **GrepMapper**: Maps files to matching key-value pairs
- **GrepReducer**: Reduces values for each matched line
- Key format: `filename:linenum`, Value: matched line content

### 5. Type Definitions (`internals/types/types.go`)

Defines data structures for communication:

- **KeyValue**: Key-value pair emitted by mappers
- **MapRequest/MapResponse**: Communication protocol for map operations
- **RunRequest/RunResponse**: Communication protocol for job requests

## Getting Started

### Prerequisites

- Go 1.25.4 or higher
- Basic understanding of MapReduce and grep

### Build

```bash
# Navigate to the project directory
cd distGrep

# Build the master
go build -o master ./cmd/master

# Build the worker
go build -o worker ./cmd/worker
```

### Installation

```bash
go mod download
go mod tidy
```

## Usage

### Step 1: Start the Master

```bash
./master
```

The master will start listening on port `9000`.

### Step 2: Start Worker Nodes

In separate terminals:

```bash
# Worker 1
./worker --master=http://localhost:9000 --port=8001

# Worker 2
./worker --master=http://localhost:9000 --port=8002

# Worker 3
./worker --master=http://localhost:9000 --port=8003
```

### Step 3: Submit a Grep Job

Use curl to submit a grep request:

```bash
curl -X POST http://localhost:9000/run \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "error|ERROR",
    "files": ["./test-logs/test1.log", "./test-logs/test2.log"]
  }'
```

### Example with Directory

```bash
curl -X POST http://localhost:9000/run \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "connection timeout",
    "files": ["./logs/"]
  }'
```

## API Endpoints

### Master Endpoints

#### 1. Register Worker

**POST** `/register`

Registers a worker node with the master.

**Request Body:**
```json
{
  "address": "http://localhost:8001"
}
```

**Response:** 200 OK

#### 2. Run Grep Job

**POST** `/run`

Submits a grep pattern matching job.

**Request Body:**
```json
{
  "pattern": "error|warning|failed",
  "files": ["path/to/file1.log", "path/to/dir/"]
}
```

**Response:**
```json
{
  "results": {
    "path/to/file1.log:42": "ERROR: Database connection failed",
    "path/to/file2.log:156": "WARNING: High memory usage detected"
  }
}
```

### Worker Endpoints

#### 1. Map Endpoint

**POST** `/map`

Performs the grep operation on a single file.

**Request Body:**
```json
{
  "pattern": "error",
  "path": "/path/to/file.log"
}
```

**Response:**
```json
{
  "kvs": [
    {
      "key": "/path/to/file.log:42",
      "value": "ERROR: Something went wrong"
    }
  ]
}
```

## How It Works

### Map Phase

1. Master receives a grep request with a pattern and list of files/directories
2. If input is a directory, it recursively collects all files
3. Master creates tasks (one per file) and distributes them to available workers
4. Each worker scans its assigned file line-by-line
5. For each line matching the regex pattern, the worker emits a key-value pair:
   - **Key**: `filename:linenumber`
   - **Value**: The matched line text

### Intermediate Grouping

1. Master collects all key-value pairs from workers
2. Groups values by key (all matches for the same line)
3. This is done in the `intermediates` map

### Reduce Phase

1. For each unique key, the reducer processes all associated values
2. Current implementation returns the first matching value (simple reducer)
3. Can be extended for more complex aggregations

### Result Compilation

1. Final results are compiled into a map: `key -> result`
2. Returned to the client as JSON response

## Configuration

### Master Configuration

The master reads the following from the environment or command-line flags:

- **Port**: Default is `9000`
- **Log File**: `distgrep.log`

### Worker Configuration

Workers accept the following flags:

- `--master`: Master server URL (e.g., `http://localhost:9000`)
- `--port`: Worker port (e.g., `8001`)
- `--id`: Worker identifier

## Future Enhancements

The project has several planned features:

1. **Enhanced Retry Logic**: Automatic retry on worker failures with exponential backoff
2. **Heartbeat Monitoring**: Regular heartbeat checks to detect and handle dead workers
3. **HTTP File Transfer**: Send file contents through HTTP requests instead of file paths
4. **Raft Consensus Algorithm**: For distributed consensus and fault tolerance
5. **Load Balancing**: Dynamic load distribution across workers
6. **Caching**: Cache frequently accessed files or results
7. **Advanced Reducers**: Support for aggregation functions (count, average, etc.)
8. **Authentication**: Secure communication between master and workers
9. **Metrics Collection**: Prometheus-style metrics for monitoring
10. **Streaming Results**: Stream results as they arrive instead of waiting for all

## Development Notes

- See `notes.md` for development design decisions and architecture diagrams
- The current reducer simply returns the first matched value; extend `GrepReducer` for more complex logic
- Worker registration is automatic upon startup
- All logs are written to `distgrep.log` for debugging

## Example Workflow

```bash
# Terminal 1: Start Master
$ ./master
[MASTER] 2025-01-04 10:30:45 Master server running on :9000

# Terminal 2: Start Worker 1
$ ./worker --master=http://localhost:9000 --port=8001
[Worker] 2025-01-04 10:31:00 Registered to master at http://localhost:9000

# Terminal 3: Start Worker 2
$ ./worker --master=http://localhost:9000 --port=8002
[Worker] 2025-01-04 10:31:05 Registered to master at http://localhost:9000

# Terminal 4: Submit Job
$ curl -X POST http://localhost:9000/run \
  -H "Content-Type: application/json" \
  -d '{"pattern":"error", "files":["./logs/app.log"]}'

# Response shows all lines containing "error" from app.log
```

## License

This project is part of a distributed systems learning initiative.

## Contributing

Contributions are welcome! Please follow the existing code structure and add tests for new features.
