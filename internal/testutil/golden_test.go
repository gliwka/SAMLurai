package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGolden_Assert(t *testing.T) {
	// Create a temp directory for golden files
	tmpDir := t.TempDir()
	goldenDir := filepath.Join(tmpDir, "golden")

	// Create the golden file first
	err := os.MkdirAll(goldenDir, 0755)
	require.NoError(t, err)

	goldenContent := "expected output"
	goldenPath := filepath.Join(goldenDir, "test.golden")
	err = os.WriteFile(goldenPath, []byte(goldenContent), 0644)
	require.NoError(t, err)

	// Test assertion
	golden := New(t, goldenDir)
	golden.Assert(goldenContent, "test.golden")
}

func TestGolden_Path(t *testing.T) {
	golden := New(t, "/base/dir")
	path := golden.Path("subdir/file.golden")
	assert.Equal(t, "/base/dir/subdir/file.golden", path)
}

func TestGolden_Exists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tmpDir, "exists.golden")
	err := os.WriteFile(testFile, []byte("content"), 0644)
	require.NoError(t, err)

	golden := New(t, tmpDir)

	assert.True(t, golden.Exists("exists.golden"))
	assert.False(t, golden.Exists("notexists.golden"))
}

func TestLoadFixture(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fixture file
	fixtureContent := "fixture content"
	fixturePath := filepath.Join(tmpDir, "test.fixture")
	err := os.WriteFile(fixturePath, []byte(fixtureContent), 0644)
	require.NoError(t, err)

	// Test loading
	loaded := LoadFixture(t, fixturePath)
	assert.Equal(t, []byte(fixtureContent), loaded)
}

func TestLoadFixtureString(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a fixture file
	fixtureContent := "fixture content as string"
	fixturePath := filepath.Join(tmpDir, "test.fixture")
	err := os.WriteFile(fixturePath, []byte(fixtureContent), 0644)
	require.NoError(t, err)

	// Test loading as string
	loaded := LoadFixtureString(t, fixturePath)
	assert.Equal(t, fixtureContent, loaded)
}
