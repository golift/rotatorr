package introtator_test

import (
	"fmt"
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

	l := &introtator.Layout{PostRotate: func(s1, s2 string) {
		assert.Equal("string1", s1)
		assert.Equal("string2", s2)
	}}
	l.Post("string1", "string2")

	l.PostRotate = nil
	l.Post("string1", "string2")
}

func TestDirs(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	// test archive dir.
	l := &introtator.Layout{ArchiveDir: "/var/log/archives"}
	f, err := l.Dirs("/var/log/service.log")
	assert.Equal([]string{"/var/log", "/var/log/archives"}, f, "the wrong directories were returned")
	assert.Nil(err, "this should not producce an error")
	assert.EqualValues(filer.Default(), l.Filer)
	assert.Equal(l.FileOrder, introtator.Ascending)

	// test invalid file order.
	l = &introtator.Layout{FileOrder: 99}
	f, err = l.Dirs("/var/log/service.log")
	assert.Equal([]string{"/var/log"}, f, "the wrong directory was returned")
	assert.Nil(err, "this should not producce an error")
	assert.Equal(l.FileOrder, introtator.Ascending)

	// test valid file order.
	l = &introtator.Layout{FileOrder: introtator.Descending}
	f, err = l.Dirs("/var/log/service.log")
	assert.Equal([]string{"/var/log"}, f, "the wrong directory was returned")
	assert.Nil(err, "this should not producce an error")
	assert.Equal(l.FileOrder, introtator.Descending)
}

func TestRotateFirst(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	l := &introtator.Layout{ArchiveDir: "/var/log/archives", Filer: mockFiler}

	// Basic test representing first rotate, ascending & descending (first rotate is the same).
	mockFiler.EXPECT().ReadDir(l.ArchiveDir).Times(2)
	mockFiler.EXPECT().Rename("/var/log/service.log", l.ArchiveDir+"/service.1.log").Times(2)
	//
	file, err := l.Rotate("/var/log/service.log")
	assert.Equal(l.ArchiveDir+"/service.1.log", file)
	assert.Nil(err)
	//
	l.FileOrder = introtator.Descending
	file, err = l.Rotate("/var/log/service.log")
	assert.Equal(l.ArchiveDir+"/service.1.log", file)
	assert.Nil(err)

	// Test a couple errors.
	mockFiler.EXPECT().ReadDir(l.ArchiveDir).Times(2)
	mockFiler.EXPECT().Rename("/var/log/service.log", l.ArchiveDir+"/service.1.log").Times(2).Return(errTest)
	//
	file, err = l.Rotate("/var/log/service.log")
	assert.Empty(file, "the file must be empty when rotation fails.")
	assert.ErrorIs(err, errTest, "the rename error must be returned.")
	//
	l.FileOrder = introtator.Ascending
	file, err = l.Rotate("/var/log/service.log")
	assert.Empty(file, "the file must be empty when rotation fails.")
	assert.ErrorIs(err, errTest, "the rename error must be returned.")
}
