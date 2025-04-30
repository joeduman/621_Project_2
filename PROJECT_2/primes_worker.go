package main

import (
	"context"
	"encoding/binary"
	"io"
	"log"
	"os"
	"strings"
	"time"
	"strconv"

	pb "primecounter/primespb" // <-- CHANGE to your real module name

	"google.golang.org/grpc"
)

// Simple struct to store config
type serverAddresses struct {
	dispatcher   string
	consolidator string
	fileserver   string
}

// Parses primes_config.txt
func parseConfig(configPath string) serverAddresses {
	content, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	lines := strings.Split(string(content), "\n")
	var addresses serverAddresses

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue
		}
		addr := parts[1] + ":" + parts[2]
		switch parts[0] {
		case "dispatcher":
			addresses.dispatcher = addr
		case "consolidator":
			addresses.consolidator = addr
		case "fileserver":
			addresses.fileserver = addr
		}
	}
	return addresses
}

// Checks if a number is prime
func isPrime(n uint64) bool {
	if n < 2 {
		return false
	}
	if n%2 == 0 && n != 2 {
		return false
	}
	for i := uint64(3); i*i <= n; i += 2 {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	startTotal := time.Now()
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run primes_worker.go <chunkSize> <configfile>")
	}

	//chunkSize := parseInt(os.Args[1]) * 1024 // Convert KB to Bytes
	configPath := os.Args[2]

	addresses := parseConfig(configPath)

	// Connect to Dispatcher
	dispatchConn, err := grpc.Dial(addresses.dispatcher, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Dispatcher: %v", err)
	}
	defer dispatchConn.Close()
	dispatchClient := pb.NewDispatcherClient(dispatchConn)

	// Connect to Fileserver
	fileConn, err := grpc.Dial(addresses.fileserver, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Fileserver: %v", err)
	}
	defer fileConn.Close()
	fileClient := pb.NewFileServerClient(fileConn)

	// Connect to Consolidator
	consConn, err := grpc.Dial(addresses.consolidator, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Consolidator: %v", err)
	}
	defer consConn.Close()
	consClient := pb.NewConsolidatorClient(consConn)

	// Main work loop
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Pull a job
		job, err := dispatchClient.GetJob(ctx, &pb.Empty{})
		if err != nil {
			log.Printf("Error getting job: %v", err)
			break
		}
		if job.Size == 0 {
			log.Println("No more jobs from Dispatcher. Exiting...")
			break
		}

		log.Printf("Worker: Pulled job offset=%d size=%d\n", job.Start, job.Size)

		// Stream file segment
		req := &pb.FileRequest{
			Filename: job.Filename,
			Start:    job.Start,
			Size:     job.Size,
		}

		stream, err := fileClient.StreamChunk(context.Background(), req)
		if err != nil {
			log.Printf("Failed to stream chunk: %v", err)
			continue
		}

		primeCount := 0

		for {
			chunk, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("Error receiving chunk: %v", err)
				break
			}

			data := chunk.Content

			for i := 0; i+8 <= len(data); i += 8 {
				num := binary.LittleEndian.Uint64(data[i : i+8])
				if isPrime(num) {
					primeCount++
				}
			}
		}

		totalElapsed := time.Since(startTotal)
		log.Printf("Worker finished processing job. Found %d primes.", primeCount)
		log.Printf("Total elapsed time: %.2fs\n", totalElapsed.Seconds())
		// Submit result to Consolidator
		_, err = consClient.SubmitResult(context.Background(), &pb.Result{
			Start:      job.Start,
			PrimeCount: int32(primeCount),
		})
		if err != nil {
			log.Printf("Error submitting result: %v", err)
		}
	}
}

// Safely parse int from string
func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal("Invalid integer:", s)
	}
	return val
}
