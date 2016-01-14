package provision

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadSpec(t *testing.T) {
	yml := `---
machines:
  ocean:
    driver: digitalocean
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	assert.NotNil(t, spec.MachineSpecs["ocean"])
}

func TestReadSpecFromString(t *testing.T) {
	yml := `---
machines:
  ocean:
    driver: digitalocean
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 1)

	m := spec.Machine("ocean")
	assert.Equal(t, m.Name(), "ocean")
	assert.Equal(t, m.Driver(), "digitalocean")
}

func TestSimpleProvision(t *testing.T) {
	yml := `---
machines:
  ocean:
    instances: 2
    driver: digitalocean
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("ocean-1")
	assert.NotNil(t, m)
	assert.Equal(t, m.Name(), "ocean-1")
	assert.Equal(t, m.Driver(), "digitalocean")
}

func TestGenCommandLine(t *testing.T) {
	yml := `---
machines:
  ocean:
    instances: 2
    driver: digitalocean
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("ocean-1")
	assert.NotNil(t, m)
	assert.Equal(t, m.CmdLine(), []string{
		"--driver", "digitalocean",
		"ocean-1"})
}

func TestGenCmdLineWithOption(t *testing.T) {
	yml := `---
machines:
  ocean:
    instances: 2
    driver: digitalocean
    options:
      engine-install-url: https://test.docker.com
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("ocean-1")
	assert.NotNil(t, m)
	assert.Equal(t, m.CmdLine(), []string{
		"--driver", "digitalocean",
		"--engine-install-url", "https://test.docker.com",
		"ocean-1"})
}

func TestGenCommandLineWithSlice(t *testing.T) {
	yml := `---
machines:
  ocean:
    instances: 2
    driver: digitalocean
    options:
      engine-install-url: https://test.docker.com
      engine-opt:
        cluster-store: consul://192.168.1.81:8500
        cluster-advertise: eth0:2376
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("ocean-1")
	assert.NotNil(t, m)
	assert.Equal(t, m.CmdLine(), []string{
		"--driver", "digitalocean",
		"--engine-install-url", "https://test.docker.com",
		"--engine-opt", "cluster-advertise=eth0:2376",
		"--engine-opt", "cluster-store=consul://192.168.1.81:8500",
		"ocean-1"})
}

func TestGenCommandLineWithSwarm(t *testing.T) {
	yml := `---
machines:
  ocean:
    instances: 2
    driver: digitalocean
    options:
      engine-install-url: https://test.docker.com
      engine-opt:
        cluster-store: consul://192.168.1.81:8500
        cluster-advertise: eth0:2376
      swarm: true
      swarm-discovery: consul://1.2.3.4:8500/test
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("ocean-1")
	assert.NotNil(t, m)
	assert.Equal(t, m.CmdLine(), []string{
		"--driver", "digitalocean",
		"--engine-install-url", "https://test.docker.com",
		"--engine-opt", "cluster-advertise=eth0:2376",
		"--engine-opt", "cluster-store=consul://192.168.1.81:8500",
		"--swarm",
		"--swarm-discovery", "consul://1.2.3.4:8500/test",
		"ocean-1"})
}

func TestCreate(t *testing.T) {
	yml := `---
machines:
  fake:
    instances: 2
    driver: none
    options:
      url: tcp://1.2.3.4:1234
      engine-install-url: https://test.docker.com
      engine-opt:
        cluster-store: consul://192.168.1.81:8500
        cluster-advertise: eth0:2376
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("fake-1")
	assert.NotNil(t, m)
	err = m.Create()
	assert.NoError(t, err)

	err = m.Delete()
	assert.NoError(t, err)
}

func TestExpadingProvision(t *testing.T) {
	yml := `---
machines:
  fake:
    instances: 2
    driver: none
    options:
      url: tcp://1.2.3.4:1234
      engine-install-url: https://test.docker.com
      engine-opt:
        cluster-store: consul://192.168.1.81:8500
        cluster-advertise: eth0:2376
    post-provision:
      - bash -c echo ${ip fake-1} ${fake-1}
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("fake-1")
	err = m.Create()
	assert.NoError(t, err)
	for _, p := range m.PostProvision() {
		assert.Equal(t, p, "bash -c echo 1.2.3.4 1.2.3.4")
	}
	err = m.Delete()
	assert.NoError(t, err)
}

func TestPostProvision(t *testing.T) {
	yml := `---
machines:
  fake:
    instances: 2
    driver: none
    options:
      url: tcp://1.2.3.4:1234
      engine-install-url: https://test.docker.com
      engine-opt:
        cluster-store: consul://192.168.1.81:8500
        cluster-advertise: eth0:2376
    post-provision:
      - bash -c "echo ${ip fake-1} ${fake-1} ${self}"
`
	spec, err := Read([]byte(yml))
	assert.NoError(t, err)
	machines := spec.Machines()
	assert.Equal(t, len(machines), 2)

	m := spec.Machine("fake-1")
	err = m.Create()
	assert.NoError(t, err)
	for _, p := range m.PostProvision() {
		assert.Equal(t, p, "bash -c \"echo 1.2.3.4 1.2.3.4 fake-1\"")
	}
	out, err := m.ExecutePostProvision()
	assert.NoError(t, err)
	assert.Equal(t, out[0], "1.2.3.4 1.2.3.4 fake-1\n")

	err = m.Delete()
	assert.NoError(t, err)
}
