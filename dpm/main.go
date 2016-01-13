package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/hashicorp/go-getter"
	"github.com/swasd/dpm/build"
)

func doInstall(c *cli.Context) {
	packageName := c.Args().First()
	err := getter.GetFile(packageName+".dpm",
		"https://raw.githubusercontent.com/swasd/dpm-repo/master/"+packageName+".dpm")
	if err != nil {
		fmt.Println(err)
	}
	// for testing
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
