package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func ValidateDirectoryContents(path string, requirements map[string]string) (bool, error) {
	// Return err if path not found
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, fmt.Errorf("path not found: %s", path)
	}

	valid := true
	foundFiles := make(map[string]bool)
	unchangedFiles := make(map[string]bool)

	// Walk through the directory
	err := filepath.WalkDir(path, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Store found file
		foundFiles[filePath] = true

		// Check if file is in requirements
		expectedHash, exists := requirements[filePath]
		if exists {
			// Compute file hash
			hash, err := computeFileHash(filePath)
			if err != nil {
				return err
			}

			// Check if hashes match
			if hash == expectedHash {
				unchangedFiles[filePath] = true
			}
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	// Check for missing files
	for reqFile := range requirements {
		if !foundFiles[reqFile] {
			fmt.Println("Missing file:", reqFile)
			valid = false
		}
	}

	// Report unchanged files
	for unchangedFile := range unchangedFiles {
		fmt.Println("Unchanged file:", unchangedFile)
		valid = false
	}

	// Report extra files
	for foundFile := range foundFiles {
		if _, exists := requirements[foundFile]; !exists {
			fmt.Println("Extra file:", foundFile)
			valid = false
		}
	}

	return valid, nil
}

func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
