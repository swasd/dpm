package build

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuild(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	assert.NotNil(t, p)
	err = p.SaveToFile("./_test/dir.dpm")
	assert.NoError(t, err)
	p2, err := LoadPackage("./_test/dir.dpm")
	assert.NoError(t, err)
	assert.Equal(t, len(p.content), len(p2.content))
	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)
}

func TestSave(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	err = p.SaveToDir("./_test")
	assert.NoError(t, err)
	_, err = os.Stat("./_test/test_0.1.0.dev-do+none.dpm")
	assert.NoError(t, err)
	err = os.Remove("./_test/test_0.1.0.dev-do+none.dpm")
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
	_, err = os.Stat("./_extract/provision.yml")
	assert.NoError(t, err)
	_, err = os.Stat("./_extract/composition.yml")
	assert.NoError(t, err)
	_, err = os.Stat("./_extract/web")
	assert.NoError(t, err)
	_, err = os.Stat("./_extract/web/Dockerfile")
	assert.NoError(t, err)
	_, err = os.Stat("./_extract/back")
	assert.NoError(t, err)
	_, err = os.Stat("./_extract/back/Dockerfile")
	assert.NoError(t, err)

	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)

	err = os.RemoveAll("./_extract")
	assert.NoError(t, err)
}

func TestSpecInfo(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	err = p.SaveToFile("./_test/dir.dpm")
	assert.NoError(t, err)
	p2, err := LoadPackage("./_test/dir.dpm")
	assert.NoError(t, err)
	assert.Equal(t, len(p.content), len(p2.content))

	spec, err := p2.Spec()
	assert.NoError(t, err)

	assert.Equal(t, spec.Provision, "provision.yml")
	assert.Equal(t, spec.Composition, "composition.yml")
	assert.Equal(t, spec.Title, "Test - dpm test package")
	assert.Equal(t, spec.Name, "test")
	assert.Equal(t, spec.Version, "0.1.0.dev")
	assert.Equal(t, spec.Description, "This is a test package.\n")
	assert.Equal(t, spec.Dirs, []string{"web", "back"})
	assert.Equal(t, spec.Dependencies["pack1"], "version=1.0")
	assert.Equal(t, spec.Dependencies["pack2"], "version=2.0")

	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)
}

func TestGetDeps(t *testing.T) {
	p, err := BuildPackage("./_test")
	assert.NoError(t, err)
	err = p.SaveToFile("./_test/dir.dpm")
	assert.NoError(t, err)
	p2, err := LoadPackage("./_test/dir.dpm")
	assert.NoError(t, err)
	assert.Equal(t, len(p.content), len(p2.content))

	deps, err := p2.Deps()
	assert.Equal(t, len(deps), 1)

	err = os.Remove("./_test/dir.dpm")
	assert.NoError(t, err)
}

func TestParseAttributes(t *testing.T) {
	m, err := parse("version=1.0.1 instances=1 x=\"x y\" ")
	assert.NoError(t, err)
	assert.Equal(t, m["version"], "1.0.1")
	assert.Equal(t, m["instances"], "1")
	assert.Equal(t, m["x"], "x y")
}

func TestMergeDepGraph(t *testing.T) {
	g1 := make(DepGraph)
	g1["a"] = []string{"b", "c"}
	g1["b"] = []string{"a", "b", "c"}

	g2 := make(DepGraph)
	g2["a"] = []string{"c", "d"}
	g2["b"] = []string{"d", "b", "c", "a"}

	out := merge(g1, g2)
	assert.Equal(t, out["a"], []string{"b", "c", "d"})
	assert.Equal(t, out["b"], []string{"a", "b", "c", "d"})
}
