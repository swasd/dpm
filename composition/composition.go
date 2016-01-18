package composition

import (
	"os"
	"os/exec"
	"path/filepath"
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

func NewProject(em provision.ExportedMachine, hash string, s *build.Spec) (*Spec, error) {
	return &Spec{em.Name, em.Mode, hash, s.Name, s.Composition}, nil
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
			entry := strings.SplitN(parts[1], "=", 2)
			entry[1] = strings.TrimLeft(strings.TrimRight(entry[1], `"`), `"`)
			result = append(result, entry[0]+"="+entry[1])
		}
	}
	return result, nil
}

func dpmHome() string {
	home := os.Getenv("HOME")
	return filepath.Join(home, ".dpm")
}

func (s *Spec) Up() error {
	home := os.Getenv("HOME")
	dir := filepath.Join(home, ".dpm", "workspace", s.hash)

	info, err := os.Stat(filepath.Join(dir, s.compositionFile))
	if err != nil {
		return err
	}

	if info.Size() == int64(0) {
		// peacefully skip
		return nil
	}

	env, err := s.GetHostEnv()
	if err != nil {
		return err
	}

	cmd := exec.Command("docker-compose",
		"-p", s.projectName,
		"-f", s.compositionFile,
		"up", "-d")
	cmd.Env = env
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
