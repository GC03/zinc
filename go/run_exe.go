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
	valid, err := ValidateDirectoryContents("test_folder", map[string]string{"test_folder/a": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"})
	if err != nil {
		log.Fatal(err)
	}
	if !valid {
		log.Fatal("Directory contents are not valid")
	} else {
		fmt.Println("Directory contents are verified")
	}

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
