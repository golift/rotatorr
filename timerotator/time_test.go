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

	l := &timerotator.Layout{PostRotate: func(s1, s2 string) {
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
	l := &timerotator.Layout{ArchiveDir: "/var/log/archives"}
	f, err := l.Dirs("/var/log/service.log")
	assert.Equal([]string{"/var/log", "/var/log/archives"}, f, "the wrong directories were returned")
	assert.Nil(err, "this should not producce an error")
	assert.EqualValues(filer.Default(), l.Filer)
	assert.Equal(l.Joiner, timerotator.DefaultJoiner)
	assert.Equal(l.Format, timerotator.FormatDefault)
}

func TestRotateOne(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFiler := mocks.NewMockFiler(mockCtrl)
	l := &timerotator.Layout{
		Filer:  mockFiler,
		UseUTC: true,
		Format: timerotator.FormatNoSecnd,
		Joiner: timerotator.DefaultJoiner,
	}
	newName := "/var/log/service" + l.Joiner + time.Now().UTC().Format(l.Format) + ".log"

	// Basic test representing first rotate (no existing files).
	mockFiler.EXPECT().ReadDir("/var/log")
	mockFiler.EXPECT().Rename("/var/log/service.log", newName)
	//
	file, err := l.Rotate("/var/log/service.log")
	assert.Equal(newName, file)
	assert.Nil(err)
}

// Make fake files to fake delete.
func testFakeFiles(mockCtrl *gomock.Controller, count int) (fakes []*mocks.MockFileInfo, files []os.FileInfo) {
	for i := 0; i < count; i++ {
		fake := mocks.NewMockFileInfo(mockCtrl)
		fakes = append(fakes, fake)
		files = append(files, fake)
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
	l := &timerotator.Layout{
		ArchiveDir: "/var/log/archives",
		Filer:      mockFiler,
		UseUTC:     true,
		Format:     timerotator.FormatNoSecnd,
		Joiner:     timerotator.DefaultJoiner,
		FileAge:    time.Minute,
		FileCount:  2,
	}
	newName := l.ArchiveDir + "/service" + l.Joiner + time.Now().UTC().Format(l.Format) + ".log"

	// Basic test representing first rotate (no existing files).
	mockFiler.EXPECT().ReadDir(l.ArchiveDir).Return(fakeFiles, nil)
	mockFiler.EXPECT().Rename("/var/log/service.log", newName)

	for i := range fakes {
		// We returned 10 fake files, so give them 10 fake file names.
		// Each name is 10 seconds older than the previous. We then test for the age
		// and if it's older than our FileAge value it should be get deleted.
		ts := time.Now().Add(-time.Duration(i*10) * time.Second).UTC()
		fileName := "service" + l.Joiner + ts.Format(l.Format) + ".log"
		fakes[i].EXPECT().Name().Return(fileName)

		if i >= l.FileCount {
			mockFiler.EXPECT().Remove(l.ArchiveDir + "/" + fileName)
		} else if time.Since(ts) > l.FileAge {
			mockFiler.EXPECT().Remove(l.ArchiveDir + "/" + fileName)
		}
	}

	//
	file, err := l.Rotate("/var/log/service.log")
	assert.Equal(newName, file)
	assert.Nil(err)
}
