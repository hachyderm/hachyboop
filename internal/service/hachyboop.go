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
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/hachyderm/hachyboop/internal/dns"
	"github.com/hachyderm/hachyboop/pkg/api"
	"github.com/sirupsen/logrus"

	"github.com/xitongsys/parquet-go-source/local"
	"github.com/xitongsys/parquet-go-source/s3"
	"github.com/xitongsys/parquet-go/parquet"
	"github.com/xitongsys/parquet-go/source"
	"github.com/xitongsys/parquet-go/writer"
)

// Configuration options for Hachyboop.
type HachyboopOptions struct {
	Verbose                      bool
	S3Output                     *S3Options
	FileOutput                   *FileOptions
	RuntimeCloudProviderMetadata *RuntimeCloudProviderMetadata
	ObserverId                   string
	ObservationRegion            string
	QuestionsRaw                 string // raw input from env/args
	Questions                    []string
	ResolversRaw                 string
	Resolvers                    []string
	TestFrequencySecondsRaw      string
	TestFrequencySeconds         int

	ObservationHandler chan *HachyboopDnsObservation
}

// Configuration options for our S3 file output.
type S3Options struct {
	Endpoint   string
	Bucket     string
	Path       string
	AccessKey  string
	Secret     string
	EnabledRaw string
}

// Configuration options for our local file output.
type FileOptions struct {
	Path       string
	FileName   string
	EnabledRaw string

	ParquetFile   source.ParquetFile
	ParquetWriter *writer.ParquetWriter
}

// Mostly for runtime provider contextual info
type RuntimeCloudProviderMetadata struct {
	// eventually we'll do something magical/dynamic to choose a provider based on detected runtime. for now, keeping it simple.
	// Bunny stuff https://docs.bunny.net/docs/magic-containers-app-metadata
	// BUNNYNET_MC_REGION
	BunnyRegion string
	// BUNNYNET_MC_PODID
	BunnyPodId string
	// BUNNYNET_MC_ZONE
	BunnyZone string
	// BUNNYNET_MC_APPID
	BunnyAppId string
}

// True if we should output to S3.
func (s *S3Options) Enabled() bool {
	if s.EnabledRaw == "" {
		return false
	}

	res, err := strconv.ParseBool(s.EnabledRaw)

	if err != nil {
		logrus.WithError(err).WithField("input", s.EnabledRaw).Warn("Couldn't parse a bool from HACHYBOOP_S3_WRITER_ENABLED")
	}

	return res
}

// True if we should output to a local file.
func (f *FileOptions) Enabled() bool {
	if f.EnabledRaw == "" {
		return false
	}

	res, err := strconv.ParseBool(f.EnabledRaw)

	if err != nil {
		logrus.WithError(err).WithField("input", f.EnabledRaw).Warn("Couldn't parse a bool from HACHYBOOP_S3_WRITER_ENABLED")
	}

	return res
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

	if !hb.Options.FileOutput.Enabled() {
		logrus.Warn("Local parquet output is not enabled. Set HACHYBOOP_LOCAL_WRITER_ENABLED to 'true' to enable.")
	}

	if !hb.Options.S3Output.Enabled() {
		logrus.Warn("S3 parquet output is not enabled. Set HACHYBOOP_S3_WRITER_ENABLED to 'true' to enable.")
	}

	for Enabled {
		hb.queryResolvers(resolvers)

		logrus.WithField("seconds", hb.Options.TestFrequencySeconds).Debug("Sleeping")
		time.Sleep(time.Duration(hb.Options.TestFrequencySeconds) * time.Second)
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
	observations := []*HachyboopDnsObservation{}

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
			observations = append(observations, obs)
		}
	}

	// TODO make this more dynamic/clean, maybe parallel
	if hb.Options.FileOutput.Enabled() {
		hb.writeObservationsToLocalFile(observations)
	}

	if hb.Options.S3Output.Enabled() {
		hb.writeObservationsToS3(observations)
	}
}

func (hb *Hachyboop) writeObservationsToLocalFile(observations []*HachyboopDnsObservation) {
	logrus.Debug("Prepping local file writer")
	path := filepath.Join(hb.Options.FileOutput.Path, time.Now().UTC().Format("2006-01-02T15.04.05.parquet"))
	logrus.WithField("filepath", path).Debug("Parquet local path prepared")

	fw, err := local.NewLocalFileWriter(path)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create local file writer")
	}
	defer fw.Close()

	pw, err := hb.newParquetFileWriter(fw)
	defer func() {
		logrus.Debug("Writing parquet footer to local file")
		if err := pw.WriteStop(); err != nil {
			logrus.WithError(err).Error("Failed to write footer to local file, parquet file likely corrupted")
		}
	}()

	logrus.Debug("Writing parquet data to local file")
	for _, obs := range observations {
		if err = pw.Write(obs); err != nil {
			logrus.WithError(err).Error("Failed to write parquet to local file")
		}
	}
}

func (hb *Hachyboop) writeObservationsToS3(observations []*HachyboopDnsObservation) {
	logrus.Debug("Preparing S3 file writer")
	path := filepath.Join(hb.Options.S3Output.Path, hb.Options.ObserverId, time.Now().UTC().Format("2006-01-02T15.04.05.parquet"))
	logrus.WithField("s3path", "s3://"+filepath.Join(hb.Options.S3Output.Bucket, path)).Debug("S3 output path prepared")

	awsCfg := &aws.Config{
		Region:      aws.String("US"),
		Credentials: credentials.NewStaticCredentials(hb.Options.S3Output.AccessKey, hb.Options.S3Output.Secret, ""),
		Endpoint:    aws.String(hb.Options.S3Output.Endpoint),
	}

	// TODO plumb context
	fw, err := s3.NewS3FileWriter(context.Background(), hb.Options.S3Output.Bucket, path, "bucket-owner-full-control", nil, awsCfg)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create S3 writer")
	}
	defer fw.Close()

	pw, err := hb.newParquetFileWriter(fw)
	defer func() {
		logrus.Debug("Writing parquet footer to S3")
		if err := pw.WriteStop(); err != nil {
			logrus.WithError(err).Error("Failed to write footer to S3, parquet file likely corrupted")
		}
	}()

	logrus.Debug("Writing parquet data to S3")
	for _, obs := range observations {
		if err = pw.Write(obs); err != nil {
			logrus.WithError(err).Error("Failed to write parquet to S3")
		}
	}
}

func (hb *Hachyboop) newParquetFileWriter(fw source.ParquetFile) (*writer.ParquetWriter, error) {
	logrus.Debug("Prepping parquet writer")
	pw, err := writer.NewParquetWriter(fw, new(HachyboopDnsObservation), 4)
	if err != nil {
		return nil, fmt.Errorf("Failed to created ParquetWriter: %w", err)
	}

	// TODO configurable
	pw.RowGroupSize = 128 * 1024 * 1024 //128M
	pw.PageSize = 8 * 1024              //8K
	pw.CompressionType = parquet.CompressionCodec_SNAPPY

	return pw, nil
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
