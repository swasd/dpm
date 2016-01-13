package provision

import (
	"fmt"
	"os"
	"os/exec"
	"sort"

	"gopkg.in/yaml.v2"
)

type Spec struct {
	MachineSpecs map[string]MachineSpec `yaml:"machines,omitempty"`
}

type MachineSpec struct {
	Driver    string
	Instances *int
	Options   map[string]interface{}
}

type Machine struct {
	name    string
	driver  string
	options map[string]interface{}
}

func Read(yml []byte) (*Spec, error) {
	spec := &Spec{}
	err := yaml.Unmarshal(yml, &spec)
	if err != nil {
		return nil, err
	}

	return spec, nil
}

func (s *Spec) Machine(name string) *Machine {
	for _, m := range s.Machines() {
		if m.name == name {
			return m
		}
	}
	return nil
}

func (s *Spec) Machines() []*Machine {
	result := []*Machine{}
	for k, v := range s.MachineSpecs {
		if v.Instances == nil {
			v.Instances = new(int)
			*v.Instances = 1
		}
		if *v.Instances == 1 {
			machine := &Machine{
				name:    k,
				driver:  v.Driver,
				options: v.Options,
			}
			result = append(result, machine)
		} else {
			for i := 1; i <= *v.Instances; i++ {
				machine := &Machine{
					name:    fmt.Sprintf("%s-%d", k, i),
					driver:  v.Driver,
					options: v.Options,
				}
				result = append(result, machine)
			}
		}
	}
	return result
}

func (m *Machine) Name() string {
	return m.name
}

func (m *Machine) Driver() string {
	return m.driver
}

func (m *Machine) CmdLine() []string {
	result := []string{"--driver", m.driver}
	keys := []string{}
	for k, _ := range m.options {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := m.options[k]
		switch val := v.(type) {
		case string:
			result = append(result, "--"+k)
			result = append(result, val)
		case map[interface{}]interface{}:
			keys := []string{}
			for kk, _ := range val {
				keys = append(keys, kk.(string))
			}
			sort.Strings(keys)
			for _, kk := range keys {
				vv := val[kk]
				result = append(result, "--"+k)
				result = append(result, kk+"="+vv.(string))
			}
		}
	}
	result = append(result, m.name)
	return result
}

func (m *Machine) Create() error {
	args := append([]string{"-s", ".dpm", "create"}, m.CmdLine()...)
	cmd := exec.Command("docker-machine", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Machine) Delete() error {
	args := append([]string{"-s", ".dpm", "rm", "-y"}, m.name)
	cmd := exec.Command("docker-machine", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
