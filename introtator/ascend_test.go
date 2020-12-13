package introtator_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/mocks"
)

func testFakeFiles(mockCtrl *gomock.Controller, count int) (fakes []*mocks.MockFileInfo, files []os.FileInfo) {
	for i := 0; i < count; i++ {
		fake := mocks.NewMockFileInfo(mockCtrl)
		fakes = append(fakes, fake)
		files = append(files, fake)
	}

	return fakes, files
}

func TestRotateAsc(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	l := &introtator.Layout{
		Filer:     mockFiler,
		FileOrder: introtator.Ascending,
		FileCount: 5,
	}

	// Simple test to start, rotate 1 file.
	mockFiler.EXPECT().ReadDir("/var/log")
	mockFiler.EXPECT().Rename("/var/log/service.log", "/var/log/service.1.log")
	//
	file, err := l.Rotate("/var/log/service.log")
	assert.Equal("/var/log/service.1.log", file)
	assert.Nil(err)

	// Make sure files rotate correctly.. we have some extras to delete too.
	fakes, fakeFiles := testFakeFiles(mockCtrl, 10)
	gomock.InOrder(
		mockFiler.EXPECT().ReadDir("/var/log").Return(fakeFiles, nil),
		mockFiler.EXPECT().Rename("/var/log/service.10.log.gz", "/var/log/service.11.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.9.log.gz", "/var/log/service.10.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.8.log.gz", "/var/log/service.9.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.7.log.gz", "/var/log/service.8.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.6.log.gz", "/var/log/service.7.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.5.log.gz", "/var/log/service.6.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.4.log.gz", "/var/log/service.5.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.3.log.gz", "/var/log/service.4.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.2.log.gz", "/var/log/service.3.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.1.log.gz", "/var/log/service.2.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.log", "/var/log/service.1.log"), // no gz.
		// We should only have 5 backup log files.
		mockFiler.EXPECT().Remove("/var/log/service.11.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.10.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.9.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.8.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.7.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.6.log.gz"),
	)
	//
	for i := range fakes {
		fakes[i].EXPECT().Name().Return("service." + strconv.Itoa(i+1) + ".log.gz")
	}
	//
	file, err = l.Rotate("/var/log/service.log")
	assert.Equal("/var/log/service.1.log", file)
	assert.Nil(err)
}
