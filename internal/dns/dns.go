package dns

import (
	"context"
	"net"
	"time"
)

type TargetedResolver struct {
	Host string
	Port string

	Resolver *net.Resolver
}

type DnsResponse struct {
	ObservedOn              time.Time
	ObservedOnUnixTimestamp int64    `parquet:"name=observedonunixtimestamp, type=INT64, convertedtype=TIMESTAMP_MILLIS"`
	Host                    string   `parquet:"name=host, type=BYTE_ARRAY"`
	RecordType              string   `parquet:"name=recordtype, type=BYTE_ARRAY"`
	Values                  []string `parquet:"name=values, type=MAP, convertedtype=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8"`
	Error                   string   `parquet:"name=error, type=BYTE_ARRAY"`
	ResovledByHost          string   `parquet:"name=resolvedby, type=BYTE_ARRAY"`
	ResolvedBy              *TargetedResolver
}

func NewTargetedResolver(host, port string, timeoutSeconds int) *TargetedResolver {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(timeoutSeconds),
			}
			return d.DialContext(ctx, network, net.JoinHostPort(host, port))
		},
	}

	return &TargetedResolver{
		Host:     host,
		Port:     port,
		Resolver: resolver,
	}
}

func (tr *TargetedResolver) Lookup(ctx context.Context, host, recordType string) (*DnsResponse, error) {
	responses, err := tr.Resolver.LookupHost(ctx, host)

	var errorText string

	if err != nil {
		errorText = err.Error()
	}

	observedOn := time.Now().UTC()

	return &DnsResponse{
		Host:                    host,
		Values:                  responses,
		Error:                   errorText,
		ResolvedBy:              tr,
		ResovledByHost:          tr.Host,
		ObservedOn:              observedOn,
		ObservedOnUnixTimestamp: observedOn.Local().UnixMilli(),
	}, err
}
