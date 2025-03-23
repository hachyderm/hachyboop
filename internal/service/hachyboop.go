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
	Verbose    bool
	Resolvers  []string
	S3Output   *S3Options
	FileOutput *FileOptions
}

type S3Options struct {
	Host      string
	Path      string
	AccessKey string
	Secret    string
}

type FileOptions struct {
	Path string
}

func (s *S3Options) Enabled() bool {
	return s.Host != ""
}

func (f *FileOptions) Enabled() bool {
	return f.Path != ""
}

// Compile check *Nova implements Runner interface
var _ api.Runner = &Hachyboop{}

type Hachyboop struct {
	// Fields
	Options *HachyboopOptions
}

func NewHachyboop() *Hachyboop {
	return &Hachyboop{}
}

var (
	Enabled bool = true
)

func (hb *Hachyboop) Run() error {
	resolvers := hb.parseResolvers()

	for Enabled {
		queryResolvers(resolvers)

		// TODO from config
		time.Sleep(30 * time.Second)
	}
	return nil
}

func (hb *Hachyboop) parseResolvers() []*dns.TargetedResolver {
	var resolvers []*dns.TargetedResolver
	for _, resolverSpec := range hb.Options.Resolvers {
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
	return resolvers
}

func queryResolvers(resolvers []*dns.TargetedResolver) {
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
}

type HachyboopDnsObservation struct {
	ObservedOn              time.Time
	ObservedOnUnixTimestamp int64    `parquet:"name=observedonunixtimestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	Host                    string   `parquet:"name=host, type=BYTE_ARRAY"`
	RecordType              string   `parquet:"name=recordtype, type=BYTE_ARRAY"`
	Values                  []string `parquet:"name=values, type=MAP, convertedtype=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8"`
	Error                   string   `parquet:"name=error, type=BYTE_ARRAY"`
	ResovledByHost          string   `parquet:"name=resolvedby, type=BYTE_ARRAY"`
	ResolvedBy              *dns.TargetedResolver
}
