package rotatorr_test

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
	"golift.io/rotatorr"
	"golift.io/rotatorr/filer"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/mocks"
)

// Basic run of the mill usage. Hits 85% of the code just doing normal things.
func TestNew(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	logger := rotatorr.NewMust(&rotatorr.Config{
		FileSize: 50,
		Rotatorr: &introtator.Layout{},
	})

	log.SetOutput(logger)
	log.Println("weeeeeeeee!")
	log.Println("weee!")

	err := log.Output(1, "weeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee!")
	require.ErrorIs(t, err, rotatorr.ErrWriteTooLarge)
	//
	_, err = logger.Rotate()
	require.NoError(t, err)
	assert.NoError(logger.Close())
}

func TestRotateSize(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRotatorr := mocks.NewMockRotatorr(mockCtrl)
	testFile, err := os.CreateTemp(t.TempDir(), "*.log")
	assert.NoError(err, "problem creating temp file")
	assert.NoError(testFile.Close(), "problem closing temp file")
	mockRotatorr.EXPECT().Dirs(gomock.Any())
	//
	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: testFile.Name(),
		FileSize: 50,
		Rotatorr: mockRotatorr,
	})
	if err != nil {
		assert.NoError(err)
		return
	}

	defer logger.Close() // release file handle so t.TempDir() cleanup can remove files on Windows

	//
	msg := "log message"                                        // len: 11
	s, err := logger.Write([]byte(msg + msg + msg + msg + msg)) // len: 55
	require.ErrorIs(t, err, rotatorr.ErrWriteTooLarge, "writing more data than our filesize must produce an error")
	assert.Equal(0, s, "size must be 0 if the write fails.")

	check := func(s int, err error) {
		require.NoError(t, err)
		assert.Equal(len(msg), s)
	}

	check(logger.Write([]byte(msg))) // 11
	check(logger.Write([]byte(msg))) // 22
	check(logger.Write([]byte(msg))) // 33
	check(logger.Write([]byte(msg))) // 44
	mockRotatorr.EXPECT().Rotate(testFile.Name())
	check(logger.Write([]byte(msg))) // 55 > 50, rotate!
}

func TestRotateEvery(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRotatorr := mocks.NewMockRotatorr(mockCtrl)
	testFile, err := os.CreateTemp(t.TempDir(), "*.log")
	assert.NoError(err, "problem creating temp file")
	assert.NoError(testFile.Close(), "problem closing temp file")
	mockRotatorr.EXPECT().Dirs(gomock.Any())
	//

	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: testFile.Name(),
		Every:    time.Second,
		Rotatorr: mockRotatorr,
	})
	if err != nil {
		assert.NoError(err)
		return
	}

	defer logger.Close() // release file handle so t.TempDir() cleanup can remove files on Windows
	//
	msg := "log message"                                        // len: 11
	s, err := logger.Write([]byte(msg + msg + msg + msg + msg)) // len: 55
	require.NoError(t, err)
	assert.Equal(len(msg)*5, s)

	check := func(s int, err error) {
		require.NoError(t, err)
		assert.Equal(len(msg), s)
	}
	check(logger.Write([]byte(msg))) // 11
	check(logger.Write([]byte(msg))) // 22
	time.Sleep(time.Second)
	mockRotatorr.EXPECT().Rotate(testFile.Name())
	check(logger.Write([]byte(msg))) // 33
}

func TestReopen(t *testing.T) {
	t.Parallel()
	testReopen(t, false)
}

func TestReopenAfterRename(t *testing.T) {
	t.Parallel()

	if runtime.GOOS == "windows" {
		t.Skip("Windows cannot rename a log file that is still open")
	}

	testReopen(t, true)
}

func testReopen(t *testing.T, renameFirst bool) {
	t.Helper()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRotatorr := mocks.NewMockRotatorr(mockCtrl)
	path := filepath.Join(t.TempDir(), "service.log")
	testFile, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, testFile.Close())

	mockRotatorr.EXPECT().Dirs(gomock.Any())

	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: path,
		FileSize: rotatorr.NoMaxSize,
		Rotatorr: mockRotatorr,
	})
	require.NoError(t, err)

	defer logger.Close() // release file handle so t.TempDir() cleanup can remove files on Windows

	msg1 := []byte("before reopen\n")
	wrote, err := logger.Write(msg1)
	require.NoError(t, err)
	assert.Equal(t, len(msg1), wrote)

	rotated := path + ".1"
	if renameFirst {
		require.NoError(t, os.Rename(path, rotated))
	}

	require.NoError(t, logger.Reopen())

	msg2 := []byte("after reopen\n")
	wrote, err = logger.Write(msg2)
	require.NoError(t, err)
	assert.Equal(t, len(msg2), wrote)

	if !renameFirst {
		got, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, string(msg1)+string(msg2), string(got))

		return
	}

	gotOld, err := os.ReadFile(rotated)
	require.NoError(t, err)
	assert.Equal(t, string(msg1), string(gotOld))

	gotNew, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(msg2), string(gotNew))
}

type statErrFiler struct {
	filer.Filer

	err error
}

func (s *statErrFiler) Stat(_ string) (*filer.FileInfo, error) {
	return nil, s.err
}

func TestReopenStatErrorDoesNotTruncate(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRotatorr := mocks.NewMockRotatorr(mockCtrl)
	path := filepath.Join(t.TempDir(), "service.log")
	require.NoError(t, os.WriteFile(path, []byte("keep me\n"), 0o600))
	mockRotatorr.EXPECT().Dirs(gomock.Any())

	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: path,
		FileSize: rotatorr.NoMaxSize,
		Rotatorr: mockRotatorr,
	})
	require.NoError(t, err)

	defer logger.Close()

	_, err = logger.Write([]byte("keep me too\n"))
	require.NoError(t, err)

	logger.Filer = &statErrFiler{Filer: filer.Default(), err: fs.ErrPermission}

	err = logger.Reopen()
	require.ErrorIs(t, err, fs.ErrPermission)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(got), "keep me\n")
	assert.Contains(t, string(got), "keep me too\n")
}
