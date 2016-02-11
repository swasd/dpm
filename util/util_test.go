package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHome(t *testing.T) {
	h := os.Getenv("DPM_HOME")
	os.Setenv("DPM_HOME", filepath.Join("/tmp", ".dpm"))
	assert.Equal(t, DpmHome(), filepath.Join("/tmp", ".dpm"))

	os.Setenv("DPM_HOME", h)
}

func TestId(t *testing.T) {
	h := os.Getenv("DPM_HOME")
	os.Setenv("DPM_HOME", filepath.Join("/tmp", ".dpm"))

	os.RemoveAll("/tmp/.dpm")
	os.MkdirAll("/tmp/.dpm", 0644)
	id, err := SpacePath()

	assert.NoError(t, err)
	assert.Equal(t, id, filepath.Join("/tmp", ".dpm", "1"))

	os.Setenv("DPM_HOME", h)
}

func TestCreateSpace(t *testing.T) {
	h := os.Getenv("DPM_HOME")
	os.Setenv("DPM_HOME", filepath.Join("/tmp", ".dpm"))

	os.RemoveAll("/tmp/.dpm")
	os.MkdirAll("/tmp/.dpm", 0644)

	result := ""
	var err error
	for i := 1; i <= 4; i++ {
		result, err = NewSpace()
	}

	assert.NoError(t, err)
	assert.Equal(t, result, filepath.Join("/tmp", ".dpm", "4"))

	os.Setenv("DPM_HOME", h)
}