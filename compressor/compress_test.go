package compressor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/rotatorr/compressor"
)

// pretty simple test. more can be done by mocking Filer.
func TestCompress(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	compressor.CompressLevel = 77

	report, err := compressor.Compress("/does/not/exist/file")
	require.Error(t, err)
	assert.Contains(err.Error(), "stating old file:")
	require.ErrorIs(t, err, report.Error)

	dir := os.TempDir()
	err = os.MkdirAll(dir, 0o750)
	require.NoError(t, err, "error creating test dir: %v", err)
	oFile, err := os.Create(filepath.Join(dir, "testfile.log"))
	require.NoError(t, err, "error creating test file: %v", err)
	_, err = oFile.Write(make([]byte, 300000))
	require.NoError(t, err, "error writing test file: %v", err)
	report, err = compressor.Compress(oFile.Name())
	require.NoError(t, err)
	require.NoError(t, report.Error)

	// XXX: check report items.
	_ = os.Remove(oFile.Name())
}
