package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mousany/gophinator/runtime"
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
					&cli.IntFlag{
						Name:    "uid",
						Aliases: []string{"u"},
						Usage:   "create the container with the specified `UID`",
					},
					&cli.StringFlag{
						Name:    "root",
						Aliases: []string{"r"},
						Usage:   "mount the root of the container at the given `ROOT`",
					},
					&cli.StringSliceFlag{
						Name:    "volume",
						Aliases: []string{"v"},
						Usage:   "mount the given `VOLUME`s into the container",
					},
				},
				Action: func(c *cli.Context) error {
					if c.Bool("debug") {
						logrus.SetLevel(logrus.DebugLevel)
					}

					if c.NArg() < 1 {
						fmt.Fprintln(os.Stderr, "Incorrect Usage: command needs an argument: run")
						fmt.Fprintln(os.Stderr)
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					if c.Args().Len() > 1 && c.Args().Get(1) != "--" {
						fmt.Fprintf(os.Stderr, "Incorrect Usage: arguments must be preceded by '--': %s", c.Args().Get(1))
						fmt.Fprintln(os.Stderr)
						cli.ShowSubcommandHelpAndExit(c, 1)
					}
					args := []string{}
					for i := 2; i < c.Args().Len(); i++ {
						args = append(args, c.Args().Get(i))
					}

					volumes := []runtime.VolumePair{}
					for _, v := range c.StringSlice("volume") {
						chunks := strings.Split(v, ":")
						if len(chunks) != 2 {
							fmt.Fprintf(os.Stderr, "Incorrect Usage: volume must be in the form 'source:target': %s", v)
							fmt.Fprintln(os.Stderr)
							cli.ShowSubcommandHelpAndExit(c, 1)
						}
						source := chunks[0]
						target := strings.TrimPrefix(chunks[1], "/")
						volumes = append(volumes, runtime.VolumePair{Source: source, Target: target})
					}

					con, err := runtime.New(c.Args().First(), args, c.Int("uid"), c.String("root"), volumes)
					if err != nil {
						logrus.Errorf("Fail to create container: %s", err)
						return err
					}

					stat, err := con.Run()
					if err != nil {
						logrus.Errorf("Fail to run container: %s", err)
					} else {
						logrus.Infof("Container exited with status %d", stat)
					}

					return err
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
