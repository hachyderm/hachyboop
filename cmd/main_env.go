/*===========================================================================*\
 *           MIT License Copyright (c) 2022 Kris Nóva <kris@nivenly.com>     *
 *                                                                           *
 *                ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓                *
 *                ┃   ███╗   ██╗ ██████╗ ██╗   ██╗ █████╗   ┃                *
 *                ┃   ████╗  ██║██╔═████╗██║   ██║██╔══██╗  ┃                *
 *                ┃   ██╔██╗ ██║██║██╔██║██║   ██║███████║  ┃                *
 *                ┃   ██║╚██╗██║████╔╝██║╚██╗ ██╔╝██╔══██║  ┃                *
 *                ┃   ██║ ╚████║╚██████╔╝ ╚████╔╝ ██║  ██║  ┃                *
 *                ┃   ╚═╝  ╚═══╝ ╚═════╝   ╚═══╝  ╚═╝  ╚═╝  ┃                *
 *                ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛                *
 *                                                                           *
 *                       This machine kills fascists.                        *
 *                                                                           *
\*===========================================================================*/

package main

import (
	"fmt"
	"os"
)

var (
	registry = []*EnvironmentVariable{
		{
			Name:        "HACHYBOOP_S3_WRITER_ENABLED",
			Value:       "false",
			Destination: &cfg.S3Output.EnabledRaw,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_ENDPOINT",
			Value:       "",
			Destination: &cfg.S3Output.Endpoint,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_BUCKET",
			Value:       "",
			Destination: &cfg.S3Output.Bucket,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_PATH",
			Value:       "",
			Destination: &cfg.S3Output.Path,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_ACCESS_KEY_ID",
			Value:       "",
			Destination: &cfg.S3Output.AccessKey,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_SECRET_ACCESS_KEY",
			Value:       "",
			Destination: &cfg.S3Output.Secret,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_LOCAL_WRITER_ENABLED",
			Value:       "false",
			Destination: &cfg.FileOutput.EnabledRaw,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_LOCAL_RESULTS_PATH",
			Value:       "data",
			Destination: &cfg.FileOutput.Path,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_LOCAL_RESULTS_FILE_NAME",
			Value:       "data",
			Destination: &cfg.FileOutput.FileName,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_OBSERVER_ID",
			Value:       "esk",
			Destination: &cfg.ObserverId,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_OBSERVER_REGION",
			Value:       "namer-central",
			Destination: &cfg.ObservationRegion,
			Required:    false,
		},
		{
			Name:        "BUNNYNET_MC_REGION",
			Destination: &cfg.RuntimeCloudProviderMetadata.BunnyRegion,
			Value:       "",
			Required:    false,
		},
		{
			Name:        "BUNNYNET_MC_PODID",
			Destination: &cfg.RuntimeCloudProviderMetadata.BunnyPodId,
			Value:       "",
			Required:    false,
		},
		{
			Name:        "BUNNYNET_MC_APPID",
			Destination: &cfg.RuntimeCloudProviderMetadata.BunnyAppId,
			Value:       "",
			Required:    false,
		},
		{
			Name:        "BUNNYNET_MC_ZONE",
			Destination: &cfg.RuntimeCloudProviderMetadata.BunnyZone,
			Value:       "",
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_QUESTIONS",
			Value:       "hachyderm.io",
			Destination: &cfg.QuestionsRaw,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_RESOLVERS",
			Value:       "hachyderm.io",
			Destination: &cfg.ResolversRaw,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_TEST_FREQUENCY_SECONDS",
			Value:       "300",
			Destination: &cfg.TestFrequencySecondsRaw,
			Required:    false,
		},
	}
)

type EnvironmentVariable struct {
	Name        string
	Value       string
	Destination *string
	Required    bool
}

func Environment() error {
	for _, v := range registry {
		readValue := os.Getenv(v.Name)
		if v.Required && readValue == "" {
			// If required and the variable is empty
			return fmt.Errorf("empty or undefined environmental variable [%s]", v.Name)
		}

		if readValue != "" {
			v.Value = readValue // we don't use v.Value but why not
			*v.Destination = readValue
		}
	}
	return nil
}
