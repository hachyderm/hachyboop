/*===========================================================================*\
 *           MIT License Copyright (c) 2022 Kris Nóva <kris@nivenly.com>     *
 * *
 *                ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                *
 *                ┃   ███╗   ██╗ ██████╗ ██╗   ██╗ █████╗   ┃                *
 *                ┃   ████╗  ██║██╔═████╗██║   ██║██╔══██╗  ┃                *
 *                ┃   ██╔██╗ ██║██║██╔██║██║   ██║███████║  ┃                *
 *                ┃   ██║╚██╗██║████╔╝██║╚██╗ ██╔╝██╔══██║  ┃                *
 *                ┃   ██║ ╚████║╚██████╔╝ ╚████╔╝ ██║  ██║  ┃                *
 *                ┃   ╚═╝  ╚═══╝ ╚═════╝   ╚═══╝  ╚═╝  ╚═╝  ┃                *
 *                ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛                *
 * *
 *                       This machine kills fascists.                        *
 * *
\*===========================================================================*/

package main

import (
	"context"
	"os"
	"strings"

	hb "github.com/hachyderm/hachyboop"
	"github.com/hachyderm/hachyboop/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var banner = `
██╗  ██╗ █████╗  ██████╗██╗  ██╗██╗   ██╗██████╗  ██████╗  ██████╗ ██████╗ 
██║  ██║██╔══██╗██╔════╝██║  ██║╚██╗ ██╔╝██╔══██╗██╔═══██╗██╔═══██╗██╔══██╗
███████║███████║██║     ███████║ ╚████╔╝ ██████╔╝██║   ██║██║   ██║██████╔╝
██╔══██║██╔══██║██║     ██╔══██║  ╚██╔╝  ██╔══██╗██║   ██║██║   ██║██╔═══╝ 
██║  ██║██║  ██║╚██████╗██║  ██║   ██║   ██████╔╝╚██████╔╝╚██████╔╝██║     
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝   ╚═╝   ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝ 
`

var cfg = &service.HachyboopOptions{
	S3Output: &service.S3Options{},
}

func main() {
	/* Change version to -V */
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "The version of the program.",
	}
	app := &cli.Command{
		Name:      hb.Name,
		Version:   hb.Version,
		Copyright: hb.Copyright,
		Usage:     "A go program.",
		UsageText: `service <options> <flags> 
A longer sentence, about how exactly to use this program`,
		Commands: []*cli.Command{
			&cli.Command{},
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"v"},
				Destination: &cfg.Verbose,
			},
		},
		HideHelp:    false,
		HideVersion: false,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			hachyboopInstance := service.NewHachyboop()
			hachyboopInstance.Options = cfg

			// TODO from config
			cfg.Resolvers = []string{
				"91.200.176.1:53", // kiki.bunny.net
				"8.8.8.8:53",
			}

			return hachyboopInstance.Run()
		},
	}

	var err error

	logrus.Info("==========================================================================")
	for _, line := range strings.Split(banner, "\n") {
		logrus.Info(line)
	}
	logrus.Info("==========================================================================")

	logrus.Debugf("Parsing config")

	// Load environment variables
	err = Environment()
	if err != nil {
		logrus.Error(err)
		os.Exit(99)
	}

	// Arbitrary (non-error) pre load
	BeforeAppRun()

	logrus.Debugf("Entering main app loop")

	// Runtime
	err = app.Run(context.Background(), os.Args)
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	AfterAppRun()
}

// BeforeAppRun will run for ALL commands, and is used
// to initalize the runtime environments of the program.
func BeforeAppRun() {
	/* Flag parsing */
	if cfg.Verbose {
		logrus.SetLevel(logrus.TraceLevel)
	} else {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func AfterAppRun() {

}
