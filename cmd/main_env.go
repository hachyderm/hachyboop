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

var env = &EnvironmentOptions{}

type EnvironmentOptions struct {

	// example fields
	s3Host      string
	s3Path      string
	s3AccessKey string
	s3Secret    string
}

var (
	envOpt   = &EnvironmentOptions{}
	registry = []*EnvironmentVariable{
		{
			Name:        "HACHYBOOP_S3_HOST",
			Value:       "",
			Destination: &envOpt.s3Host,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_PATH",
			Value:       "",
			Destination: &envOpt.s3Path,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_ACCESS_KEY",
			Value:       "",
			Destination: &envOpt.s3AccessKey,
			Required:    false,
		},
		{
			Name:        "HACHYBOOP_S3_SECRET",
			Value:       "",
			Destination: &envOpt.s3Secret,
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
		v.Value = os.Getenv(v.Name)
		if v.Required && v.Value == "" {
			// If required and the variable is empty
			return fmt.Errorf("empty or undefined environmental variable [%s]", v.Name)
		}
		*v.Destination = v.Value
	}
	return nil
}
