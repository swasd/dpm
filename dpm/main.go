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
	_, err = os.Stat(filepath.Join(home, ".dpm", "workspace", entry.Hash))
	if err != nil {
		err = p.Extract(filepath.Join(home, ".dpm", "workspace", entry.Hash))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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

	err = provSpec.Provision()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	em := provSpec.ExportedMachine()

	compose, err := composition.NewProject(em, p)
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
	}

	app.Run(os.Args)
}
