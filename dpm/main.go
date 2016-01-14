package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

func doInstall(c *cli.Context) {
	packageName := c.Args().First()
	if _, err := os.Stat(packageName); err == nil {
		fmt.Println("Install from a local package")
		home := os.Getenv("HOME")
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
		err := getter.GetFile(packageName+".dpm",
			"https://raw.githubusercontent.com/swasd/dpm-repo/master/"+packageName+".dpm")
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
