package main

import (
	"os"

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
				Usage:   "enable debug mode",
			},
		},
		Commands: []*cli.Command{
			{
				Name:      "run",
				Aliases:   []string{"r"},
				Usage:     "run a command in a new container",
				ArgsUsage: `COMMAND [ARG...]`,
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

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
