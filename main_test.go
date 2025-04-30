package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempFile(content string) (*os.File, error) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "testfile-*.txt")
	if err != nil {
		return nil, err
	}

	// Write content to the temporary file
	if _, err := tempFile.WriteString(content); err != nil {
		return nil, err
	}

	// Close the file and return it
	if err := tempFile.Close(); err != nil {
		return nil, err
	}

	return tempFile, nil
}

func TestFindStringInFile(t *testing.T) {
	fileString := "@serranolabs.io\nThis is a test file with @serranolabs.io in it.\n"

	tempFile, _ := createTempFile(fileString)

	// Define the search string and replacement string
	searchString := "@serranolabs.io"
	localPath := "bookera-module-hub/packages/shared/module/module"

	// Call the function
	found, err := findStringInFile(tempFile.Name(), searchString, localPath)
	if err != nil {
		t.Fatalf("Error calling findStringInFile: %v", err)
	}

	// Verify the result
	if !found {
		t.Errorf("Expected to find the string '%s' in the file, but it was not found", searchString)
	}

	// Reopen the file to verify the replacement
	updatedContent, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	expectedContent := fmt.Sprintf("%s\nThis is a test file with %s in it.\n", localPath, localPath)

	if string(updatedContent) != expectedContent {
		t.Errorf("Expected file content to be '%s',\n but got '%s'", expectedContent, string(updatedContent))
	}
}

func TestCreateRelativePath(t *testing.T) {
	root := "/Users/davidserrano/Documents/dev/projects/bookera-rewrite/ui"
	path := "/Users/davidserrano/Documents/dev/projects/bookera-rewrite/ui/src/components/example.go"

	expected := "../../../"
	result := createRelativePath(root, path)

	if result != expected {
		t.Errorf("Expected '%s', but got '%s'", expected, result)
	}
}

func TestWalkThroughFiles(t *testing.T) {
	tempDir := t.TempDir()
	// Create a temporary file structure
	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")
	nodeModule := "@serranolabs.io/module/module"
	originalFile1 := fmt.Sprintf("This is a test file with %s in it.", nodeModule)
	err := os.WriteFile(file1Path, []byte(originalFile1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}

	err = os.WriteFile(file2Path, []byte("This file does not contain the search string."), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}

	newDir, _ := os.MkdirTemp(tempDir, "dir")
	nextDir, _ := os.MkdirTemp(newDir, "dir2")

	file3Path := filepath.Join(nextDir, "file3.txt")
	originalFile3 := fmt.Sprintf("This is a test file with %s in it.", nodeModule)
	err = os.WriteFile(file3Path, []byte(originalFile3), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file3: %v", err)
	}

	searchString := nodeModule
	localPath := "../../bookera-extensions-hub/packages/shared/module/module"

	err = walkThroughFiles(tempDir, searchString, localPath)
	if err != nil {
		t.Fatalf("walkThroughFiles returned an error: %v", err)
	}

	// Verify the content of the modified file
	updatedContent, err := os.ReadFile(file1Path)
	if err != nil {
		t.Fatalf("Failed to read updated file1: %v", err)
	}

	expectedContent := "This is a test file with ../../../bookera-extensions-hub/packages/shared/module/module in it."
	if string(updatedContent) != expectedContent {
		t.Errorf("Expected file content to be '%s', but got '%s'", expectedContent, string(updatedContent))
	}

	// Verify the content of the modified file
	updatedContent, err = os.ReadFile(file3Path)
	if err != nil {
		t.Fatalf("Failed to read updated file1: %v", err)
	}

	expectedContent = "This is a test file with ../../../../../bookera-extensions-hub/packages/shared/module/module in it."
	if string(updatedContent) != expectedContent {
		t.Errorf("Expected file content to be '%s', but got '%s'", expectedContent, string(updatedContent))
	}

	// Verify the unmodified file
	unmodifiedContent, err := os.ReadFile(file2Path)
	if err != nil {
		t.Fatalf("Failed to read file2: %v", err)
	}

	expectedUnmodifiedContent := "This file does not contain the search string."
	if string(unmodifiedContent) != expectedUnmodifiedContent {
		t.Errorf("Expected file2 content to be '%s', but got '%s'", expectedUnmodifiedContent, string(unmodifiedContent))
	}

	for path, replacePath := range replacedFiles {
		fmt.Printf("File: %s replaced with %s\n", path, replacePath)
	}

	savePathsToFile()

	relativePathToOriginalPath(replacedFiles, nodeModule)

	contents, _ := os.ReadFile(file3Path)
	assert.Equal(t, originalFile3, string(contents))
	contents, _ = os.ReadFile(file1Path)
	assert.Equal(t, originalFile1, string(contents))
}
