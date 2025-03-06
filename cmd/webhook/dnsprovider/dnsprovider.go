package dnsprovider

import (
	"fmt"
	"github.com/tarantini-io/external-dns-pihole-webhook/cmd/webhook/configuration"
	"github.com/tarantini-io/external-dns-pihole-webhook/cmd/webhook/log"
	"github.com/tarantini-io/external-dns-pihole-webhook/internal/pihole"
	"regexp"
	"strings"

	"github.com/caarlos0/env/v11"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/provider"
)

type PiholeProviderFactory func(baseProvider *provider.BaseProvider, piholeConfig *pihole.Config) provider.Provider

func Init(config configuration.Config) (provider.Provider, error) {
	var domainFilter endpoint.DomainFilter
	createMsg := "creating pihole provider with "

	if config.RegexDomainFilter != "" {
		createMsg += fmt.Sprintf("regexp domain filter: '%s', ", config.RegexDomainFilter)
		if config.RegexDomainExclusion != "" {
			createMsg += fmt.Sprintf("with exclusion: '%s', ", config.RegexDomainExclusion)
		}
		domainFilter = endpoint.NewRegexDomainFilter(
			regexp.MustCompile(config.RegexDomainFilter),
			regexp.MustCompile(config.RegexDomainExclusion),
		)
	} else {
		if config.DomainFilter != nil && len(config.DomainFilter) > 0 {
			createMsg += fmt.Sprintf("domain filter: '%s', ", strings.Join(config.DomainFilter, ","))
		}
		if config.ExcludeDomains != nil && len(config.ExcludeDomains) > 0 {
			createMsg += fmt.Sprintf("exclude domain filter: '%s', ", strings.Join(config.ExcludeDomains, ","))
		}
		domainFilter = endpoint.NewDomainFilterWithExclusions(config.DomainFilter, config.ExcludeDomains)
	}

	createMsg = strings.TrimSuffix(createMsg, ", ")
	if strings.HasSuffix(createMsg, "with ") {
		createMsg += "no kind of domain filters"
	}
	log.Info(createMsg)

	piholeConfig := pihole.Config{}
	if err := env.Parse(&piholeConfig); err != nil {
		return nil, fmt.Errorf("reading configuration failed: %v", err)
	}
	piholeConfig.DomainFilter = domainFilter

	return pihole.NewPiholeProvider(piholeConfig)
}
