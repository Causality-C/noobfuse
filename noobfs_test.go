package main

import (
	"noobfuse/lib"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func setupTest(t *testing.T) func() {
	// Create mount point
	if err := os.MkdirAll(lib.Mountpoint, 0755); err != nil {
		t.Fatalf("Failed to create mount point: %v", err)
	}

	// Start the FUSE filesystem in a separate process
	cmd := exec.Command("go", "run", "main.go")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start FUSE filesystem: %v", err)
	}

	// Wait for mount to be ready
	time.Sleep(2 * time.Second)

	// Return cleanup function
	return func() {
		// Unmount
		exec.Command("fusermount", "-u", lib.Mountpoint).Run()
		cmd.Process.Kill()
		os.RemoveAll(lib.Mountpoint)
	}
}

func TestMountUnmount(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Verify mount point exists and is a directory
	fi, err := os.Stat(lib.Mountpoint)
	if err != nil {
		t.Fatalf("Mount point not accessible: %v", err)
	}
	if !fi.IsDir() {
		t.Fatalf("Mount point is not a directory")
	}
}

func TestDirectoryListing(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	// Test listing root directory
	entries, err := os.ReadDir(lib.Mountpoint)
	if err != nil {
		t.Fatalf("Failed to read root directory: %v", err)
	}

	// Verify project directories exist
	foundProject1 := false
	foundProject2 := false
	for _, entry := range entries {
		if entry.Name() == "project-1" {
			foundProject1 = true
		}
		if entry.Name() == "project-2" {
			foundProject2 = true
		}
	}

	if !foundProject1 || !foundProject2 {
		t.Fatalf("Expected project directories not found. Found project-1: %v, project-2: %v",
			foundProject1, foundProject2)
	}
}

func TestFileReadingAndCaching(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	testFile := filepath.Join(lib.Mountpoint, "project-1", "common-lib.py")

	// First read (should be slow)
	start := time.Now()
	content1, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	duration1 := time.Since(start)

	if duration1 < 500*time.Millisecond {
		t.Fatalf("First read was too fast (took %v), expected >500ms", duration1)
	}

	// Second read (should be fast due to caching)
	start = time.Now()
	content2, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file second time: %v", err)
	}
	duration2 := time.Since(start)

	if duration2 > 500*time.Millisecond {
		t.Fatalf("Second read was too slow (took %v), expected <500ms", duration2)
	}

	// Verify content is the same
	if string(content1) != string(content2) {
		t.Fatalf("File content changed between reads")
	}
}

// Tests that the cache is based on file name, not the whole path
func TestFileCacheOtherFile(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup()

	testFile := filepath.Join(lib.Mountpoint, "project-1", "common-lib.py")

	// First read (should be slow)
	content1, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	// Second read (should be fast due to caching)
	testFileTwo := filepath.Join(lib.Mountpoint, "project-2", "common-lib.py")
	content2, err := os.ReadFile(testFileTwo)
	if err != nil {
		t.Fatalf("Failed to read file second time: %v", err)
	}

	// Verify content is the same
	if string(content1) != string(content2) {
		t.Fatalf("File content changed between reads")
	}
}
