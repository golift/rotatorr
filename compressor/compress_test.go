package compressor_test

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golift.io/rotatorr/compressor"
	"golift.io/rotatorr/filer"
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

type statCountFiler struct {
	filer.Filer

	n *atomic.Int32
}

func (s *statCountFiler) Stat(name string) (*filer.FileInfo, error) {
	s.n.Add(1)

	return s.Filer.Stat(name)
}

func TestCompressEmptyFileName(t *testing.T) { //nolint:paralleltest
	orig := compressor.Filer

	t.Cleanup(func() { compressor.Filer = orig })

	var stats atomic.Int32

	compressor.Filer = &statCountFiler{Filer: orig, n: &stats}

	report, err := compressor.Compress("")
	require.NoError(t, err)
	require.NotNil(t, report)
	require.Empty(t, report.OldFile)

	compressor.CompressWithLog("", nil)
	compressor.CompressBackgroundWithLog("", nil)
	compressor.CompressPostRotate("/unused", "")
	compressor.CompressBackgroundPostRotate("/unused", "")

	called := make(chan struct{})

	compressor.CompressBackground("", func(*compressor.Report) { close(called) })

	select {
	case <-called:
		t.Fatal("empty fileName must not start a compression")
	case <-time.After(150 * time.Millisecond):
	}

	require.Zero(t, stats.Load(), "empty fileName must not Stat")
}
