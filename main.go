package main

import (
	"fmt"
	"os"

	"github.com/mousany/gophinator/container"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:    "gophinator",
		Version: "v0.1.0",
		Usage:   "A minimal container runtime implemented in Go",

		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"d"},
				Usage:   "enable debug logging",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "run",
				Aliases:   []string{"r"},
				Usage:     "run a command in a new container",
				ArgsUsage: `COMMAND [-- ARGUMENTS]`,
				Flags: []cli.Flag{
					&cli.UintFlag{
						Name:    "uid",
						Aliases: []string{"u"},
						Usage:   "create the container with the specified `UID`",
					},
					&cli.StringFlag{
						Name:    "volume",
						Aliases: []string{"v"},
						Usage:   "mount the root of the container at the given `VOLUME`",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					}

					if c.NArg() < 1 {
						fmt.Print("Incorrect Usage: command needs an argument: run\n\n")
						cli.ShowSubcommandHelpAndExit(c, 1)
					}

					if c.Args().Len() > 1 && c.Args().Get(1) != "--" {
						fmt.Printf("Incorrect Usage: arguments must be preceded by '--': %s\n\n", c.Args().Get(1))
						cli.ShowSubcommandHelpAndExit(c, 1)
					}

					args := []string{}
					for i := 2; i < c.Args().Len(); i++ {
						args = append(args, c.Args().Get(i))
					}

					con := container.New(c.Args().First(), args, c.Uint("uid"), c.String("volume"))

					return con.Run()
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
