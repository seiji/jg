package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var (
	Version string = "HEAD"
)

func main() {
	newApp().Run(os.Args)
}

func newApp() (app *cli.App) {
	app = cli.NewApp()
	app.Name = "jg"
	app.Usage = "json to go struct"
	app.Version = Version
	app.Action = Generate
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Value: "User",
			Usage: "Name for starting struct",
		},
		cli.BoolFlag{
			Name:  "omitempty, o",
			Usage: "Add field tag of omitempty",
		},
		cli.StringFlag{
			Name:  "package, p",
			Value: "main",
			Usage: "Name for this package",
		},
	}
	return
}
