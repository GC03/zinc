package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Define a struct to hold the line data
type LineData struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

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

	// Start the executable
	cmd := exec.Command("../run")

	// Get the stdout pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Create a file to write the logs
	filename := "logs.json"
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Initialize line counter
	lineNumber := 0

	// Create and initialize a struct that contains an array to hold the output lines, and this entire struct is to be Marshaled into json
	type Output struct {
		Lines []LineData `json:"lines"`
	}

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

		// Marshal the output into JSON
		jsonData, err := json.Marshal(output)
		if err != nil {
			log.Fatal(err)
		}

		// Write to temporary file
		tempFile := filename + ".tmp"
		if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
			log.Fatal(fmt.Errorf("error writing temporary file: %w", err))
		}

		// Rename temporary file to target file
		if err := os.Rename(tempFile, filename); err != nil {
			log.Fatal(fmt.Errorf("error renaming file: %w", err))
		}

		// Uploads the json file to minio
		_, err = minioClient.FPutObject(context.Background(), "logs", "log", filename, minio.PutObjectOptions{ContentType: "application/json"})
		if err != nil {
			log.Fatal(err)
		}

		// Increment line number
		lineNumber++
	}

	// Check for errors in the scanner
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
