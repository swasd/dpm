package build

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/jhoonb/archivex"
)

type Package struct {
	content []byte
}

func BuildPackage(dir string) (*Package, error) {
	buf := new(bytes.Buffer)
	tarfile := new(archivex.TarFile)
	tarfile.Writer = tar.NewWriter(buf)

	specContent, err := ioutil.ReadFile(filepath.Join(dir, "SPEC.yml"))
	root := make(map[string]*Spec)
	err = yaml.Unmarshal(specContent, &root)
	if err != nil {
		return nil, err
	}
	spec := root["spec"]

	tarfile.AddFileWithName(filepath.Join(dir, "SPEC.yml"), "SPEC.yml")
	tarfile.AddFileWithName(filepath.Join(dir, spec.Provision), spec.Provision)
	tarfile.AddFileWithName(filepath.Join(dir, spec.Composition), spec.Composition)
	for _, d := range spec.Dirs {
		tarfile.AddAll(filepath.Join(dir, d), true)
	}
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

func (p *Package) SaveToDir(dir string) error {
	spec, err := p.Spec()
	if err != nil {
		return err
	}
	filename := spec.Name + "_" + spec.Version + "-" + spec.Type + ".dpm"
	return p.SaveToFile(dir + "/" + filename)
}

func (p *Package) SaveToFile(filename string) error {
	return ioutil.WriteFile(filename, p.content, 0644)
}

type Spec struct {
	Name        string
	Version     string
	Type        string
	Provision   string
	Composition string
	Title       string
	Description string
	Dirs        []string
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
	root := make(map[string]*Spec)
	err = yaml.Unmarshal(specContent, &root)
	if err != nil {
		return nil, err
	}

	return root["spec"], nil
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
