package introtator_test

import (
	"path/filepath"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/mocks"
)

func TestRotateDesc(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	layout := &introtator.Layout{
		Filer:     mockFiler,
		FileOrder: introtator.Descending,
		FileCount: 5,
	}

	// Simple test to start, rotate 1 file.
	mockFiler.EXPECT().ReadDir(filepath.Join("/", "var", "log"))
	mockFiler.EXPECT().Rename("/var/log/service.log", filepath.Join("/", "var", "log", "service.1.log"))
	//
	file, err := layout.Rotate("/var/log/service.log")
	assert.Equal("/var/log/service.1.log", file)
	assert.Nil(err)

	// Make sure files rotate correctly.. we have some extras to delete too.
	fakes, fakeFiles := testFakeFiles(mockCtrl, 10)
	gomock.InOrder(
		mockFiler.EXPECT().ReadDir("/var/log").Return(fakeFiles, nil),
		// We should only have 5 backup log files.
		mockFiler.EXPECT().Remove("/var/log/service.1.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.2.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.3.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.4.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.5.log.gz"),
		mockFiler.EXPECT().Remove("/var/log/service.6.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.7.log.gz", "/var/log/service.1.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.8.log.gz", "/var/log/service.2.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.9.log.gz", "/var/log/service.3.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.10.log.gz", "/var/log/service.4.log.gz"),
		mockFiler.EXPECT().Rename("/var/log/service.log", "/var/log/service.5.log"), // no gz.
	)
	//
	for i := range fakes {
		fakes[i].EXPECT().Name().Return("service." + strconv.Itoa(i+1) + ".log.gz")
	}
	//
	file, err = layout.Rotate("/var/log/service.log")
	assert.Equal("/var/log/service.5.log", file)
	assert.Nil(err)

	// Make sure a delete failure returns an error.
	gomock.InOrder(
		mockFiler.EXPECT().ReadDir("/var/log").Return(fakeFiles, nil),
		mockFiler.EXPECT().Remove("/var/log/service.1.log.gz").Return(errTest),
	)
	//
	for i := range fakes {
		fakes[i].EXPECT().Name().Return("service." + strconv.Itoa(i+1) + ".log.gz")
	}
	//
	file, err = layout.Rotate("/var/log/service.log")
	assert.Empty(file)
	assert.ErrorIs(err, errTest)
}
