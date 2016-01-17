package repo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadEntries(t *testing.T) {
	e := make(Entries, 0)
	e = append(e, &Entry{
		PackageName: "consul-discovery",
		Version:     "1.0.0",
		Hash:        "c00756411ad128488cf8f4e862e118acf1c59d29bd6c0568d527eece823d910e",
		Filename:    "consul-discovery_1.0.0-do.dpm",
	})
	err := e.Save("/tmp/dpm.index")
	assert.NoError(t, err)

	e2, err := LoadIndex("/tmp/dpm.index")
	assert.NoError(t, err)
	assert.Equal(t, len(e2), 1)
}
