package provision

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/mattn/go-shellwords"

	"gopkg.in/yaml.v2"
)

type Spec struct {
	MachineSpecs map[string]MachineSpec `yaml:"machines,omitempty"`
}

type MachineSpec struct {
	Driver        string
	Instances     *int
	Options       map[string]interface{}
	PreProvision  []string `yaml:"pre-provision,omitempty"`
	PostProvision []string `yaml:"post-provision,omitempty"`
}

type Machine struct {
	name    string
	driver  string
	options map[string]interface{}
	pre     []string
	post    []string
}

func LoadFromFile(filename string) (*Spec, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return Read(content)
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
				pre:     v.PreProvision,
				post:    v.PostProvision,
			}
			result = append(result, machine)
		} else {
			for i := 1; i <= *v.Instances; i++ {
				machine := &Machine{
					name:    fmt.Sprintf("%s-%d", k, i),
					driver:  v.Driver,
					options: v.Options,
					pre:     v.PreProvision,
					post:    v.PostProvision,
				}
				result = append(result, machine)
			}
		}
	}
	return result
}

func (s *Spec) Provision() error {
	for _, m := range s.Machines() {
		err := m.create()
		if err != nil {
			return err
		}
		// TODO logging outputs
		_, err = m.executePostProvision()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Machine) Name() string {
	return m.name
}

func (m *Machine) Driver() string {
	return m.driver
}

func (m *Machine) cmdLine() []string {
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
		case bool:
			if val {
				result = append(result, "--"+k)
			}
		}
	}
	result = append(result, m.name)
	return result
}

func (m *Machine) create() error {
	args := append([]string{"-s", ".dpm", "create"}, m.cmdLine()...)
	cmd := exec.Command("docker-machine", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Machine) doDelete() error {
	args := append([]string{"-s", ".dpm", "rm", "-y"}, m.name)
	cmd := exec.Command("docker-machine", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Machine) postProvision() []string {
	result := []string{}
	for _, p := range m.post {
		expanded := os.Expand(p, func(key string) string {

			if key == "self" {
				return m.name
			}

			val := os.Getenv(key)
			if val == "" {
				parts := strings.SplitN(key, " ", 2)
				cmd := ""
				arg := ""
				if len(parts) == 1 {
					cmd = "ip"
					arg = parts[0]
				} else if len(parts) == 2 && parts[0] == "ip" {
					cmd = "ip"
					arg = parts[1]
				}
				out, err := exec.Command("docker-machine", "-s", ".dpm", cmd, arg).Output()
				if err != nil {
					val = ""
				}

				val = strings.TrimSpace(string(out))
				if cmd == "ip" {
					parts := strings.SplitN(val, ":", 2)
					val = parts[0]
				}
			}
			return val
		})
		result = append(result, expanded)
	}
	return result
}

func (m *Machine) executePostProvision() ([]string, error) {
	out := []string{}
	for _, p := range m.postProvision() {
		args, err := shellwords.Parse(p)
		if err != nil {
			return []string{}, err
		}
		cmd := exec.Command(args[0], args[1:]...)
		o, err := cmd.CombinedOutput()
		out = append(out, string(o))
	}
	return out, nil
}
