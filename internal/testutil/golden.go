package testutil

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

// Golden provides utilities for golden file testing
type Golden struct {
	t       *testing.T
	baseDir string
}

// New creates a new Golden test helper
func New(t *testing.T, baseDir string) *Golden {
	return &Golden{
		t:       t,
		baseDir: baseDir,
	}
}

// Assert compares the actual output with the golden file
// If the -update flag is set, it updates the golden file with the actual output
func (g *Golden) Assert(actual string, goldenFile string) {
	g.t.Helper()

	goldenPath := filepath.Join(g.baseDir, goldenFile)

	if *update {
		g.update(goldenPath, actual)
		return
	}

	expected := g.load(goldenPath)
	require.Equal(g.t, expected, actual, "Output does not match golden file: %s\nRun with -update flag to update golden files", goldenFile)
}

// AssertBytes compares byte output with the golden file
func (g *Golden) AssertBytes(actual []byte, goldenFile string) {
	g.Assert(string(actual), goldenFile)
}

// load reads the golden file content
func (g *Golden) load(path string) string {
	g.t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(g.t, err, "Failed to read golden file: %s", path)

	return string(data)
}

// update writes the actual content to the golden file
func (g *Golden) update(path string, content string) {
	g.t.Helper()

	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	require.NoError(g.t, err, "Failed to create golden file directory: %s", dir)

	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(g.t, err, "Failed to update golden file: %s", path)

	g.t.Logf("Updated golden file: %s", path)
}

// Path returns the full path to a golden file
func (g *Golden) Path(goldenFile string) string {
	return filepath.Join(g.baseDir, goldenFile)
}

// Exists checks if a golden file exists
func (g *Golden) Exists(goldenFile string) bool {
	_, err := os.Stat(g.Path(goldenFile))
	return err == nil
}

// ShouldUpdate returns true if golden files should be updated
func ShouldUpdate() bool {
	return *update
}

// LoadFixture loads a test fixture file
func LoadFixture(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err, "Failed to load fixture: %s", path)

	return data
}

// LoadFixtureString loads a test fixture file as a string
func LoadFixtureString(t *testing.T, path string) string {
	return string(LoadFixture(t, path))
}
