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
	"fmt"
	"os"
	"strconv"
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
	S3Output:                     &service.S3Options{},
	FileOutput:                   &service.FileOptions{},
	RuntimeCloudProviderMetadata: &service.RuntimeCloudProviderMetadata{},
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

			cfg.Questions = strings.Split(cfg.QuestionsRaw, ",")
			cfg.Resolvers = strings.Split(cfg.ResolversRaw, ",")
			cfg.ObservationHandler = make(chan *service.HachyboopDnsObservation, 32)

			testSec, err := strconv.Atoi(cfg.TestFrequencySecondsRaw)
			if err != nil {
				logrus.WithField("rawValue", cfg.TestFrequencySecondsRaw).Warn("couldn't convert HACHYBOOP_TEST_FREQUENCY_SECONDS to an int, defaulting to 300")
				testSec = 300
			}
			cfg.TestFrequencySeconds = testSec
			logrus.WithField("seconds", cfg.TestFrequencySeconds).Info("Running tests every x seconds")

			// TODO need to clean the observer id to conform to S3 or reject if doesn't conform. also no /

			// TODO make this dynamic and less hardcoded
			if cfg.RuntimeCloudProviderMetadata.BunnyPodId != "" {
				logrus.Info("Cloud provider: Bunny Magic Container Runtime detected")
				cfg.ObservationRegion = fmt.Sprintf("%s/%s", cfg.RuntimeCloudProviderMetadata.BunnyRegion, cfg.RuntimeCloudProviderMetadata.BunnyZone)
				cfg.ObserverId = fmt.Sprintf("%s/%s", cfg.RuntimeCloudProviderMetadata.BunnyAppId, cfg.RuntimeCloudProviderMetadata.BunnyPodId)
			}

			logrus.WithField("observer", cfg.ObserverId).WithField("region", cfg.ObservationRegion).Debug("Set region and observer")

			// TODO validate at least one question & one resolver

			return hachyboopInstance.Run()
		},
	}

	var err error

	logrus.Info("==========================================================================")
	for _, line := range strings.Split(banner, "\n") {
		logrus.Info(line)
	}
	logrus.Info("==========================================================================")

	logrus.Debug("Parsing config")

	// Load environment variables
	err = Environment()
	if err != nil {
		logrus.Error(err)
		os.Exit(99)
	}

	// Arbitrary (non-error) pre load
	BeforeAppRun()

	logrus.Debug("Entering main app loop")

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
