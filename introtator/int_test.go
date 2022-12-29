package introtator_test

import (
	"fmt"
	"path"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr/filer"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/mocks"
)

var errTest = fmt.Errorf("this is a test error")

func TestPost(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	layout := &introtator.Layout{PostRotate: func(s1, s2 string) {
		assert.Equal("string1", s1)
		assert.Equal("string2", s2)
	}}
	layout.Post("string1", "string2")

	layout.PostRotate = nil
	layout.Post("string1", "string2")
}

func TestDirs(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	// test archive dir.
	layout := &introtator.Layout{ArchiveDir: "/var/log/archives"}
	dirs, err := layout.Dirs("/var/log/service.log")
	assert.Equal([]string{filepath.Join("/", "var", "log"), path.Join("/", "var", "log", "archives")},
		dirs, "the wrong directories were returned")
	assert.Nil(err, "this should not producce an error")
	assert.EqualValues(filer.Default(), layout.Filer)
	assert.Equal(layout.FileOrder, introtator.Ascending)

	// test invalid file order.
	layout = &introtator.Layout{FileOrder: 99}
	dirs, err = layout.Dirs("/var/log/service.log")
	assert.Equal([]string{filepath.Join("/", "var", "log")}, dirs, "the wrong directory was returned")
	assert.Nil(err, "this should not producce an error")
	assert.Equal(layout.FileOrder, introtator.Ascending)

	// test valid file order.
	layout = &introtator.Layout{FileOrder: introtator.Descending}
	dirs, err = layout.Dirs("/var/log/service.log")
	assert.Equal([]string{filepath.Join("/", "var", "log")}, dirs, "the wrong directory was returned")
	assert.Nil(err, "this should not producce an error")
	assert.Equal(layout.FileOrder, introtator.Descending)
}

func TestRotateFirst(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	layout := &introtator.Layout{ArchiveDir: "/var/log/archives", Filer: mockFiler}

	// Basic test representing first rotate, ascending & descending (first rotate is the same).
	mockFiler.EXPECT().ReadDir(layout.ArchiveDir).Times(2)
	mockFiler.EXPECT().Rename(path.Join("/", "var", "log", "service.log"),
		filepath.Join(layout.ArchiveDir, "service.1.log")).Times(2)
	//
	file, err := layout.Rotate("/var/log/service.log")
	assert.Equal(filepath.Join(layout.ArchiveDir, "service.1.log"), file)
	assert.Nil(err)
	//
	layout.FileOrder = introtator.Descending
	file, err = layout.Rotate("/var/log/service.log")
	assert.Equal(filepath.Join(layout.ArchiveDir, "service.1.log"), file)
	assert.Nil(err)

	// Test a couple errors.
	mockFiler.EXPECT().ReadDir(layout.ArchiveDir).Times(2)
	mockFiler.EXPECT().Rename(filepath.Join("/", "var", "log", "service.log"),
		filepath.Join(layout.ArchiveDir, "service.1.log")).Times(2).Return(errTest)
	//
	file, err = layout.Rotate("/var/log/service.log")
	assert.Empty(file, "the file must be empty when rotation fails.")
	assert.ErrorIs(err, errTest, "the rename error must be returned.")
	//
	layout.FileOrder = introtator.Ascending
	file, err = layout.Rotate("/var/log/service.log")
	assert.Empty(file, "the file must be empty when rotation fails.")
	assert.ErrorIs(err, errTest, "the rename error must be returned.")
}
