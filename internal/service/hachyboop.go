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

// Configuration options for Hachyboop.
type HachyboopOptions struct {
	Verbose           bool
	Resolvers         []string
	S3Output          *S3Options
	FileOutput        *FileOptions
	ObserverId        string
	ObservationRegion string
	QueryTargetsRaw   string // raw input from env/args
}

// Configuration options for our S3 file output.
type S3Options struct {
	Host      string
	Path      string
	AccessKey string
	Secret    string
}

// Configuration options for our local file output.
type FileOptions struct {
	Path string
}

// True if we should output to S3.
func (s *S3Options) Enabled() bool {
	return s.Host != ""
}

// True if we should output to a local file.
func (f *FileOptions) Enabled() bool {
	return f.Path != ""
}

// Compile check *Hachyboop implements Runner interface
var _ api.Runner = &Hachyboop{}

// The Hachyboop! The worker class that does all the work.
type Hachyboop struct {
	// Fields
	Options *HachyboopOptions
}

func NewHachyboop() *Hachyboop {
	return &Hachyboop{}
}

var (
	// If the work loop should continue
	Enabled bool = true // TODO handle SIG*
)

// Entrypoint for Hachyboop!
func (hb *Hachyboop) Run() error {
	resolvers := hb.parseResolvers()

	for Enabled {
		queryResolvers(resolvers)

		// TODO from config
		time.Sleep(30 * time.Second)
	}
	return nil
}

// Convert our string representation of resovlers to instances
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

// Core work loop that queries the resolvers
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

// Model object suitable for serializing to parquet
type HachyboopDnsObservation struct {
	ObservedOnUnixTimestamp int64    `parquet:"name=observedonunixtimestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	ObservedBy              string   `parquet:"name=observedby, type=BYTE_ARRAY"`
	ObservedFromRegion      string   `parquet:"name=observedfromregion, type=BYTE_ARRAY"`
	Host                    string   `parquet:"name=host, type=BYTE_ARRAY"`
	RecordType              string   `parquet:"name=recordtype, type=BYTE_ARRAY"`
	Values                  []string `parquet:"name=values, type=MAP, convertedtype=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8"`
	Error                   string   `parquet:"name=error, type=BYTE_ARRAY"`
	ResovledByHost          string   `parquet:"name=resolvedby, type=BYTE_ARRAY"`
}

// Converts a DNS response into our model object ready to serialize to parquet
func (hb *Hachyboop) NewHachyboopDnsObservationFromDnsResponse(d *dns.DnsResponse) *HachyboopDnsObservation {
	return &HachyboopDnsObservation{
		ObservedOnUnixTimestamp: d.ObservedOn.UnixMilli(),
		Host:                    d.Host,
		RecordType:              d.RecordType,
		Values:                  d.Values,
		Error:                   d.Error,
		ResovledByHost:          d.ResolvedBy.Host,
		ObservedBy:              hb.Options.ObserverId,
		ObservedFromRegion:      hb.Options.ObservationRegion,
	}
}
