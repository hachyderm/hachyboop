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

package service

import (
	"context"
	"strings"
	"time"

	"github.com/hachyderm/hachyboop/internal/dns"
	"github.com/hachyderm/hachyboop/pkg/api"
	"github.com/sirupsen/logrus"
)

type HachyboopOptions struct {
	Verbose   bool
	Resolvers []string
	S3Output  *S3Options
}

type S3Options struct {
	Host      string
	Path      string
	AccessKey string
	Secret    string
}

func (s *S3Options) Enabled() bool {
	return s.Host != ""
}

// Compile check *Nova implements Runner interface
var _ api.Runner = &Hachyboop{}

type Hachyboop struct {
	// Fields
	Config *HachyboopOptions
}

func NewHachyboop() *Hachyboop {
	return &Hachyboop{}
}

var (
	runtimeHachyboop bool = true
)

func (n *Hachyboop) Run() error {
	var resolvers []*dns.TargetedResolver
	for _, resolverSpec := range n.Config.Resolvers {
		parts := strings.Split(resolverSpec, ":")

		if len(parts) != 2 {
			logrus.WithField("resolverSpec", resolverSpec).Warn("Resolver must be provided as host:port, this resolver didn't fit that. Ignorning.")
			continue
		}

		host := parts[0]
		port := parts[1]

		// TODO take timeout via config
		resolver := dns.NewTargetedResolver(host, port, 5)

		resolvers = append(resolvers, resolver)
	}

	for runtimeHachyboop {
		// TODO extract this out

		for _, resolver := range resolvers {
			// TODO extract this out

			// TODO from config
			lookupHost := "hachyderm.io"

			// TODO impl record type (or get rid of it)
			response, err := resolver.Lookup(context.Background(), lookupHost, "A")

			logFields := logrus.Fields{
				"host":       response.Host,
				"response":   response.Values,
				"resolvedBy": response.ResolvedBy.Host,
			}

			if err != nil {
				logFields["error"] = err.Error()
				logrus.WithFields(logFields).Warnf("DNS lookup failed")
			} else {
				logrus.WithFields(logFields).Infof("DNS lookup completed")
			}
		}

		// TODO from config
		time.Sleep(30 * time.Second)
	}
	return nil
}
