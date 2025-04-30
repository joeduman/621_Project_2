# Prime Counter - Distributed Version (Project 2)

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
go run generate.go 9600 # Default generates 75KB numbers file
```
---
## Expected Output and Coding Instructions
```
> go run fileserver.go
2025/04/29 22:19:08 Starting Fileserver...
2025/04/29 22:19:08 Fileserver is running on port 5003
2025/04/29 22:19:11 Fileserver: Finished streaming segment from offset 0, size 16384
2025/04/29 22:19:11 Fileserver: Finished streaming segment from offset 32768, size 16384
2025/04/29 22:19:11 Fileserver: Finished streaming segment from offset 16384, size 16384
2025/04/29 22:19:11 Fileserver: Finished streaming segment from offset 49152, size 16384
2025/04/29 22:21:14 Fileserver: Finished streaming segment from offset 65536, size 16384
```
---
```
> go run primes_main.go 16 1 numbers_75KB.dat primes_config.txt
2025/04/29 22:19:09 Starting Prime Counter Server...
2025/04/29 22:19:09 Dispatcher: Created 5 jobs.
2025/04/29 22:19:09 Wrote configuration to primes_config.txt
2025/04/29 22:19:09 Consolidator server running on port 5002
2025/04/29 22:19:09 Dispatcher server running on port 5001
2025/04/29 22:19:11 Dispatcher: Assigned job Offset=0 Size=16384
2025/04/29 22:19:11 Dispatcher: Assigned job Offset=16384 Size=16384
2025/04/29 22:19:11 Dispatcher: Assigned job Offset=32768 Size=16384
2025/04/29 22:19:11 Dispatcher: Assigned job Offset=49152 Size=16384
2025/04/29 22:21:14 Consolidator: Received primes=38 for segment starting at 0. Total so far: 38
2025/04/29 22:21:14 Dispatcher: Assigned job Offset=65536 Size=16384
2025/04/29 22:21:23 Consolidator: Received primes=44 for segment starting at 16384. Total so far: 82
2025/04/29 22:21:39 Consolidator: Received primes=52 for segment starting at 32768. Total so far: 134
2025/04/29 22:21:52 Consolidator: Received primes=57 for segment starting at 49152. Total so far: 191
2025/04/29 22:24:14 Consolidator: Received primes=58 for segment starting at 65536. Total so far: 249
```
---
# Terminal 1
```
> go run primes_worker.go
2025/04/29 22:19:11 Worker: Pulled job offset=32768 size=16384
2025/04/29 22:21:39 Worker finished processing job. Found 52 primes.
2025/04/29 22:21:39 Total elapsed time: 148.05s
2025/04/29 22:21:39 No more jobs from Dispatcher. Exiting...
```
---
# Terminal 2
```
> go run primes_worker.go
2025/04/29 22:19:11 Worker: Pulled job offset=49152 size=16384
2025/04/29 22:21:52 Worker finished processing job. Found 57 primes.
2025/04/29 22:21:52 Total elapsed time: 160.84s
2025/04/29 22:21:52 No more jobs from Dispatcher. Exiting...
```
---
# Terminal 3
```
> go run primes_worker.go
2025/04/29 22:19:11 Worker: Pulled job offset=16384 size=16384
2025/04/29 22:21:23 Worker finished processing job. Found 44 primes.
2025/04/29 22:21:23 Total elapsed time: 132.22s
2025/04/29 22:21:23 No more jobs from Dispatcher. Exiting...
```
---
# Terminal 4
```
> go run primes_worker.go
2025/04/29 22:19:11 Worker: Pulled job offset=0 size=16384
2025/04/29 22:21:14 Worker finished processing job. Found 38 primes.
2025/04/29 22:21:14 Total elapsed time: 123.20s
2025/04/29 22:21:14 Worker: Pulled job offset=65536 size=16384
2025/04/29 22:24:14 Worker finished processing job. Found 58 primes.
2025/04/29 22:24:14 Total elapsed time: 303.10s
2025/04/29 22:24:14 No more jobs from Dispatcher. Exiting...
```

# Running run_experiment.bat or run_experiment.bash script
```
> .\run_experiment.bat
[INFO] Killing previous servers...
Enter Segment Size N (in KB): 16 <-------------------------------------- User Input
Enter Chunk Size C (in KB): 1 <----------------------------------------- User Input
Enter Number of Workers (M): 4 <---------------------------------------- User Input
Enter Datafile Name (e.g., numbers_1MB.dat): numbers_80KB.dat <--------- User Input
Enter Config File Name (e.g., primes_config.txt): primes_config.txt <--- User Input
[INFO] Starting Fileserver...
[INFO] Starting Dispatcher and Consolidator...
[INFO] Launching 4 Worker Processes...
[INFO] Launching Worker 0...
[INFO] Launching Worker 1...
[INFO] Launching Worker 2...
[INFO] Launching Worker 3...
[INFO] Waiting for Workers to Finish...
[INFO] All workers finished. Combining logs...
[INFO] Logs combined into workers_combined.log
[INFO] Completed.
------------------------
> .\run_experiment.bash
[INFO] Killing previous servers...
Enter Segment Size N (in KB): 16 <-------------------------------------- User Input
Enter Chunk Size C (in KB): 1 <----------------------------------------- User Input
Enter Number of Workers (M): 4 <---------------------------------------- User Input
Enter Datafile Name (e.g., numbers_1MB.dat): numbers_80KB.dat <--------- User Input
Enter Config File Name (e.g., primes_config.txt): primes_config.txt <--- User Input
[INFO] Starting Fileserver...
[INFO] Starting Dispatcher and Consolidator...
[INFO] Launching 4 Worker Processes...
[INFO] Launching Worker 0...
[INFO] Launching Worker 1...
[INFO] Launching Worker 2...
[INFO] Launching Worker 3...
[INFO] Waiting for Workers to Finish...
[INFO] All workers finished. Combining logs...
[INFO] Logs combined into workers_combined.log
[INFO] Completed.
```

Each time an experiment is run, **.log** files are saved for the fileserver, main (dispatcher & consolidator), and each worker. When all workers are finished, the script combines the worker logs into one file for easier parsing (see above output [**Expected Output and Coding Instructions**] to see what is stored in the log files). 





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



