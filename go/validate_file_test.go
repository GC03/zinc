package main

import (
	"os"
	"testing"
)

func TestValidRootFile(t *testing.T) {
	// Create a new folder for testing purpose
	testFolder := "test_folder"
	err := os.Mkdir(testFolder, 0755)
	if err != nil {
		t.Fatalf("Failed to create test folder: %s", err)
	}

	// Create a new file for testing purpose
	testFile := "test_folder/test_file.txt"
	file, err := os.Create(testFile)
	if err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to create test file: %s", err)
	}
	file.Close()

	// Validate the folder contents
	requirements := make(map[string]string)
	requirements[testFile] = "haha"
	valid, err := ValidateDirectoryContents(testFolder, requirements)

	if !valid || err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to validate directory contents: %s", err)
	}

	// Clean up the test folder
	err = os.RemoveAll(testFolder)
	if err != nil {
		t.Fatalf("Failed to remove test folder: %s", err)
	}
}

func TestValidSubdirFile(t *testing.T) {
	// Create a new folder for testing purpose
	testFolder := "test_folder"
	err := os.Mkdir(testFolder, 0755)
	if err != nil {
		t.Fatalf("Failed to create test folder: %s", err)
	}

	// Create a new subfolder for testing purpose
	subFolder := "test_folder/sub_folder"
	err = os.Mkdir(subFolder, 0755)
	if err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to create sub folder: %s", err)
	}

	// Create a new file for testing purpose
	testFile := "test_folder/sub_folder/test_file.txt"
	file, err := os.Create(testFile)
	if err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to create test file: %s", err)
	}
	file.Close()

	// Validate the folder contents
	requirements := make(map[string]string)
	requirements[testFile] = "haha"
	valid, err := ValidateDirectoryContents(testFolder, requirements)

	if !valid || err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to validate directory contents: %s", err)
	}

	// Clean up the test folder
	err = os.RemoveAll(testFolder)
	if err != nil {
		t.Fatalf("Failed to remove test folder: %s", err)
	}
}

func TestUnchangedFile(t *testing.T) {
	// Create a new folder for testing purpose
	testFolder := "test_folder"
	err := os.Mkdir(testFolder, 0755)
	if err != nil {
		t.Fatalf("Failed to create test folder: %s", err)
	}

	// Create a new file for testing purpose
	testFile := "test_folder/test_file.txt"
	file, err := os.Create(testFile)
	if err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to create test file: %s", err)
	}
	file.Close()

	// Compute file hash of test file
	hash, err := computeFileHash(testFile)
	if err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to compute file hash: %s", err)
	}

	// Validate the folder contents
	requirements := make(map[string]string)
	requirements[testFile] = hash
	valid, err := ValidateDirectoryContents(testFolder, requirements)

	if valid || err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to validate directory contents: %s", err)
	}

	// Clean up the test folder
	err = os.RemoveAll(testFolder)
	if err != nil {
		t.Fatalf("Failed to remove test folder: %s", err)
	}
}

func TestMissingFile(t *testing.T) {
	// Create a new folder for testing purpose
	testFolder := "test_folder"
	err := os.Mkdir(testFolder, 0755)
	if err != nil {
		t.Fatalf("Failed to create test folder: %s", err)
	}

	// Validate the folder contents
	requirements := make(map[string]string)
	requirements["test_folder/missing_file.txt"] = "haha"
	valid, err := ValidateDirectoryContents(testFolder, requirements)

	if valid || err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to validate directory contents: %s", err)
	}

	// Clean up the test folder
	err = os.RemoveAll(testFolder)
	if err != nil {
		t.Fatalf("Failed to remove test folder: %s", err)
	}
}

func TestExtraFile(t *testing.T) {
	// Create a new folder for testing purpose
	testFolder := "test_folder"
	err := os.Mkdir(testFolder, 0755)
	if err != nil {
		t.Fatalf("Failed to create test folder: %s", err)
	}

	// Create an extra file for testing purpose
	extraFile := "test_folder/extra_file.txt"
	file, err := os.Create(extraFile)
	if err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to create extra file: %s", err)
	}
	file.Close()

	// Validate the folder contents
	requirements := make(map[string]string)
	valid, err := ValidateDirectoryContents(testFolder, requirements)

	if valid || err != nil {
		// Clean up the test folder
		err = os.RemoveAll(testFolder)
		if err != nil {
			t.Fatalf("Failed to remove test folder: %s", err)
		}

		t.Fatalf("Failed to validate directory contents: %s", err)
	}

	// Clean up the test folder
	err = os.RemoveAll(testFolder)
	if err != nil {
		t.Fatalf("Failed to remove test folder: %s", err)
	}
}
