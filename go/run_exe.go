package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// Set up connection to minio
	endpoint := "localhost:9000"
	accessKeyID := "adminhaha"
	secretAccessKey := "adminhaha"
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%#v\n", minioClient)

	testProgramPath := "../run"
	testFilePath := "../testFile.txt"

	// Create a new ExecutableRunner
	runner := NewExecutableRunner(testProgramPath, testFilePath, minioClient, "logs", "log", 1)
	fmt.Println("Runner created")

	var wg sync.WaitGroup
	wg.Add(1)

	// Run the executable
	go func() {
		defer wg.Done()
		runner.Run()
	}()

	time.Sleep(time.Second)

	// // Print memory usage
	go runner.GetMemoryUsage()

	// Print CPU usage
	go runner.GetCPUUsage()

	wg.Wait()
}
