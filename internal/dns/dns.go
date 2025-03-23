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
	ObservedOn time.Time
	Host       string
	RecordType string
	Values     []string
	Error      string
	ResolvedBy *TargetedResolver
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

	return &DnsResponse{
		Host:       host,
		Values:     responses,
		Error:      errorText,
		ResolvedBy: tr,
		ObservedOn: time.Now().UTC(),
	}, err
}
