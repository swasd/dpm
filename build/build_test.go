package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	err = p.SaveToFile("./_test/dir.dpm")
	assert.NoError(t, err)
	p2, err := LoadPackage("./_test/dir.dpm")
	assert.NoError(t, err)
	assert.Equal(t, len(p.content), len(p2.content))
	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)
}

func TestGetSpec(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	err = p.SaveToFile("./_test/dir.dpm")
	assert.NoError(t, err)
	p2, err := LoadPackage("./_test/dir.dpm")
	assert.NoError(t, err)
	assert.Equal(t, len(p.content), len(p2.content))

	spec, err := p2.Spec()
	assert.Equal(t, spec.Provision, "provision.yml")
	assert.Equal(t, spec.Composition, "composition.yml")

	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)
}

func TestExtract(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	err = p.SaveToFile("./_test/dir.dpm")
	assert.NoError(t, err)
	p2, err := LoadPackage("./_test/dir.dpm")
	assert.NoError(t, err)
	assert.Equal(t, len(p.content), len(p2.content))

	err = p2.Extract("./_extract")
	assert.NoError(t, err)

	_, err = os.Stat("./_extract/SPEC.yml")
	assert.NoError(t, err)

	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)

	err = os.RemoveAll("./_extract")
	assert.NoError(t, err)
}
