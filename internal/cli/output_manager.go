// Package cli provides output directory management for the thinktank tool
package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/misty-step/thinktank/internal/logutil"
)

// OutputManager handles intelligent output directory naming and management
type OutputManager struct {
	logger logutil.LoggerInterface

	memorableOffset  int
	memorableStride  int
	memorableTotal   int
	memorableCounter atomic.Uint32
}

// NewOutputManager creates a new output manager instance
func NewOutputManager(logger logutil.LoggerInterface) *OutputManager {
	if logger == nil {
		logger = logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	total := len(adjectives) * len(verbs) * len(nouns)
	offset := 0
	stride := 1
	if total > 0 {
		offset = rng.Intn(total)
		stride = pickCoprimeStride(rng, total)
	}

	return &OutputManager{
		logger:          logger,
		memorableOffset: offset,
		memorableStride: stride,
		memorableTotal:  total,
	}
}

// Global counter for ensuring uniqueness across concurrent operations
var incrementalCounter uint32

const (
	memorableNameMinLength = 20
	memorableNameMaxLength = 40
	memorableNameAttempts  = 50
	maxCollisionAttempts   = 10
)

var adjectives = []string{
	"ancient",
	"amber",
	"auburn",
	"azure",
	"bright",
	"bronze",
	"calm",
	"candid",
	"cerulean",
	"chilly",
	"crimson",
	"copper",
	"cozy",
	"dapper",
	"distant",
	"gentle",
	"golden",
	"granite",
	"graceful",
	"hollow",
	"humble",
	"indigo",
	"jovial",
	"lively",
	"magenta",
	"mellow",
	"modern",
	"molten",
	"muted",
	"nimble",
	"oaken",
	"orange",
	"placid",
	"polished",
	"primal",
	"proud",
	"quiet",
	"russet",
	"saffron",
	"sable",
	"silken",
	"silver",
	"simple",
	"sleek",
	"smooth",
	"steady",
	"sturdy",
	"subtle",
	"sunlit",
	"tender",
	"timber",
	"tranquil",
	"vivid",
	"warm",
	"weighty",
	"yellow",
	"zealous",
}

var verbs = []string{
	"ambling",
	"baking",
	"blooming",
	"bouncing",
	"building",
	"carving",
	"chasing",
	"circling",
	"climbing",
	"cooking",
	"curling",
	"dancing",
	"drifting",
	"drumming",
	"floating",
	"flowing",
	"flying",
	"gliding",
	"growing",
	"hiking",
	"humming",
	"jumping",
	"laughing",
	"leaping",
	"lifting",
	"marching",
	"mixing",
	"moving",
	"nesting",
	"pacing",
	"painting",
	"playing",
	"pouring",
	"racing",
	"roaming",
	"rolling",
	"running",
	"sailing",
	"scaling",
	"seeking",
	"shaping",
	"shining",
	"singing",
	"skating",
	"sliding",
	"soaring",
	"spinning",
	"springing",
	"stirring",
	"streaming",
	"striding",
	"swinging",
	"swirling",
	"tending",
	"trailing",
	"traveling",
	"turning",
	"wandering",
	"working",
	"zooming",
}

var nouns = []string{
	"acorn",
	"badger",
	"bamboo",
	"bison",
	"butter",
	"canyon",
	"cedar",
	"cherry",
	"comet",
	"coral",
	"coyote",
	"crystal",
	"dahlia",
	"dolphin",
	"falcon",
	"forest",
	"galaxy",
	"garden",
	"glacier",
	"harbor",
	"hazelnut",
	"heron",
	"juniper",
	"kingfisher",
	"lagoon",
	"lantern",
	"meadow",
	"meteor",
	"monarch",
	"mountain",
	"orchard",
	"otter",
	"palmier",
	"prairie",
	"quartz",
	"rabbit",
	"raven",
	"river",
	"rocket",
	"saffron",
	"satchel",
	"savanna",
	"sierra",
	"sparrow",
	"spruce",
	"starfish",
	"summit",
	"thistle",
	"tigress",
	"truffle",
	"tundra",
	"valley",
	"violet",
	"walnut",
	"willow",
	"window",
	"zephyr",
}

func init() {
	if len(adjectives) == 0 || len(verbs) == 0 || len(nouns) == 0 {
		panic("word lists cannot be empty")
	}
}

// Keep word lists additive to preserve cleanup recognition for older runs.
var adjectiveSet = makeWordSet(adjectives)
var verbSet = makeWordSet(verbs)
var nounSet = makeWordSet(nouns)

// GenerateTimestampedDirName generates a unique directory name in the format:
// thinktank_YYYYMMDD_HHMMSS_NNNNNNNNN
// where NNNNNNNNN combines nanoseconds, random number, and counter for uniqueness
func (om *OutputManager) GenerateTimestampedDirName() string {
	// Get current time
	now := time.Now()

	// Generate timestamp in format YYYYMMDD_HHMMSS
	timestamp := now.Format("20060102_150405")

	// Use multiple strategies to ensure uniqueness:
	// 1. Nanoseconds from the current time (0-999999999)
	// 2. Random number (0-999)
	// 3. Incremental counter (0-999999)

	// Extract nanoseconds (last 3 digits)
	nanos := now.Nanosecond() % 1000

	// Generate a random number
	rng := rand.New(rand.NewSource(now.UnixNano()))
	randNum := rng.Intn(1000)

	// Increment the counter atomically (thread-safe)
	counter := atomic.AddUint32(&incrementalCounter, 1) % 1000

	// Combine all three components for a truly unique value
	// This gives us a billion possibilities (1000 × 1000 × 1000) within the same second
	uniqueNum := (nanos * 1000000) + (randNum * 1000) + int(counter)

	// Combine with prefix and format with leading zeros (9 digits)
	return fmt.Sprintf("thinktank_%s_%09d", timestamp, uniqueNum)
}

// GenerateMemorableDirName generates a three-word memorable directory name.
// Format: adjective-verb-noun (lowercase, hyphen-separated).
func (om *OutputManager) GenerateMemorableDirName() string {
	if om.memorableTotal == 0 {
		return fmt.Sprintf("%s-%s-%s", adjectives[0], verbs[0], nouns[0])
	}

	for i := 0; i < memorableNameAttempts; i++ {
		idx := om.nextMemorableIndex()
		adj, verb, noun := om.memorableParts(idx)
		name := fmt.Sprintf("%s-%s-%s", adj, verb, noun)
		if len(name) >= memorableNameMinLength && len(name) <= memorableNameMaxLength {
			return name
		}
	}

	if name, ok := firstMemorableNameWithinLimits(); ok {
		return name
	}

	return fmt.Sprintf("%s-%s-%s", adjectives[0], verbs[0], nouns[0])
}

func (om *OutputManager) nextMemorableIndex() int {
	seq := int64(om.memorableCounter.Add(1) - 1)
	idx := int64(om.memorableOffset) + (seq * int64(om.memorableStride))
	if om.memorableTotal == 0 {
		return 0
	}
	return int(idx % int64(om.memorableTotal))
}

func (om *OutputManager) memorableParts(index int) (string, string, string) {
	nounIndex := index % len(nouns)
	verbIndex := (index / len(nouns)) % len(verbs)
	adjIndex := (index / (len(nouns) * len(verbs))) % len(adjectives)
	return adjectives[adjIndex], verbs[verbIndex], nouns[nounIndex]
}

func firstMemorableNameWithinLimits() (string, bool) {
	for _, adj := range adjectives {
		for _, verb := range verbs {
			for _, noun := range nouns {
				name := fmt.Sprintf("%s-%s-%s", adj, verb, noun)
				if len(name) >= memorableNameMinLength && len(name) <= memorableNameMaxLength {
					return name, true
				}
			}
		}
	}
	return "", false
}

func makeWordSet(words []string) map[string]struct{} {
	set := make(map[string]struct{}, len(words))
	for _, word := range words {
		set[word] = struct{}{}
	}
	return set
}

func pickCoprimeStride(rng *rand.Rand, total int) int {
	if total <= 2 {
		return 1
	}

	for {
		stride := rng.Intn(total-1) + 1
		if gcd(stride, total) == 1 {
			return stride
		}
	}
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	if a < 0 {
		return -a
	}
	return a
}

// CreateOutputDirectory creates an output directory with collision detection
// If basePath is empty, uses current working directory
func (om *OutputManager) CreateOutputDirectory(basePath string, permissions os.FileMode) (string, error) {
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			om.logger.Printf("Error getting current working directory: %v", err)
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = cwd
	}

	// Generate unique directory name
	dirName, err := om.generateAvailableMemorableDirName(basePath, maxCollisionAttempts)
	if err != nil {
		om.logger.Printf("Memorable name collisions exhausted, falling back to timestamp: %v", err)
		dirName, err = om.generateAvailableTimestampDirName(basePath, maxCollisionAttempts)
		if err != nil {
			return "", err
		}
	}
	fullPath := filepath.Join(basePath, dirName)

	// Create the directory with proper permissions (0755)
	if err := os.MkdirAll(fullPath, permissions); err != nil {
		om.logger.Printf("Error creating output directory %s: %v", fullPath, err)
		return "", fmt.Errorf("failed to create output directory %s: %w", fullPath, err)
	}

	om.logger.Printf("Created output directory: %s", fullPath)
	return fullPath, nil
}

func (om *OutputManager) generateAvailableMemorableDirName(basePath string, maxAttempts int) (string, error) {
	for attempt := 0; attempt < maxAttempts; attempt++ {
		dirName := om.GenerateMemorableDirName()
		fullPath := filepath.Join(basePath, dirName)
		exists, err := dirExists(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to check directory %s: %w", fullPath, err)
		}
		if !exists {
			return dirName, nil
		}
		om.logger.Printf("Directory collision detected, trying: %s", dirName)
	}

	return "", fmt.Errorf("failed to find unique memorable directory name after %d attempts", maxAttempts)
}

func (om *OutputManager) generateAvailableTimestampDirName(basePath string, maxAttempts int) (string, error) {
	dirName := om.GenerateTimestampedDirName()
	fullPath := filepath.Join(basePath, dirName)
	exists, err := dirExists(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to check directory %s: %w", fullPath, err)
	}
	if !exists {
		return dirName, nil
	}

	originalName := dirName
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		retryName := fmt.Sprintf("%s_retry%d", originalName, attempt)
		fullPath = filepath.Join(basePath, retryName)
		exists, err = dirExists(fullPath)
		if err != nil {
			return "", fmt.Errorf("failed to check directory %s: %w", fullPath, err)
		}
		if !exists {
			return retryName, nil
		}
		om.logger.Printf("Directory collision detected, trying: %s", retryName)
	}

	return "", fmt.Errorf("failed to find unique timestamp directory name after %d attempts", maxAttempts)
}

func dirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CleanupOldDirectories removes output directories older than the specified duration
// It only removes directories that match the thinktank naming pattern
func (om *OutputManager) CleanupOldDirectories(basePath string, olderThan time.Duration) error {
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = cwd
	}

	cutoffTime := time.Now().Add(-olderThan)
	om.logger.Printf("Cleaning up directories older than %v (before %v)", olderThan, cutoffTime.Format(time.RFC3339))

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	cleanedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Only process directories that match our naming pattern
		if !om.isThinktankOutputDir(entry.Name()) {
			continue
		}

		dirPath := filepath.Join(basePath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			om.logger.Printf("Warning: Cannot get info for directory %s: %v", dirPath, err)
			continue
		}

		// Check if directory is older than cutoff
		if info.ModTime().Before(cutoffTime) {
			if err := os.RemoveAll(dirPath); err != nil {
				om.logger.Printf("Warning: Failed to remove old directory %s: %v", dirPath, err)
				continue
			}

			om.logger.Printf("Removed old output directory: %s (last modified: %v)",
				entry.Name(), info.ModTime().Format(time.RFC3339))
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		om.logger.Printf("Cleanup completed: removed %d old directories", cleanedCount)
	} else {
		om.logger.Printf("Cleanup completed: no old directories found")
	}

	return nil
}

// isThinktankOutputDir checks if a directory name matches the thinktank output pattern
func (om *OutputManager) isThinktankOutputDir(name string) bool {
	return om.isTimestampedOutputDir(name) || om.isMemorableOutputDir(name)
}

func (om *OutputManager) isTimestampedOutputDir(name string) bool {
	baseName := name
	if idx := strings.Index(name, "_retry"); idx != -1 {
		retryPart := name[idx+len("_retry"):]
		if retryPart == "" {
			return false
		}
		for _, r := range retryPart {
			if r < '0' || r > '9' {
				return false
			}
		}
		baseName = name[:idx]
	}

	// Check if it starts with "thinktank_" and has the expected format
	if !strings.HasPrefix(baseName, "thinktank_") {
		return false
	}

	// Split by underscores and check structure
	parts := strings.Split(baseName, "_")
	if len(parts) != 4 {
		return false
	}

	// Verify the date part (YYYYMMDD)
	dateStr := parts[1]
	if len(dateStr) != 8 {
		return false
	}

	// Verify the time part (HHMMSS)
	timeStr := parts[2]
	if len(timeStr) != 6 {
		return false
	}

	// Verify the unique part (NNNNNNNNN)
	uniquePart := parts[3]
	return len(uniquePart) == 9
}

func (om *OutputManager) isMemorableOutputDir(name string) bool {
	if len(name) < memorableNameMinLength || len(name) > memorableNameMaxLength {
		return false
	}

	parts := strings.Split(name, "-")
	if len(parts) != 3 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if r < 'a' || r > 'z' {
				return false
			}
		}
	}

	if _, ok := adjectiveSet[parts[0]]; !ok {
		return false
	}
	if _, ok := verbSet[parts[1]]; !ok {
		return false
	}
	if _, ok := nounSet[parts[2]]; !ok {
		return false
	}

	return true
}

// GetOutputDirStats returns statistics about output directories in the base path
func (om *OutputManager) GetOutputDirStats(basePath string) (map[string]interface{}, error) {
	if basePath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		basePath = cwd
	}

	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", basePath, err)
	}

	stats := map[string]interface{}{
		"total_dirs":       0,
		"thinktank_dirs":   0,
		"oldest_dir":       "",
		"newest_dir":       "",
		"total_size_bytes": int64(0),
	}

	var oldestTime, newestTime time.Time
	var oldestDir, newestDir string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		stats["total_dirs"] = stats["total_dirs"].(int) + 1

		if om.isThinktankOutputDir(entry.Name()) {
			stats["thinktank_dirs"] = stats["thinktank_dirs"].(int) + 1

			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime()
			if oldestTime.IsZero() || modTime.Before(oldestTime) {
				oldestTime = modTime
				oldestDir = entry.Name()
			}
			if newestTime.IsZero() || modTime.After(newestTime) {
				newestTime = modTime
				newestDir = entry.Name()
			}

			// Calculate directory size
			dirPath := filepath.Join(basePath, entry.Name())
			if size, err := om.calculateDirSize(dirPath); err == nil {
				stats["total_size_bytes"] = stats["total_size_bytes"].(int64) + size
			}
		}
	}

	stats["oldest_dir"] = oldestDir
	stats["newest_dir"] = newestDir

	return stats, nil
}

// calculateDirSize calculates the total size of a directory
func (om *OutputManager) calculateDirSize(dirPath string) (int64, error) {
	var size int64

	err := filepath.WalkDir(dirPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}

		if !d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return nil // Skip errors and continue
			}
			size += info.Size()
		}

		return nil
	})

	return size, err
}

// CleanupOldDirectoriesWithDefault performs cleanup with default 30-day retention
func (om *OutputManager) CleanupOldDirectoriesWithDefault(basePath string) error {
	return om.CleanupOldDirectories(basePath, 30*24*time.Hour)
}
