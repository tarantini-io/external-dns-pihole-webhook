package pihole

import (
	"context"
	"github.com/scaleway/scaleway-sdk-go/logger"
	"sigs.k8s.io/external-dns/endpoint"
	"sigs.k8s.io/external-dns/plan"
	"sigs.k8s.io/external-dns/provider"
)

type PiholeProvider struct {
	provider.BaseProvider
	api piholeApi
}

// NewPiholeProvider initializes a new PiHole Local DNS based Provider
func NewPiholeProvider(cfg Config) (*PiholeProvider, error) {
	api, err := newPiholeClient(cfg)
	if err != nil {
		return nil, err
	}
	return &PiholeProvider{api: api}, nil
}

func (p *PiholeProvider) Records(ctx context.Context) ([]*endpoint.Endpoint, error) {
	aRecords, err := p.api.listRecords(ctx, endpoint.RecordTypeA)
	if err != nil {
		return nil, err
	}
	aaaRecords, err := p.api.listRecords(ctx, endpoint.RecordTypeAAAA)
	if err != nil {
		return nil, err
	}
	cnameRecords, err := p.api.listRecords(ctx, endpoint.RecordTypeCNAME)
	if err != nil {
		return nil, err
	}
	aRecords = append(aRecords, aaaRecords...)
	return append(aRecords, cnameRecords...), nil
}

func (p *PiholeProvider) ApplyChanges(ctx context.Context, changes *plan.Changes) error {
	for _, ep := range changes.Delete {
		if err := p.api.deleteRecord(ctx, ep); err != nil {
			logger.Errorf("error deleting record %s: %v", ep.DNSName, err)
			return err
		}
	}

	updateNew := make(map[piholeEntryKey]*endpoint.Endpoint)
	for _, ep := range changes.UpdateNew {
		key := piholeEntryKey{ep.DNSName, ep.RecordType}
		updateNew[key] = ep
	}

	for _, ep := range changes.UpdateOld {
		key := piholeEntryKey{ep.DNSName, ep.RecordType}
		if newRecord := updateNew[key]; newRecord != nil {
			if newRecord.Targets[0] == ep.Targets[0] {
				delete(updateNew, key)
				continue
			}
		}
		if err := p.api.deleteRecord(ctx, ep); err != nil {
			logger.Errorf("error deleting record %s: %v", ep.DNSName, err)
			return err
		}
	}

	for _, ep := range changes.Create {
		if err := p.api.createRecord(ctx, ep); err != nil {
			logger.Errorf("error creating record %s: %v", ep.DNSName, err)
			return err
		}
	}
	for _, ep := range updateNew {
		if err := p.api.createRecord(ctx, ep); err != nil {
			logger.Errorf("error creating record %s: %v", ep.DNSName, err)
			return err
		}
	}

	return nil
}
