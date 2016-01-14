package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/go-getter"
	"github.com/swasd/dpm/build"
	"github.com/swasd/dpm/provision"
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

var repo = "https://raw.githubusercontent.com/swasd/dpm-repo/master/"

func findDpmFromIndex(packageName string) (dpm string, hash string) {
	home := os.Getenv("HOME")
	index, err := ioutil.ReadFile(filepath.Join(home, "/.dpm/index/", "dpm.index"))
	if err != nil {
		return "", ""
	}
	lines := strings.Split(string(index), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "\t", 3)
		if parts[0] == packageName {
			return parts[1], parts[2]
		}
	}
	return "", ""
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
	} else {
		fmt.Println("Install from a remote repository")
		err := getter.GetFile(filepath.Join(home, "/.dpm/index/", "dpm.index"), repo+"dpm.index")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// TODO merge index
		filename, _ := findDpmFromIndex(packageName)
		packageFile = filepath.Join(home, "/.dpm/cache", filename)
		err = getter.GetFile(packageFile, repo+filename)
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

func doRun(c *cli.Context) {
	home := os.Getenv("HOME")
	packageName := c.Args().First()
	filename, hash := findDpmFromIndex(packageName)
	packageFile := filepath.Join(home, "/.dpm/cache/", filename)
	_, err := os.Stat(packageFile)
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

	provisionFile := filepath.Join(home, "/.dpm/workspace", hash, packageSpec.Provision)
	spec, err := provision.LoadFromFile(provisionFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	spec.Provision()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func doBuild(c *cli.Context) {
	outputDir := c.String("dir")
	sourceDir := "."
	if len(c.Args()) >= 1 {
		sourceDir = c.Args().First()
	}

	p, err := build.BuildPackage(sourceDir)
	if err != nil {
		fmt.Println(err)
	}

	err = p.SaveToDir(outputDir)
	if err != nil {
		fmt.Println(err)
	}
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
			Action:  doInstall,
		},
		{
			Name:   "run",
			Usage:  "install and run the package",
			Action: doRun,
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "build .dpm package",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "dir, d",
					Value: ".",
					Usage: "output directory",
				},
			},
			Action: doBuild,
		},
	}

	app.Run(os.Args)
}
