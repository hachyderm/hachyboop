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
	"path/filepath"
	"strings"
	"time"

	"github.com/hachyderm/hachyboop/internal/dns"
	"github.com/hachyderm/hachyboop/pkg/api"
	"github.com/sirupsen/logrus"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"
)

// Configuration options for Hachyboop.
type HachyboopOptions struct {
	Verbose           bool
	S3Output          *S3Options
	FileOutput        *FileOptions
	ObserverId        string
	ObservationRegion string
	QuestionsRaw      string // raw input from env/args
	Questions         []string
	ResolversRaw      string
	Resolvers         []string

	ObservationHandler chan *HachyboopDnsObservation
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
	Path          string
	FileName      string
	ParquetFile   source.ParquetFile
	ParquetWriter *writer.ParquetWriter
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

func (hb *Hachyboop) Interrupt() {
	logrus.Warn("Got system interrupt, closing down")

	if hb.Options.FileOutput.ParquetWriter != nil {
		logrus.Info("Calling write stop on ParquetWriter")

		if err := hb.Options.FileOutput.ParquetWriter.WriteStop(); err != nil {
			logrus.WithError(err).Error("Failed to close parquet file")
		}
	}
}

// Entrypoint for Hachyboop!
func (hb *Hachyboop) Run() error {

	err := hb.handleFileOuptutOptions()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse local file options")
	}

	// TODO do this at app startup and store in hb config
	resolvers := hb.parseResolvers()

	for Enabled {
		hb.queryResolvers(resolvers)

		// TODO from config
		time.Sleep(10 * time.Second)
	}
	return nil
}

func (hb *Hachyboop) handleFileOuptutOptions() error {
	if hb.Options.FileOutput != nil {
		if hb.Options.FileOutput.Path != "" {
			// If no filename provided, make one
			if hb.Options.FileOutput.FileName == "" {
				// Autogenerate a file name
				hb.Options.FileOutput.FileName = time.Now().UTC().Format("2006-01-02T15.04.05.parquet")
			}
		}
	}

	if hb.Options.FileOutput.Enabled() {

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
func (hb *Hachyboop) queryResolvers(resolvers []*dns.TargetedResolver) {
	logrus.Debug("Prepping local file writer")

	path := filepath.Join(hb.Options.FileOutput.Path, time.Now().UTC().Format("2006-01-02T15.04.05.parquet"))
	logrus.WithField("filepath", path).Info("Parquet output path prepared")

	logrus.Debug("preparing local file writer")
	fw, err := local.NewLocalFileWriter(path)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create local file writer")
	}
	defer fw.Close()

	logrus.Debug("Prepping parquet writer")
	pw, err := writer.NewParquetWriter(fw, new(HachyboopDnsObservation), 4)
	if err != nil {
		logrus.WithError(err).Fatal("Couldn't create parquet writer")
	}
	defer pw.WriteStop()

	// TODO configurable
	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.PageSize = 8 * 1024              //8K
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	for _, resolver := range resolvers {
		for _, question := range hb.Options.Questions {
			// TODO impl record type (or get rid of it)
			response, err := resolver.Lookup(context.Background(), question, "A")

			// TODO move this async

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

			obs := hb.NewHachyboopDnsObservationFromDnsResponse(response)
			hb.Options.ObservationHandler <- obs

			logrus.WithField("obs", obs).Debug("converted object")

			if err = pw.Write(obs); err != nil {
				logrus.WithError(err).Error("failed to write parquet")
			}
		}
	}

	if err := pw.WriteStop(); err != nil {
		logrus.WithError(err).Error("Failed to close parquet file")
	}
}

// Model object suitable for serializing to parquet
type HachyboopDnsObservation struct {
	ObservedOnUnixTimestamp int64    `parquet:"name=observedonunixtimestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	ObservedBy              string   `parquet:"name=observedby, type=BYTE_ARRAY"`
	ObservedFromRegion      string   `parquet:"name=observedfromregion, type=BYTE_ARRAY"`
	Host                    string   `parquet:"name=host, type=BYTE_ARRAY"`
	RecordType              string   `parquet:"name=recordtype, type=BYTE_ARRAY"`
	Values                  []string `parquet:"name=values, type=LIST, valuetype=BYTE_ARRAY"`
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
