package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"

	pb "primecounter/primespb"

	"google.golang.org/grpc"
)

var (
	jobQueue    []pb.Job
	jobQueueMux sync.Mutex

	totalPrimes int
	totalMux    sync.Mutex
)

// Dispatcher server
type dispatcherServer struct {
	pb.UnimplementedDispatcherServer
}

// Consolidator server
type consolidatorServer struct {
	pb.UnimplementedConsolidatorServer
}

// Dispatcher gRPC method: Workers call this to get a Job
func (s *dispatcherServer) GetJob(ctx context.Context, empty *pb.Empty) (*pb.Job, error) {
	jobQueueMux.Lock()
	defer jobQueueMux.Unlock()

	if len(jobQueue) == 0 {
		return &pb.Job{}, nil // No more jobs
	}

	job := jobQueue[0]
	jobQueue = jobQueue[1:]
	log.Printf("Dispatcher: Assigned job Offset=%d Size=%d\n", job.Start, job.Size)
	return &job, nil
}

// Consolidator gRPC method: Workers call this to submit results
func (s *consolidatorServer) SubmitResult(ctx context.Context, result *pb.Result) (*pb.Ack, error) {
	totalMux.Lock()
	defer totalMux.Unlock()

	totalPrimes += int(result.PrimeCount)
	log.Printf("Consolidator: Received primes=%d for segment starting at %d. Total so far: %d\n",
		result.PrimeCount, result.Start, totalPrimes)

	return &pb.Ack{Message: "Result received"}, nil
}

// Main function
func main() {
	if len(os.Args) < 5 {
		log.Fatal("Usage: go run primes_main.go <N> <C> <datafile> <configfile>")
	}

	log.Println("Starting Prime Counter Server...")

	// Parse command-line arguments
	segmentSize := parseInt(os.Args[1]) * 1024 // N in KB
	//chunkSize := parseInt(os.Args[2]) * 1024    // C in KB
	dataFilePath := os.Args[3]
	configPath := os.Args[4]

	// Open the data file
	fileInfo, err := os.Stat(dataFilePath)
	if err != nil {
		log.Fatalf("Failed to open datafile: %v", err)
	}
	fileSize := fileInfo.Size()

	// Create the job queue by splitting file into segments
	for offset := int64(0); offset < fileSize; offset += int64(segmentSize) {
		size := segmentSize
		if offset+int64(segmentSize) > fileSize {
			size = int(fileSize - offset)
		}
		jobQueue = append(jobQueue, pb.Job{
			Filename: dataFilePath,
			Start:    offset,
			Size:     int32(size),
		})
	}
	log.Printf("Dispatcher: Created %d jobs.\n", len(jobQueue))

	// Write primes_config.txt file
	configContent := fmt.Sprintf(`dispatcher localhost 5001
consolidator localhost 5002
fileserver localhost 5003
`)
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		log.Fatalf("Failed to write config file: %v", err)
	}
	log.Printf("Wrote configuration to %s", configPath)

	// Start the servers
	go startDispatcherServer()
	go startConsolidatorServer()

	// Wait forever
	select {}
}

// Start Dispatcher gRPC server
func startDispatcherServer() {
	listener, err := net.Listen("tcp", ":5001")
	if err != nil {
		log.Fatalf("Failed to listen on port 5001: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterDispatcherServer(grpcServer, &dispatcherServer{})
	log.Println("Dispatcher server running on port 5001")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve Dispatcher: %v", err)
	}
}

// Start Consolidator gRPC server
func startConsolidatorServer() {
	listener, err := net.Listen("tcp", ":5002")
	if err != nil {
		log.Fatalf("Failed to listen on port 5002: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterConsolidatorServer(grpcServer, &consolidatorServer{})
	log.Println("Consolidator server running on port 5002")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve Consolidator: %v", err)
	}
}

// Parses string to int safely
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("Invalid integer input: %v", s)
	}
	return val
}
