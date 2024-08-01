package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/minio/minio-go/v7"
)

// Define a struct to hold the line data
type LineData struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

// Create and initialize a struct that contains an array to hold the output lines, and this entire struct is to be Marshaled into json
type Log struct {
	Lines []LineData `json:"lines"`
}

type ExecutableRunner struct {
	executablePath string
	testcasePath   string
	minioClient    *minio.Client
	bucketName     string
	objectName     string
	timeOut        int
	cmd            *exec.Cmd
}

func NewExecutableRunner(executablePath string, testcasePath string, minioClient *minio.Client, bucketName string, objectName string, timeout int) *ExecutableRunner {
	return &ExecutableRunner{
		executablePath: executablePath,
		testcasePath:   testcasePath,
		minioClient:    minioClient,
		bucketName:     bucketName,
		objectName:     objectName,
		timeOut:        timeout,
	}
}

func (e *ExecutableRunner) Run() {
	// Set a timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(e.timeOut)*time.Minute)
	defer cancel()
	e.cmd = exec.CommandContext(ctx, e.executablePath)

	// Get the stdin pipe
	stdin, err := e.cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Get the stdout pipe
	stdout, err := e.cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	// Start the command
	if err := e.cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Initialize line counter
	lineNumber := 0

	// Create a new Log struct
	logContent := Log{
		Lines: []LineData{},
	}

	go func() {
		// Read the test file
		file, err := os.Open(e.testcasePath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		// Create a new scanner
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			// Write the line to the stdin of the command
			_, err := fmt.Fprintln(stdin, scanner.Text())
			if err != nil {
				log.Fatal(err)
			}

			// Create a new line struct
			line := LineData{
				ID:   lineNumber,
				Text: scanner.Text(),
			}

			// Append the line to the output struct
			logContent.Lines = append(logContent.Lines, line)

			// Increment line number
			lineNumber++
			fmt.Println(line)

			// Upload after 5 lines accumulated
			if lineNumber%5 == 0 {
				uploadToMinio(e.minioClient, e.bucketName, e.objectName, &logContent)
			}

			// time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		// Create a new scanner
		scanner := bufio.NewScanner(stdout)

		// Loop over the stdout and write to the file
		for scanner.Scan() {
			line := LineData{
				ID:   lineNumber,
				Text: scanner.Text(),
			}

			// Append the line to the output struct
			logContent.Lines = append(logContent.Lines, line)

			// Increment line number
			lineNumber++
			fmt.Println(line)

			// Upload after 5 lines accumulated
			if lineNumber%5 == 0 {
				uploadToMinio(e.minioClient, e.bucketName, e.objectName, &logContent)
			}
		}

		// Check for errors in the scanner
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait for the command to finish
	if err := e.cmd.Wait(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			// Add err to the output
			line := LineData{
				ID:   lineNumber,
				Text: fmt.Sprintf("Command timed out at %s", time.Now().String()),
			}
			logContent.Lines = append(logContent.Lines, line)

			// Upload output's content before timed out
			uploadToMinio(e.minioClient, e.bucketName, e.objectName, &logContent)
			log.Fatal("Command timed out")
		} else {
			log.Fatal(err)
		}
	}

	// Upload the final output
	uploadToMinio(e.minioClient, e.bucketName, e.objectName, &logContent)
}

func uploadToMinio(minioClient *minio.Client, bucketName string, objectName string, content *Log) {
	// Marshal the output into JSON
	jsonData, err := json.Marshal(*content)
	if err != nil {
		log.Fatal(err)
	}

	// Upload the json as an object directly to minio
	_, err = minioClient.PutObject(context.Background(), bucketName, objectName, bytes.NewReader(jsonData), int64(len(jsonData)), minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Uploaded to minio")
}

func (e *ExecutableRunner) GetMemoryUsage() {
	for {
		pid := e.cmd.Process.Pid
		mem, err := calculateMemory(pid)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Memory used: %d KB\n", mem)
		time.Sleep(1 * time.Second)
	}
}

func (e *ExecutableRunner) GetCPUUsage() {
	for {
		pid := e.cmd.Process.Pid
		cpu, err := calculateCPU(pid)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("CPU used: %f\n", cpu)
		time.Sleep(1 * time.Second)
	}
}

func calculateMemory(pid int) (uint64, error) {

	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	res := uint64(0)
	pfx := []byte("Pss:")
	r := bufio.NewScanner(f)
	for r.Scan() {
		line := r.Bytes()
		if bytes.HasPrefix(line, pfx) {
			var size uint64
			_, err := fmt.Sscanf(string(line[4:]), "%d", &size)
			if err != nil {
				return 0, err
			}
			res += size
		}
	}
	if err := r.Err(); err != nil {
		return 0, err
	}

	return res, nil
}

func calculateCPU(pid int) (float64, error) {
	total0, process0 := readTotalCPUSnapshot(), readProcessCPUSnapshot(pid)
	time.Sleep(1 * time.Second)
	total1, process1 := readTotalCPUSnapshot(), readProcessCPUSnapshot(pid)

	cpuUsage := float64(process1-process0) / float64(total1-total0) * 100

	return cpuUsage, nil
}

func readTotalCPUSnapshot() int {
	f, err := os.Open("/proc/stat")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // read the first line
	line := scanner.Text()

	var user, nice, system, idle, iowait, irq, softirq, steal, guest, guest_nice int

	// fmt.Println(line)
	_, err = fmt.Sscanf(line, "cpu %d %d %d %d %d %d %d %d %d %d", &user, &nice, &system, &idle, &iowait, &irq, &softirq, &steal, &guest, &guest_nice)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("user: %d\n", user)
	// fmt.Printf("nice: %d\n", nice)
	// fmt.Printf("system: %d\n", system)
	// fmt.Printf("idle: %d\n", idle)
	// fmt.Printf("iowait: %d\n", iowait)
	// fmt.Printf("irq: %d\n", irq)
	// fmt.Printf("softirq: %d\n", softirq)
	// fmt.Printf("steal: %d\n", steal)
	// fmt.Printf("guest: %d\n", guest)
	// fmt.Printf("guest_nice: %d\n", guest_nice)

	fmt.Printf("Total: %d\n", user+nice+system+idle+iowait+irq+softirq+steal+guest+guest_nice)

	return (user + nice + system + idle + iowait + irq + softirq + steal + guest + guest_nice)
}

func readProcessCPUSnapshot(pid int) int {
	f, err := os.Open(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan() // read the first line
	line := scanner.Text()

	var p, ppid, pgrp, session, flags, tty_nr, tpgid, minflt, cminflt, majflt, cmajflt, utime, stime, cutime, cstime int
	var comm string
	var state rune

	// fmt.Println(line)
	_, err = fmt.Sscanf(line, "%d %s %c %d %d %d %d %d %d %d %d %d %d %d %d %d %d", &p, &comm, &state, &ppid, &pgrp, &session, &tty_nr, &tpgid, &flags, &minflt, &cminflt, &majflt, &cmajflt, &utime, &stime, &cutime, &cstime)
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("utime: %d\n", utime)
	// fmt.Printf("stime: %d\n", stime)
	// fmt.Printf("cutime: %d\n", cutime)
	// fmt.Printf("cstime: %d\n", cstime)

	fmt.Printf("Process: %d\n", utime+stime+cutime+cstime)

	return (utime + stime + cutime + cstime)
}
