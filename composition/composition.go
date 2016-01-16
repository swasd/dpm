package composition

import (
	"os"
	"os/exec"
	"strings"

	"github.com/swasd/dpm/build"
	"github.com/swasd/dpm/provision"
)

type Spec struct {
	host            string
	mode            provision.ExportedMode
	hash            string
	projectName     string
	compositionFile string
}

func NewProject(em provision.ExportedMachine, p *build.Package) (*Spec, error) {
	packageSpec, err := p.Spec()
	if err != nil {
		return nil, err
	}
	hash := p.Sha256()
	file := p.Composition

	return &Spec{em.Name, em.Mode, hash, packageSpec.Name, file}, nil
}

func (s *Spec) GetHostEnv() ([]string, error) {
	result := []string{}
	cmd := exec.Command("docker-machine",
		"-s", dpmHome(), "env", "--shell", "sh", s.host)
	out, err := cmd.Output()
	if err != nil {
		return []string{}, nil
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 && parts[0] == "export" {
			result = append(result, parts[1])
		}
	}
	return result, nil
}

func dpmHome() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".dpm")
}

func (s *Spec) Up() {
	env := s.GetHostEnv()

	home := os.Getenv("HOME")
	cmd := exec.Command("docker-compose",
		"-p", s.projectName,
		"-f", s.compositionFile,
		"up", "-d")
	cmd.Env
	cmd.Dir = filepath.Jon(home, ".dpm/workspace", hash)

}
