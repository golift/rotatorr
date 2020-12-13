package rotatorr_test

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golift.io/rotatorr"
	"golift.io/rotatorr/introtator"
	"golift.io/rotatorr/mocks"
)

// Basic run of the mill usage. Hits 85% of the code just doing normal things.
func TestNew(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	l := rotatorr.NewMust(&rotatorr.Config{
		FileSize: 50,
		Rotatorr: &introtator.Layout{},
	})

	log.SetOutput(l)
	log.Println("weeeeeeeee!")
	log.Println("weee!")
	err := log.Output(1, "weeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee!")
	assert.ErrorIs(err, rotatorr.ErrWriteTooLarge)
	//
	_, err = l.Rotate()
	assert.Nil(err)
	assert.Nil(l.Close())
}

func TestRotateSize(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRotatorr := mocks.NewMockRotatorr(mockCtrl)
	testFile := filepath.Join(os.TempDir(), "mylog.log")
	_ = os.Remove(testFile)

	mockRotatorr.EXPECT().Dirs(gomock.Any())
	//
	l, err := rotatorr.New(&rotatorr.Config{
		Filepath: testFile,
		FileSize: 50,
		Rotatorr: mockRotatorr,
	})
	if err != nil {
		assert.Nil(err)

		return
	}
	//
	msg := []byte("log message")                                                           // len: 11
	s, err := l.Write(append(append(append(append(msg, msg...), msg...), msg...), msg...)) // len: 55
	assert.ErrorIs(err, rotatorr.ErrWriteTooLarge, "writing more data than our filesize must produce an error")
	assert.Equal(0, s, "size must be 0 if the write fails.")

	check := func(s int, err error) {
		assert.Nil(err)
		assert.Equal(len(msg), s)
	}
	check(l.Write(msg)) // 11
	check(l.Write(msg)) // 22
	check(l.Write(msg)) // 33
	check(l.Write(msg)) // 44
	mockRotatorr.EXPECT().Rotate(testFile)
	check(l.Write(msg)) // 55 > 50, rotate!
}

func TestRotateEvery(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRotatorr := mocks.NewMockRotatorr(mockCtrl)
	testFile := filepath.Join(os.TempDir(), "mylog.log")
	_ = os.Remove(testFile)

	mockRotatorr.EXPECT().Dirs(gomock.Any())
	//

	l, err := rotatorr.New(&rotatorr.Config{
		Filepath: testFile,
		Every:    time.Second,
		Rotatorr: mockRotatorr,
	})
	if err != nil {
		assert.Nil(err)

		return
	}
	//
	msg := []byte("log message")                                                           // len: 11
	s, err := l.Write(append(append(append(append(msg, msg...), msg...), msg...), msg...)) // len: 55
	assert.Nil(err)
	assert.Equal(len(msg)*5, s)

	check := func(s int, err error) {
		assert.Nil(err)
		assert.Equal(len(msg), s)
	}
	check(l.Write(msg)) // 11
	check(l.Write(msg)) // 22
	time.Sleep(time.Second)
	mockRotatorr.EXPECT().Rotate(testFile)
	check(l.Write(msg)) // 33
}
