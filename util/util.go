package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

func GetHomeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE")
	}
	return os.Getenv("HOME")
}

func DpmHome() string {
	dir := os.Getenv("DPM_HOME")
	if dir == "" {
		home := GetHomeDir()
		dir = filepath.Join(home, ".dpm")
	}

	return dir
}

func nextId() (string, error) {
	infos, err := ioutil.ReadDir(DpmHome())
	if err != nil {
		return "", err
	}

	max := 0
	for _, info := range infos {
		i, err := strconv.Atoi(info.Name())
		if err != nil {
			continue
		}

		if i > max {
			max = i
		}
	}

	return fmt.Sprintf("%d", max+1), nil
}

// machine space
func SpacePath() (string, error) {
	id, err := nextId()
	if err != nil {
		return "", err
	}

	return filepath.Join(DpmHome(), id), nil
}

func NewSpace() (string, error) {
	space, err := SpacePath()
	if err != nil {
		return "", err
	}

	err = os.MkdirAll(space, 0755)
	if err != nil {
		return "", err
	}

	return space, nil
}
