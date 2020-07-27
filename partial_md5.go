package main

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
	"log"
	"os"
	"reflect"
	"sync"
)

//From Reddit, bad idea
func copyHash(src hash.Hash) hash.Hash {
	typ := reflect.TypeOf(src)
	val := reflect.ValueOf(src)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	elem := reflect.New(typ).Elem()
	elem.Set(val)
	return elem.Addr().Interface().(hash.Hash)
}

func Hasher(filename string, expectedHashBytes []byte, hasher hash.Hash, start int64, end int64, running *bool) {
	log.Printf("Worker hashing from %d to %d", start, end)
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Seek(start, io.SeekStart)
	if err != nil {
		log.Printf("Error seeking to start position: %v", err)
		return
	}

	reader := bufio.NewReader(file)

	data := make([]byte, 1)
	var readBytes int64 = start
	for *running {
		count, err := reader.Read(data)
		readBytes += int64(count)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error reading: %v", err)
		}
		hasher.Write(data[:count])
		hhash := hasher.Sum(nil)
		if bytes.Compare(hhash, expectedHashBytes) == 0 {
			log.Printf("Found hash %x after %d bytes", expectedHashBytes, readBytes)
			*running = false
			break
		}
		if readBytes > end {
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

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	hasher := md5.New()

	var running bool
	running = true
	var wg sync.WaitGroup

	var cur, pos int64
	var chunkSize int64
	chunkSize = 128 * 1024 * 1024
	buf := make([]byte, 1024)

	for pos < fi.Size() {

		wg.Add(1)
		go func(startByte int64, hasher hash.Hash) {
			Hasher(filename, expectedHashBytes, hasher, startByte, startByte+chunkSize, &running)
			wg.Done()
		}(pos, copyHash(hasher))

		//Advance the file and the hasher by chunkSize bytes
		for cur < pos+chunkSize {
			count, err := file.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Fatal(err)
				}
				break
			}
			cur += int64(count)
			hasher.Write(buf[:count])
		}
		pos = cur
	}
	wg.Wait()

}
