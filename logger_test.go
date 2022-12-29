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
	logger := rotatorr.NewMust(&rotatorr.Config{
		FileSize: 50,
		Rotatorr: &introtator.Layout{},
	})

	log.SetOutput(logger)
	log.Println("weeeeeeeee!")
	log.Println("weee!")
	err := log.Output(1, "weeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee!")
	assert.ErrorIs(err, rotatorr.ErrWriteTooLarge)
	//
	_, err = logger.Rotate()
	assert.Nil(err)
	assert.Nil(logger.Close())
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
	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: testFile,
		FileSize: 50,
		Rotatorr: mockRotatorr,
	})
	if err != nil {
		assert.Nil(err)

		return
	}
	//
	msg := []byte("log message")                                                                // len: 11
	s, err := logger.Write(append(append(append(append(msg, msg...), msg...), msg...), msg...)) // len: 55
	assert.ErrorIs(err, rotatorr.ErrWriteTooLarge, "writing more data than our filesize must produce an error")
	assert.Equal(0, s, "size must be 0 if the write fails.")

	check := func(s int, err error) {
		assert.Nil(err)
		assert.Equal(len(msg), s)
	}
	check(logger.Write(msg)) // 11
	check(logger.Write(msg)) // 22
	check(logger.Write(msg)) // 33
	check(logger.Write(msg)) // 44
	mockRotatorr.EXPECT().Rotate(testFile)
	check(logger.Write(msg)) // 55 > 50, rotate!
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

	logger, err := rotatorr.New(&rotatorr.Config{
		Filepath: testFile,
		Every:    time.Second,
		Rotatorr: mockRotatorr,
	})
	if err != nil {
		assert.Nil(err)

		return
	}
	//
	msg := []byte("log message")                                                                // len: 11
	s, err := logger.Write(append(append(append(append(msg, msg...), msg...), msg...), msg...)) // len: 55
	assert.Nil(err)
	assert.Equal(len(msg)*5, s)

	check := func(s int, err error) {
		assert.Nil(err)
		assert.Equal(len(msg), s)
	}
	check(logger.Write(msg)) // 11
	check(logger.Write(msg)) // 22
	time.Sleep(time.Second)
	mockRotatorr.EXPECT().Rotate(testFile)
	check(logger.Write(msg)) // 33
}
