package build

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/jhoonb/archivex"
	"github.com/swasd/dpm/provision"
)

type Package struct {
	content []byte
}

func BuildPackage(dir string) (*Package, error) {
	buf := new(bytes.Buffer)
	tarfile := new(archivex.TarFile)
	tarfile.Writer = tar.NewWriter(buf)

	specContent, err := ioutil.ReadFile(filepath.Join(dir, "SPEC.yml"))
	var root Root
	err = yaml.Unmarshal(specContent, &root)
	if err != nil {
		return nil, err
	}
	spec := root.Spec

	tarfile.AddFileWithName(filepath.Join(dir, "SPEC.yml"), "SPEC.yml")
	tarfile.AddFileWithName(filepath.Join(dir, spec.Provision), spec.Provision)
	tarfile.AddFileWithName(filepath.Join(dir, spec.Composition), spec.Composition)
	for _, d := range spec.Dirs {
		tarfile.AddAll(filepath.Join(dir, d), true)
	}

	// resolve dependencies on build
	// to gaurantee that the package will have
	// the same behaviour everytime we deploy it

	/*
			allDependencies := []string{}
			for name, attributes := range spec.Dependencies {
				attrs := parse(attributes)
				pack, err := repo.Get(name, attrs["version"])
				allDependencies = append(allDependencies, pack.Dependencies()...)
				// lookup
				// not found, resolve
				// patch
				// collect into array
			}
				for _, d := range resolveDependencies {
					//   - patch override attribute:
					//   - tarfile.AddAll(hash)
				}
		for _, hash := range allDependencies {
			tarfile.AddAll(workspace+"/"+hash, true)
		}
	*/
	tarfile.Close()
	return &Package{buf.Bytes()}, nil
}

func LoadPackage(filename string) (*Package, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Package{content}, nil
}

func (p *Package) Save() error {
	return p.SaveToDir(".")
}

func (p *Package) Sha256() string {
	s := sha256.Sum256(p.content)
	return hex.EncodeToString(s[:])
}

func (p *Package) SaveToDir(dir string) error {
	spec, err := p.Spec()
	if err != nil {
		return err
	}
	platforms, err := p.platforms()
	if err != nil {
		return err
	}
	filename := spec.Name + "_" + spec.Version + "-" + platforms + ".dpm"
	return p.SaveToFile(filepath.Join(dir, filename))
}

func (p *Package) SaveToFile(filename string) error {
	return ioutil.WriteFile(filename, p.content, 0644)
}

type Spec struct {
	Name         string
	Version      string
	Provision    string
	Composition  string
	Title        string
	Description  string
	Dirs         []string
	Dependencies map[string]string // in `"package": version=number` format
}

type Root struct {
	SpecVersion string `yaml:"specVersion"`
	Spec        *Spec
}

func (p *Package) Spec() (*Spec, error) {
	br := bytes.NewReader(p.content)
	tr := tar.NewReader(br)
	hdr, err := tr.Next()
	if hdr.Name != "SPEC.yml" {
		return nil, fmt.Errorf("File format incorrect")
	}
	specContent := make([]byte, hdr.Size)
	n, err := io.ReadFull(tr, specContent)
	if int64(n) != hdr.Size {
		return nil, fmt.Errorf("Size not match")
	}
	root := Root{}
	err = yaml.Unmarshal(specContent, &root)
	if err != nil {
		return nil, err
	}

	if root.SpecVersion != "0.1.0" {
		return nil, fmt.Errorf("Spec version '%s' is not supported.", root.SpecVersion)
	}

	return root.Spec, nil
}

func (p *Package) platforms() (string, error) {
	result := make(map[string]bool)
	pp, err := p.provision()
	if err != nil {
		return "", err
	}
	for _, m := range pp.Machines() {
		switch m.Driver() {
		case "amazonec2":
			result["aws"] = true
		case "azure":
			result["az"] = true
		case "exoscale":
			result["ex"] = true
		case "google":
			result["gce"] = true
		case "generic":
			result["ge"] = true
		case "hyperv":
			result["hv"] = true
		case "openstack":
			result["os"] = true
		case "rackspace":
			result["rs"] = true
		case "softlayer":
			result["sl"] = true
		case "virtualbox":
			result["vbox"] = true
		case "vmwarevcloudair":
			result["vca"] = true
		case "vmwarefusion":
			result["vf"] = true
		case "vmwarevsphere":
			result["vs"] = true
		case "digitalocean":
			result["do"] = true
		case "none":
			result["none"] = true
		}
	}
	keys := []string{}
	for k, _ := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "+"), nil
}

func (p *Package) provision() (*provision.Spec, error) {
	spec, err := p.Spec()
	if err != nil {
		return nil, err
	}
	prov := spec.Provision

	br := bytes.NewReader(p.content)
	tr := tar.NewReader(br)
	hdr, err := tr.Next()
	for {
		if err != nil {
			return nil, err
		}
		if hdr.Name == prov {
			break
		}
		hdr, err = tr.Next()
	}

	provisionContent := make([]byte, hdr.Size)
	n, err := io.ReadFull(tr, provisionContent)
	if int64(n) != hdr.Size {
		return nil, fmt.Errorf("Size not match")
	}
	return provision.Read(provisionContent)
}

func (p *Package) Extract(dest string) error {

	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(bytes.NewReader(p.content))
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if hdr.Name == "." {
			continue
		}

		err = extractTarArchiveFile(hdr, dest, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func extractTarArchiveFile(header *tar.Header, dest string, input io.Reader) error {
	filePath := filepath.Join(dest, header.Name)
	fileInfo := header.FileInfo()

	if fileInfo.IsDir() {
		err := os.MkdirAll(filePath, fileInfo.Mode())
		if err != nil {
			return err
		}
	} else {
		err := os.MkdirAll(filepath.Dir(filePath), 0755)
		if err != nil {
			return err
		}

		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return os.Symlink(header.Linkname, filePath)
		}

		fileCopy, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileInfo.Mode())
		if err != nil {
			return err
		}
		defer fileCopy.Close()

		_, err = io.Copy(fileCopy, input)
		if err != nil {
			return err
		}
	}

	return nil
}
