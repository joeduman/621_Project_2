# Prime Counter - 621 (Project 2)

## Overview
This project extends Project 1 by distributing the computation of prime numbers across independent worker processes communicating over gRPC.  
The system consists of:
- Dispatcher Server (assigns jobs)
- Fileserver (streams file data)
- Consolidator Server (collects results)
- Multiple Worker processes (fetch jobs, stream file segments, process primes)

All components communicate over TCP using gRPC (localhost is sufficient).

---
## Prerequisites
- Go (Golang) installed: https://go.dev/doc/install
- Protocol Buffers Compiler (protoc) installed
- gRPC and protobuf Go plugins installed:

```
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```
---
## Building the Project
### 1. Initialize the Go module (if not already):

```
go mod init prime-counter go mod tidy
```

### 2. Compile the `.proto` file:

```
protoc --go_out=. --go-grpc_out=. primes.proto
```
(This generates primes.pb.go and primes_grpc.pb.go.)

3. Build each component if desired:

```
go build -o primes_main primes_main.go
go build -o primes_worker primes_worker.go
go build -o fileserver fileserver.go
```

---
## Running the Project

You must start these processes separately:

### 1. Start Fileserver
```
go run fileserver.go
```

### 2. Start Dispatcher and Consolidator (Main Controller)
```
go run primes_main.go <SegmentSize> <ChunkSize> <DatafilePath> <ConfigFilePath>
```

SegmentSize and ChunkSize are already calculated as KB inside of the code. Please enter integers instead.

**Example:**
```
go run primes_main.go 64 1 numbers.dat primes_config.txt
```

- `<SegmentSize>` : Size of each file segment (in KB) that a worker will process.
- `<ChunkSize>` : Streaming chunk size (in KB) for reading file data.
- `<DatafilePath>` : Path to the .dat binary file containing 64-bit integers.
- `<ConfigFilePath>` : Path to the configuration file (see below).

### 3. Start Workers
Launch multiple independent workers in separate terminals:

go run primes_worker.go <ChunkSize> <ConfigFilePath>

**You can start 4, 8, or 16 workers manually by opening multiple terminals.**

---
## Configuration File (primes_config.txt)
This file stores the addresses for all servers.

**Example primes_config.txt**
```
dispatcher localhost 5001
consolidator localhost 5002
fileserver localhost 5003
<service_name> <host> <port>
```

All workers and main processes read this file at startup.
---
## Generating a Sample Data File
You can generate a random binary file for testing:
```
go run generate.go 1000000 # Default generates 10MB numbers file
```
---
## Expected Output


---
## Notes
- Ensure the Fileserver, Dispatcher, and Consolidator are started **before** starting workers.
- More workers improve speed, but only up to the number of available jobs.
- Optimal performance requires careful tuning of:
  - Number of workers (M)
  - Segment size (N)
  - Chunk size (C)
- This project assumes all components run on localhost, but supports distribution across multiple machines via IPs in primes_config.txt.

---



