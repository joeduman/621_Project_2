package main

import (
	"io"
	"log"
	"net"
	"os"

	pb "primecounter/primespb" // <-- Replace with your real module name

	"google.golang.org/grpc"
)

// Fileserver struct
type fileServer struct {
	pb.UnimplementedFileServerServer
}

// StreamChunk RPC implementation (server-side streaming)
func (s *fileServer) StreamChunk(req *pb.FileRequest, stream pb.FileServer_StreamChunkServer) error {
	file, err := os.Open(req.Filename)
	if err != nil {
		log.Printf("Fileserver: Failed to open file: %v", err)
		return err
	}
	defer file.Close()

	// Seek to requested offset
	_, err = file.Seek(req.Start, 0)
	if err != nil {
		log.Printf("Fileserver: Failed to seek to offset %d: %v", req.Start, err)
		return err
	}

	remaining := req.Size
	bufferSize := int32(1024) // 1KB buffer chunks
	buffer := make([]byte, bufferSize)

	for remaining > 0 {
		toRead := bufferSize
		if remaining < bufferSize {
			toRead = remaining
		}

		n, err := file.Read(buffer[:toRead])
		if err != nil && err != io.EOF {
			log.Printf("Fileserver: Error reading file: %v", err)
			return err
		}
		if n == 0 {
			break
		}

		chunk := &pb.FileChunk{
			Content: buffer[:n],
		}

		if err := stream.Send(chunk); err != nil {
			log.Printf("Fileserver: Error sending chunk: %v", err)
			return err
		}

		remaining -= int32(n)
	}

	log.Printf("Fileserver: Finished streaming segment from offset %d, size %d", req.Start, req.Size)
	return nil
}

func main() {
	log.Println("Starting Fileserver...")

	listener, err := net.Listen("tcp", ":5003")
	if err != nil {
		log.Fatalf("Failed to listen on port 5003: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFileServerServer(grpcServer, &fileServer{})

	log.Println("Fileserver is running on port 5003")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve Fileserver: %v", err)
	}
}
