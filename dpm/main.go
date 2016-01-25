package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/swasd/dpm/build"
	"github.com/swasd/dpm/composition"
	"github.com/swasd/dpm/provision"
	"github.com/swasd/dpm/repo"
)

func cp(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func doInstall(c *cli.Context) {
	home := os.Getenv("HOME")
	packageName := c.Args().First()
	packageFile := ""
	if _, err := os.Stat(packageName); err == nil {
		fmt.Println("Install from a local package")
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		packageFile = filepath.Join(home, "/.dpm/cache/", packageName)
		err = cp(filepath.Join(pwd, packageName), packageFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = generateIndex(filepath.Join(home, "/.dpm/cache/"),
			filepath.Join(home, "/.dpm/index/"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Install from a remote repository")
		_, err = repo.Get(packageName, "")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if packageFile != "" {
		p, err := build.LoadPackage(packageFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		p.Extract(filepath.Join(home, "/.dpm/workspace", p.Sha256()))
	}
}

func install(c *cli.Context) {
	home := os.Getenv("HOME")
	packageName := c.Args().First()
	entry, err := repo.Get(packageName, "")
	if err != nil {
		fmt.Println("Cannot find package in the index")
		os.Exit(1)
	}

	packageFile := filepath.Join(home, ".dpm", "cache", entry.Filename)
	_, err = os.Stat(packageFile)
	if err != nil {
		// not existed
		doInstall(c)
	}

	p, err := build.LoadPackage(packageFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// extract the package
	// it will extract all dependencies in process
	_, err = os.Stat(filepath.Join(home, ".dpm", "workspace", entry.Hash))
	if err != nil {
		err = p.Extract(filepath.Join(home, ".dpm", "workspace", entry.Hash))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	hashes, err := p.Order()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Dependencies resolved...")

	var em provision.ExportedMachine
	for _, hash := range hashes {

		packageSpec, err := build.ReadSpec(hash)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Installing %s:%s (%s)...\n", packageSpec.Name, packageSpec.Version, hash[0:8])

		provisionFile := filepath.Join(home, ".dpm", "workspace", hash, packageSpec.Provision)
		provSpec, err := provision.LoadFromFile(provisionFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		times := 0
	loop:
		err = provSpec.Provision()
		if err != nil {
			fmt.Println(err)
			times++
			if times < 10 {
				goto loop
			}
			os.Exit(1)
		}

		err = provSpec.ExportEnvsToFile(filepath.Join(home, ".dpm", "workspace", hash, ".env"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		em = provSpec.ExportedMachine()

		compose, err := composition.NewProject(em, hash, packageSpec)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		err = compose.Up()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	flag := ""
	mode := "engine"
	if em.Mode == provision.Swarm {
		flag = "--swarm "
		mode = "cluster"
	}
	fmt.Printf("\nExported machine is %s.\n", em.Name)
	fmt.Printf("Run \"docker-machine env %s%s\" to see how to connect to your Docker %s.\n", flag, em.Name, mode)

}

func doInit(c *cli.Context) {
	force := c.Bool("force")
	spec := `---
specVersion: 0.1.0
spec:
  name: unnamed-cluster
  version: 0.1.0
  title: Unnamed cluster
  provision: provision.yml
  composition: composition.yml
  dependencies:
    package: version=1.0.0
  description: >
    This is the description.
`
	write := force
	if write == false {
		_, err := os.Stat("SPEC.yml")
		if err != nil {
			write = true
		}
	}

	if write {
		err := ioutil.WriteFile("SPEC.yml",
			[]byte(spec),
			0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	provision := `---
machines:
  node:
    driver: none

`
	write = force
	if write == false {
		_, err := os.Stat("provision.yml")
		if err != nil {
			write = true
		}
	}

	if write {
		err := ioutil.WriteFile("provision.yml",
			[]byte(provision),
			0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	write = force
	if write == false {
		_, err := os.Stat("composition.yml")
		if err != nil {
			write = true
		}
	}

	if write {
		err := ioutil.WriteFile("composition.yml",
			[]byte{},
			0644)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

}

func doBuild(c *cli.Context) {
	home := os.Getenv("HOME")
	outputDir := os.ExpandEnv(c.String("dir"))
	sourceDir := "."
	if len(c.Args()) >= 1 {
		sourceDir = c.Args().First()
	}

	p, err := build.BuildPackage(sourceDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = p.SaveToDir(outputDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = generateIndex(filepath.Join(home, "/.dpm/cache/"),
		filepath.Join(home, "/.dpm/index/"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	packageSpec, err := p.Spec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(packageSpec.Name)
}

func generateIndex(dir string, outdir string) error {
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	os.MkdirAll(outdir, 0755)
	if err != nil {
		return err
	}

	entries := make(repo.Entries, 0)
	for _, f := range infos {
		if strings.HasSuffix(f.Name(), ".dpm") {
			p, err := build.LoadPackage(filepath.Join(dir, f.Name()))
			if err != nil {
				return err
			}
			s, err := p.Spec()
			if err != nil {
				return err
			}

			entries = append(entries, &repo.Entry{
				s.Name,
				s.Version,
				f.Name(),
				p.Sha256(),
			})
		}
	}
	return entries.Save(filepath.Join(outdir, "dpm.index"))
}

func doIndex(c *cli.Context) {
	home := os.Getenv("HOME")
	dir := "."
	outdir := "."
	if len(c.Args()) >= 1 {
		dir = c.Args().First()
		outdir = dir
	} else if c.Bool("local") {
		dir = filepath.Join(home, ".dpm", "cache")
		outdir = filepath.Join(home, ".dpm", "index")
	}

	err := generateIndex(dir, outdir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doRemove(c *cli.Context) {
	home := os.Getenv("HOME")
	packageName := c.Args().First()
	e, err := repo.LoadIndex(filepath.Join(home, ".dpm", "index", "dpm.index"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	entry := e.FindByName(packageName)
	if entry == nil {
		fmt.Println("Cannot find package in the index")
		os.Exit(1)
	}

	packageFile := filepath.Join(home, "/.dpm/cache/", entry.Filename)
	_, err = os.Stat(packageFile)
	if err != nil {
		// not existed
		doInstall(c)
	}

	p, err := build.LoadPackage(packageFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	packageSpec, err := p.Spec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	provisionFile := filepath.Join(home, "/.dpm/workspace", entry.Hash, packageSpec.Provision)
	provSpec, err := provision.LoadFromFile(provisionFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = provSpec.RemoveMachines()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func doInfo(c *cli.Context) {
	home := os.Getenv("HOME")
	packageName := c.Args().First()
	e, err := repo.LoadIndex(filepath.Join(home, ".dpm", "index", "dpm.index"))

	entry := e.FindByName(packageName)
	if entry == nil {
		fmt.Println("Cannot find package in the index")
		os.Exit(1)
	}

	packageFile := filepath.Join(home, "/.dpm/cache/", entry.Filename)
	_, err = os.Stat(packageFile)
	if err != nil {
		// not existed
		doInstall(c)
	}

	p, err := build.LoadPackage(packageFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	packageSpec, err := p.Spec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Package Information:")
	fmt.Printf("  Title:   %s\n", packageSpec.Title)
	fmt.Printf("  Name:    %s\n", packageSpec.Name)
	fmt.Printf("  Version: %s\n", packageSpec.Version)
	fmt.Printf("  SHA256:  %s\n", p.Sha256())
	fmt.Printf("  %s\n", packageSpec.Description)
}

func main() {
	app := cli.NewApp()
	app.Name = "dpm"
	app.Usage = "A package manager for Docker"
	app.Version = "0.1-dev"

	app.Commands = []cli.Command{
		{
			Name:    "install",
			Aliases: []string{"i"},
			Usage:   "install the package",
			Action:  install,
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "build .dpm package",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir, d",
					Value: "$HOME/.dpm/cache",
					Usage: "output directory",
				},
			},
			Action: doBuild,
		},
		{
			Name:  "index",
			Usage: "generate dpm.index",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "local",
					Usage: "generate index for the local repository",
				},
			},
			Action: doIndex,
		},
		{
			Name:    "remove",
			Aliases: []string{"rm"},
			Usage:   "remove the package",
			Action:  doRemove,
		},
		{
			Name:   "info",
			Usage:  "show info of the package",
			Action: doInfo,
		},
		{
			Name:   "init",
			Usage:  "init the package files",
			Action: doInit,
		},
	}

	app.Run(os.Args)
}
