package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"sync"
)

func Hasher(filename string, expectedHashBytes []byte, start int64, end int64, running *bool) {
	log.Printf("Worker hashing from %d to %d", start, end)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	h := md5.New()

	data := make([]byte, 8192)
	var readBytes int64
	for *running {
		count, err := file.Read(data)
		readBytes += int64(count)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading: %v", err)
		}
		h.Write(data[:count])
		hhash := h.Sum(nil)
		if bytes.Compare(hhash, expectedHashBytes) == 0 {
			log.Printf("Found hash %x after %d bytes", expectedHashBytes, readBytes)
			*running = false
			break
		}
		if readBytes >= (start) && len(data) > 10 {
			//log.Printf("Skipped until %d", readBytes)
			data = make([]byte, 1)
		}
		if readBytes > end+8192 {
			break
		}
	}
}

func main() {
	if len(os.Args) < 3 {
		log.Printf("Usage: %s filename md5", os.Args[0])
		os.Exit(1)
	}
	filename := os.Args[1]
	expectedHash := os.Args[2]

	log.Printf("Hashing %s until %s is found", filename, expectedHash)
	expectedHashBytes, err := hex.DecodeString(expectedHash)
	if err != nil {
		log.Fatal(err)
	}

	fi, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("The file is %d bytes long", fi.Size())

	var chunkSize int64
	chunkSize = (fi.Size() / 12) + 1024

	var wg sync.WaitGroup

	var start int64
	var running bool
	running = true
	for start = 0; start < fi.Size(); start += chunkSize {
		wg.Add(1)
		go func(startByte int64) {
			Hasher(filename, expectedHashBytes, startByte, startByte+chunkSize, &running)
			wg.Done()
		}(start)
	}
	wg.Wait()

}
