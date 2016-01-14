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

func doInstall(c *cli.Context) {
	home := os.Getenv("HOME")
	packageName := c.Args().First()
	if _, err := os.Stat(packageName); err == nil {
		fmt.Println("Install from a local package")
		pwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = cp(filepath.Join(pwd, packageName), filepath.Join(home, "/.dpm/cache/", packageName))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fmt.Println("Install from a remote repository")
		err := getter.GetFile(filepath.Join(home, "/.dpm/index/", "dpm.index"), repo+"dpm.index")
		if err != nil {
			fmt.Println(err)
		}
		index, err := ioutil.ReadFile(filepath.Join(home, "/.dpm/index/", "dpm.index"))
		lines := strings.Split(string(index), "\n")
		filename := ""
		for _, line := range lines {
			parts := strings.SplitN(line, "\t", 3)
			if parts[0] == packageName {
				filename = parts[1]
				break
			}
		}

		err = getter.GetFile(filepath.Join(home, "/.dpm/cache", filename), repo+filename)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func doBuild(c *cli.Context) {
	if c.Args().First() == "." {
		p, err := build.BuildPackage(c.Args().First())
		if err != nil {
			fmt.Println(err)
		}

		p.Save()
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
			Usage:   "install and run the package",
			Action:  doInstall,
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "build .dpm package",
			Action:  doBuild,
		},
	}

	app.Run(os.Args)
}
