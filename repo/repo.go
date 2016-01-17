package repo

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/hashicorp/go-getter"
)

func Get(nameOrId string, version string) (*Entry, error) {
	entries, err := getIndex()
	if err != nil {
		return nil, err
	}

	var entry *Entry
	// if version is not specified, assume it may be an ID
	if version == "" {
		_, err := hex.DecodeString(nameOrId)
		if err != nil {
			entry = entries.FindByName(nameOrId)
		}
		entry = entries.findByPartialHash(nameOrId)
	}

	entry = entries.findByNameAndVersion(nameOrId, version)

	if entry == nil {
		return nil, fmt.Errorf("Entry not found")
	}

	home := os.Getenv("HOME")
	err = getter.GetFile(
		filepath.Join(home, ".dpm", "cache", entry.Filename),
		Repo+entry.Filename)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

type Entry struct {
	PackageName string
	Version     string
	Filename    string
	Hash        string
}

type Entries []*Entry

func (e Entries) FindByName(packageName string) *Entry {
	for _, ee := range e {
		if ee.PackageName == packageName {
			return ee
		}
	}
	return nil
}

func (e Entries) findByNameAndVersion(name string, version string) *Entry {
	for _, ee := range e {
		if ee.PackageName == name && ee.Version == version {
			return ee
		}
	}
	return nil
}

func (e Entries) findByHash(id string) *Entry {
	if _, err := hex.DecodeString(id); err != nil {
		return nil
	}

	if len(id) != 64 {
		return nil
	}
	for _, ee := range e {
		if ee.Hash == id {
			return ee
		}
	}
	return nil
}

func (e Entries) findByPartialHash(partialId string) *Entry {
	for _, ee := range e {
		if strings.HasSuffix(ee.Hash, partialId) {
			return ee
		}
	}
	return nil
}

func (e Entries) Save(filename string) error {
	data, err := yaml.Marshal(e)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}

func LoadIndex(filename string) (Entries, error) {
	e := make(Entries, 0)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}

	return e, nil
}

const (
	Repo = "https://raw.githubusercontent.com/swasd/dpm-repo/master/"
)

func getIndex() (Entries, error) {
	home := os.Getenv("HOME")
	err := getter.GetFile(filepath.Join(home, ".dpm", "index", "dpm.index.remote"), Repo+"dpm.index")
	if err != nil {
		return nil, err
	}

	entries, err := LoadIndex(filepath.Join(home, ".dpm", "index", "dpm.index.remote"))
	if err != nil {
		return nil, err
	}

	return entries, nil
}
