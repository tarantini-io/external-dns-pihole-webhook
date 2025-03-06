package pihole

import "sigs.k8s.io/external-dns/endpoint"

type Session struct {
	Valid bool   `json:"valid"`
	Sid   string `json:"sid"`
}

type DNS struct {
	Hosts        []string `json:"hosts"`
	CnameRecords []string `json:"cnameRecords"`
}

type RecordsConfig struct {
	DNS DNS `json:"dns"`
}

type Host struct {
	name   string
	target string
}

type Config struct {
	Server                string `env:"PIHOLE_SERVER" envDefault:"http://pi.hole:80"`
	Password              string `env:"PIHOLE_PASSWORD" envDefault:""`
	TLSInsecureSkipVerify bool   `env:"PIHOLE_TLS_INSECURE" envDefault:"false"`
	DryRun                bool   `env:"PIHOLE_DRY_RUN" envDefault:"false"`
	DomainFilter          endpoint.DomainFilter
}

type piholeEntryKey struct {
	Target     string
	RecordType string
}
