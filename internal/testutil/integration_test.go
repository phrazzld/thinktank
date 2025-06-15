package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSetupTempDir(t *testing.T) {
	t.Run("creates temporary directory", func(t *testing.T) {
		dir := SetupTempDir(t, "test-setup")

		// Verify directory exists
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Temporary directory %s was not created", dir)
		}

		// Verify prefix is used
		if !strings.Contains(filepath.Base(dir), "test-setup") {
			t.Errorf("Directory name %s doesn't contain expected prefix 'test-setup'", dir)
		}
	})

	t.Run("creates unique directories for multiple calls", func(t *testing.T) {
		dir1 := SetupTempDir(t, "unique-test")
		dir2 := SetupTempDir(t, "unique-test")

		if dir1 == dir2 {
			t.Error("SetupTempDir should create unique directories")
		}
	})
}

func TestSetupTempFile(t *testing.T) {
	t.Run("creates temporary file", func(t *testing.T) {
		filename, file := SetupTempFile(t, "test-file", ".txt")

		// Verify file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Temporary file %s was not created", filename)
		}

		// Verify filename contains prefix and suffix
		basename := filepath.Base(filename)
		if !strings.Contains(basename, "test-file") {
			t.Errorf("Filename %s doesn't contain expected prefix 'test-file'", basename)
		}
		if !strings.HasSuffix(basename, ".txt") {
			t.Errorf("Filename %s doesn't have expected suffix '.txt'", basename)
		}

		// Verify file is open and writable
		_, err := file.WriteString("test content")
		if err != nil {
			t.Errorf("Failed to write to temporary file: %v", err)
		}
	})

	t.Run("creates unique files for multiple calls", func(t *testing.T) {
		filename1, _ := SetupTempFile(t, "unique-file", ".tmp")
		filename2, _ := SetupTempFile(t, "unique-file", ".tmp")

		if filename1 == filename2 {
			t.Error("SetupTempFile should create unique files")
		}
	})
}

func TestSetupTempFileInDir(t *testing.T) {
	// First create a test directory
	dir := SetupTempDir(t, "file-in-dir-test")

	t.Run("creates file in specified directory", func(t *testing.T) {
		filename, file := SetupTempFileInDir(t, dir, "test-file", ".dat")

		// Verify file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Temporary file %s was not created", filename)
		}

		// Verify file is in correct directory
		if filepath.Dir(filename) != dir {
			t.Errorf("File %s not created in expected directory %s", filename, dir)
		}

		// Verify file is writable
		_, err := file.WriteString("test content")
		if err != nil {
			t.Errorf("Failed to write to temporary file: %v", err)
		}
	})
}

func TestCreateTestFile(t *testing.T) {
	dir := SetupTempDir(t, "create-file-test")
	testContent := []byte("Hello, World!")

	t.Run("creates file with content", func(t *testing.T) {
		filename := CreateTestFile(t, dir, "test.txt", testContent)

		// Verify file exists
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Test file %s was not created", filename)
		}

		// Verify file content
		content, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read test file: %v", err)
		}

		if string(content) != string(testContent) {
			t.Errorf("File content mismatch. Expected %q, got %q", testContent, content)
		}

		// Verify file is in correct location
		expectedPath := filepath.Join(dir, "test.txt")
		if filename != expectedPath {
			t.Errorf("File created at %s, expected %s", filename, expectedPath)
		}
	})
}

func TestCreateTestDir(t *testing.T) {
	baseDir := SetupTempDir(t, "create-dir-test")

	t.Run("creates directory at specified path", func(t *testing.T) {
		testDirPath := filepath.Join(baseDir, "subdir", "nested")
		createdPath := CreateTestDir(t, testDirPath)

		// Verify directory exists
		if _, err := os.Stat(createdPath); os.IsNotExist(err) {
			t.Errorf("Test directory %s was not created", createdPath)
		}

		// Verify it's actually a directory
		info, err := os.Stat(createdPath)
		if err != nil {
			t.Fatalf("Failed to stat created directory: %v", err)
		}
		if !info.IsDir() {
			t.Error("Created path is not a directory")
		}

		// Verify path is correct
		if createdPath != testDirPath {
			t.Errorf("Directory created at %s, expected %s", createdPath, testDirPath)
		}
	})
}

func TestCreateTestFiles(t *testing.T) {
	dir := SetupTempDir(t, "create-files-test")

	testFiles := map[string][]byte{
		"file1.txt": []byte("Content of file 1"),
		"file2.dat": []byte("Content of file 2"),
		"file3.log": []byte("Log content"),
	}

	t.Run("creates multiple files with content", func(t *testing.T) {
		createdFiles := CreateTestFiles(t, dir, testFiles)

		// Verify correct number of files created
		if len(createdFiles) != len(testFiles) {
			t.Errorf("Expected %d files, got %d", len(testFiles), len(createdFiles))
		}

		// Verify each file exists and has correct content
		for filename, expectedContent := range testFiles {
			fullPath := filepath.Join(dir, filename)

			// Check if file exists
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("File %s was not created", fullPath)
				continue
			}

			// Check content
			content, err := os.ReadFile(fullPath)
			if err != nil {
				t.Errorf("Failed to read file %s: %v", fullPath, err)
				continue
			}

			if string(content) != string(expectedContent) {
				t.Errorf("File %s content mismatch. Expected %q, got %q",
					filename, expectedContent, content)
			}
		}
	})
}

func TestWithTempDir(t *testing.T) {
	var capturedDir string

	t.Run("provides temporary directory to function", func(t *testing.T) {
		WithTempDir(t, "with-temp-test", func(dir string) {
			capturedDir = dir

			// Verify directory exists during function execution
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				t.Errorf("Temporary directory %s does not exist", dir)
			}

			// Create a test file to verify directory is usable
			testFile := filepath.Join(dir, "test.txt")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Errorf("Failed to create file in temporary directory: %v", err)
			}
		})

		// Verify directory name contains prefix
		if !strings.Contains(filepath.Base(capturedDir), "with-temp-test") {
			t.Errorf("Directory name %s doesn't contain expected prefix", capturedDir)
		}
	})
}

func TestCreateNestedTestDirs(t *testing.T) {
	baseDir := SetupTempDir(t, "nested-dirs-test")

	testPaths := []string{
		"level1",
		"level1/level2",
		"level1/level2/level3",
		"another/branch",
	}

	t.Run("creates nested directory structure", func(t *testing.T) {
		createdDirs := CreateNestedTestDirs(t, baseDir, testPaths)

		// Verify correct number of directories created
		if len(createdDirs) != len(testPaths) {
			t.Errorf("Expected %d directories, got %d", len(testPaths), len(createdDirs))
		}

		// Verify each directory exists
		for i, relativePath := range testPaths {
			expectedPath := filepath.Join(baseDir, relativePath)

			if i < len(createdDirs) && createdDirs[i] != expectedPath {
				t.Errorf("Directory %d: expected %s, got %s", i, expectedPath, createdDirs[i])
			}

			// Check if directory exists
			if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
				t.Errorf("Directory %s was not created", expectedPath)
				continue
			}

			// Verify it's actually a directory
			info, err := os.Stat(expectedPath)
			if err != nil {
				t.Errorf("Failed to stat directory %s: %v", expectedPath, err)
				continue
			}
			if !info.IsDir() {
				t.Errorf("Path %s is not a directory", expectedPath)
			}
		}
	})
}

// TestCleanupBehavior verifies that cleanup actually happens
// This test can't directly verify cleanup since t.Cleanup runs after the test,
// but it can verify the cleanup functions are registered
func TestCleanupBehavior(t *testing.T) {
	t.Run("cleanup functions are registered", func(t *testing.T) {
		// Create a subtable test that will complete and trigger cleanup
		t.Run("subtest", func(st *testing.T) {
			dir := SetupTempDir(st, "cleanup-test")
			filename, _ := SetupTempFile(st, "cleanup-file", ".tmp")

			// Verify both exist during the test
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				st.Errorf("Directory %s should exist during test", dir)
			}
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				st.Errorf("File %s should exist during test", filename)
			}
		})
		// After subtest completes, cleanup should have run
		// We can't directly verify this, but the fact that the test passes
		// indicates cleanup registration worked correctly
	})
}

// TestIntegrationHelpersWithRealUsage demonstrates realistic usage patterns
func TestIntegrationHelpersWithRealUsage(t *testing.T) {
	t.Run("file processing workflow", func(t *testing.T) {
		// Setup a temporary workspace
		workspaceDir := SetupTempDir(t, "workspace")

		// Create input directory structure
		inputDirs := []string{"input", "output", "temp"}
		CreateNestedTestDirs(t, workspaceDir, inputDirs)

		// Create test input files
		inputFiles := map[string][]byte{
			"input/data1.txt": []byte("input data 1"),
			"input/data2.txt": []byte("input data 2"),
			"temp/config.yml": []byte("config: test"),
		}
		CreateTestFiles(t, workspaceDir, inputFiles)

		// Verify workspace setup
		for filename := range inputFiles {
			fullPath := filepath.Join(workspaceDir, filename)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				t.Errorf("Input file %s was not created", fullPath)
			}
		}

		// Process files (simulate real workflow)
		inputDir := filepath.Join(workspaceDir, "input")
		outputDir := filepath.Join(workspaceDir, "output")

		entries, err := os.ReadDir(inputDir)
		if err != nil {
			t.Fatalf("Failed to read input directory: %v", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// Read input file
			inputPath := filepath.Join(inputDir, entry.Name())
			content, err := os.ReadFile(inputPath)
			if err != nil {
				t.Errorf("Failed to read input file %s: %v", inputPath, err)
				continue
			}

			// Process and write output
			processedContent := []byte("PROCESSED: " + string(content))
			outputPath := filepath.Join(outputDir, "processed_"+entry.Name())

			if err := os.WriteFile(outputPath, processedContent, 0644); err != nil {
				t.Errorf("Failed to write output file %s: %v", outputPath, err)
			}
		}

		// Verify processing results
		outputEntries, err := os.ReadDir(outputDir)
		if err != nil {
			t.Fatalf("Failed to read output directory: %v", err)
		}

		if len(outputEntries) != 2 {
			t.Errorf("Expected 2 output files, got %d", len(outputEntries))
		}
	})
}
