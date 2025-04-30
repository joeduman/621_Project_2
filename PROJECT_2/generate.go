package main

import (
	"encoding/binary"
	"log"
	"math/rand"
	"os"
	"strconv"
)

// Default number of entries (~10MB)
const defaultEntries = 1310720

func main() {
	// Read number of entries from command-line argument (if provided)
	numEntries := defaultEntries
	if len(os.Args) > 1 {
		n, err := strconv.Atoi(os.Args[1])
		if err != nil || n <= 0 {
			log.Fatal("Invalid number of entries. Please enter a positive integer.")
		}
		numEntries = n
	}

	// Create binary file
	dataFile, err := os.Create("numbers_85KB.dat") // CHANGE OUTPUT NAME
	if err != nil {
		log.Fatal("Error creating numbers.dat:", err)
	}
	defer dataFile.Close()

	// Generate and write random 64-bit unsigned integers
	for i := 0; i < numEntries; i++ {
		num := uint64(rand.Uint64()) // Generate full 64-bit unsigned integer
		binary.Write(dataFile, binary.LittleEndian, num)
	}

	log.Printf("Generated %d numbers in numbers_XMB/KB.dat\n", numEntries)
}
