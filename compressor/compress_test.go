package compressor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr/compressor"
)

// pretty simple test. more can be done by mocking Filer.
func TestCompress(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	compressor.CompressLevel = 77

	r, err := compressor.Compress("/does/not/exist/file")
	assert.NotNil(err)
	assert.Contains(err.Error(), "stating old file:")
	assert.ErrorIs(err, r.Error)

	dir := os.TempDir()
	err = os.MkdirAll(dir, 0755)
	assert.Nilf(err, "error creating test dir: %v", err)
	f, err := os.Create(filepath.Join(dir, "testfile.log"))
	assert.Nilf(err, "error creating test file: %v", err)
	_, err = f.Write(make([]byte, 300000))
	assert.Nilf(err, "error writing test file: %v", err)
	r, err = compressor.Compress(f.Name())
	assert.Nil(err)
	assert.Nil(r.Error)

	// XXX: check report items.
	os.Remove(f.Name())
}
