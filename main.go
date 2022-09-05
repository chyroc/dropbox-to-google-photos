package main

import (
	"log"
	"os"

	"github.com/chyroc/dropbox-to-google-photos/pkg/app"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "dropbox-to-google-photos",
		Usage:     "sync dropbox to google photos",
		UsageText: "dropbox-to-google-photos [command]",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "init config",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "force", Usage: "force init config"},
				},
				Action: func(c *cli.Context) error {
					ins := app.NewApp("")
					return ins.InitConfig(c.Bool("force"))
				},
			},
			{
				Name:  "auth",
				Usage: "auth to google photos",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Usage: "config file path"},
				},
				Action: func(c *cli.Context) error {
					ins := app.NewApp(c.String("config"))
					if err := ins.Start(); err != nil {
						return err
					}
					defer ins.Close()

					return ins.TryAuth()
				},
			},
			{
				Name:  "sync",
				Usage: "sync dropbox to google photos",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "config", Usage: "config file path"},
				},
				Action: func(c *cli.Context) error {
					ins := app.NewApp(c.String("config"))
					if err := ins.Start(); err != nil {
						return err
					}
					defer ins.Close()

					if err := ins.TryAuth(); err != nil {
						return err
					}

					return ins.Sync()
				},
			},
		},
	}

	if err := app.Run(os.Args[:]); err != nil {
		log.Fatalln(err)
	}
}
