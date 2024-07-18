package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Define a struct to hold the line data
type LineData struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// Create and initialize a struct that contains an array to hold the output lines, and this entire struct is to be Marshaled into json
type Output struct {
	Lines []LineData `json:"lines"`
}

var MAX_TIMEOUT = 15 * time.Minute

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

	// Set a timeout
	ctx, cancel := context.WithTimeout(context.Background(), MAX_TIMEOUT) // 15 minutes timeout
	defer cancel()
	cmd := exec.CommandContext(ctx, "../run")

	// Get the stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Initialize line counter
	lineNumber := 0

	// Create a new Output struct
	output := Output{
		Lines: []LineData{},
	}

	// Create a new scanner
	scanner := bufio.NewScanner(stdout)

	// Loop over the stdout and write to the file
	for scanner.Scan() {
		line := LineData{
			ID:   lineNumber,
			Text: scanner.Text(),
		}

		// Append the line to the output struct
		output.Lines = append(output.Lines, line)

		// Increment line number
		lineNumber++
		fmt.Println(lineNumber)

		// Upload after 5 lines accumulated
		if lineNumber%5 == 0 {
			uploadToMinio(minioClient, output)
		}
	}

	// Check for errors in the scanner
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			// Upload output's content before timed out
			uploadToMinio(minioClient, output)
			log.Fatal("Command timed out")
		} else {
			log.Fatal(err)
		}
	}

	// Upload the final output
	uploadToMinio(minioClient, output)
}

func uploadToMinio(minioClient *minio.Client, content Output) {
	// Marshal the output into JSON
	jsonData, err := json.Marshal(content)
	if err != nil {
		log.Fatal(err)
	}

	// Upload the json as an object directly to minio
	_, err = minioClient.PutObject(context.Background(), "logs", "log", bytes.NewReader(jsonData), int64(len(jsonData)), minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Uploaded to minio")
}
