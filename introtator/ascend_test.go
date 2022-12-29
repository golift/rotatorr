package introtator_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/mocks"
)

func testFakeFiles(mockCtrl *gomock.Controller, count int) ([]*mocks.MockFileInfo, []os.FileInfo) {
	var (
		fakes = make([]*mocks.MockFileInfo, count)
		files = make([]os.FileInfo, count)
	)

	for i := 0; i < count; i++ {
		fake := mocks.NewMockFileInfo(mockCtrl)
		fakes[i] = fake
		files[i] = fake
	}

	return fakes, files
}

func TestRotateAsc(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	layout := &introtator.Layout{
		Filer:     mockFiler,
		FileOrder: introtator.Ascending,
		FileCount: 5,
	}

	// Simple test to start, rotate 1 file.
	mockFiler.EXPECT().ReadDir(filepath.Join("/", "var", "log"))
	mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.log"),
		filepath.Join("/", "var", "log", "service.1.log"))
	//
	file, err := layout.Rotate(filepath.Join("/", "var", "log", "service.log"))
	assert.Equal(filepath.Join("/", "var", "log", "service.1.log"), file)
	assert.Nil(err)

	// Make sure files rotate correctly.. we have some extras to delete too.
	fakes, fakeFiles := testFakeFiles(mockCtrl, 10)
	gomock.InOrder(
		mockFiler.EXPECT().ReadDir(filepath.Join("/", "var", "log")).Return(fakeFiles, nil),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.10.log.gz"),
			filepath.Join("/", "var", "log", "service.11.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.9.log.gz"),
			filepath.Join("/", "var", "log", "service.10.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.8.log.gz"),
			filepath.Join("/", "var", "log", "service.9.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.7.log.gz"),
			filepath.Join("/", "var", "log", "service.8.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.6.log.gz"),
			filepath.Join("/", "var", "log", "service.7.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.5.log.gz"),
			filepath.Join("/", "var", "log", "service.6.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.4.log.gz"),
			filepath.Join("/", "var", "log", "service.5.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.3.log.gz"),
			filepath.Join("/", "var", "log", "service.4.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.2.log.gz"),
			filepath.Join("/", "var", "log", "service.3.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.1.log.gz"),
			filepath.Join("/", "var", "log", "service.2.log.gz")),
		mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.log"),
			filepath.Join("/", "var", "log", "service.1.log")), // no gz.
		// We should only have 5 backup log files.
		mockFiler.EXPECT().Remove(filepath.Join("/", "var", "log", "service.11.log.gz")),
		mockFiler.EXPECT().Remove(filepath.Join("/", "var", "log", "service.10.log.gz")),
		mockFiler.EXPECT().Remove(filepath.Join("/", "var", "log", "service.9.log.gz")),
		mockFiler.EXPECT().Remove(filepath.Join("/", "var", "log", "service.8.log.gz")),
		mockFiler.EXPECT().Remove(filepath.Join("/", "var", "log", "service.7.log.gz")),
		mockFiler.EXPECT().Remove(filepath.Join("/", "var", "log", "service.6.log.gz")),
	)
	//
	for i := range fakes {
		fakes[i].EXPECT().Name().Return("service." + strconv.Itoa(i+1) + ".log.gz")
	}
	//
	file, err = layout.Rotate(filepath.Join("/", "var", "log", "service.log"))
	assert.Equal(filepath.Join("/", "var", "log", "service.1.log"), file)
	assert.Nil(err)
}
