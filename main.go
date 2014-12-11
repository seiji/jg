package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "jg"
	app.Usage = "json to go struct"
	app.Action = Generate
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Value: "User",
			Usage: "name for the struct",
		},
	}
	app.Run(os.Args)
}
