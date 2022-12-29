package timerotator_test

import (
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr/filer"
	"golift.io/rotatorr/mocks"
	"golift.io/rotatorr/timerotator"
)

func TestPost(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	layout := &timerotator.Layout{PostRotate: func(s1, s2 string) {
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
	layout := &timerotator.Layout{ArchiveDir: "/var/log/archives"}
	f, err := layout.Dirs("/var/log/service.log")
	assert.Equal([]string{"/var/log", "/var/log/archives"}, f, "the wrong directories were returned")
	assert.Nil(err, "this should not producce an error")
	assert.EqualValues(filer.Default(), layout.Filer)
	assert.Equal(layout.Joiner, timerotator.DefaultJoiner)
	assert.Equal(layout.Format, timerotator.FormatDefault)
}

func TestRotateOne(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	layout := &timerotator.Layout{
		Filer:  mockFiler,
		UseUTC: true,
		Format: timerotator.FormatNoSecnd,
		Joiner: timerotator.DefaultJoiner,
	}
	newName := "/var/log/service" + layout.Joiner + time.Now().UTC().Format(layout.Format) + ".log"

	// Basic test representing first rotate (no existing files).
	mockFiler.EXPECT().ReadDir("/var/log")
	mockFiler.EXPECT().Rename("/var/log/service.log", newName)
	//
	file, err := layout.Rotate("/var/log/service.log")
	assert.Equal(newName, file)
	assert.Nil(err)
}

// Make fake files to fake delete.
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

func TestRotateDelete(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	fakes, fakeFiles := testFakeFiles(mockCtrl, 10)
	layout := &timerotator.Layout{
		ArchiveDir: "/var/log/archives",
		Filer:      mockFiler,
		UseUTC:     true,
		Format:     timerotator.FormatNoSecnd,
		Joiner:     timerotator.DefaultJoiner,
		FileAge:    time.Minute,
		FileCount:  2,
	}
	newName := layout.ArchiveDir + "/service" + layout.Joiner + time.Now().UTC().Format(layout.Format) + ".log"

	// Basic test representing first rotate (no existing files).
	mockFiler.EXPECT().ReadDir(layout.ArchiveDir).Return(fakeFiles, nil)
	mockFiler.EXPECT().Rename("/var/log/service.log", newName)

	for idx := range fakes {
		// We returned 10 fake files, so give them 10 fake file names.
		// Each name is 10 seconds older than the previous. We then test for the age
		// and if it's older than our FileAge value it should be get deleted.
		fileTime := time.Now().Add(-time.Duration(idx*10) * time.Second).UTC()
		fileName := "service" + layout.Joiner + fileTime.Format(layout.Format) + ".log"
		fakes[idx].EXPECT().Name().Return(fileName)

		if idx >= layout.FileCount {
			mockFiler.EXPECT().Remove(layout.ArchiveDir + "/" + fileName)
		} else if time.Since(fileTime) > layout.FileAge {
			mockFiler.EXPECT().Remove(layout.ArchiveDir + "/" + fileName)
		}
	}

	//
	file, err := layout.Rotate("/var/log/service.log")
	assert.Equal(newName, file)
	assert.Nil(err)
}
