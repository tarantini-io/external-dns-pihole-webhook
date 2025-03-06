package pihole

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/scaleway/scaleway-sdk-go/logger"
	"net/http"
	"net/http/cookiejar"
	"sigs.k8s.io/external-dns/endpoint"
	"strings"
)

// piholeAPI declares the "API" actions performed against the Pihole server.
type piholeApi interface {
	// listRecords returns endpoints for the given record type (A or CNAME).
	listRecords(ctx context.Context, rtype string) ([]*endpoint.Endpoint, error)
	// createRecord will create a new record for the given endpoint.
	createRecord(ctx context.Context, ep *endpoint.Endpoint) error
	// deleteRecord will delete the given record.
	deleteRecord(ctx context.Context, ep *endpoint.Endpoint) error
}

// piholeClient implements the piholeAPI.
type piholeClient struct {
	cfg        Config
	httpClient *http.Client
	session    *Session
}

func (r *RecordsResponse) Records(rtype string) *[]Host {
	var result []Host
	if strings.EqualFold(rtype, endpoint.RecordTypeCNAME) {
		Map[string, Host](r.Config.DNS.CnameRecords, &result, func(s string) (Host, error) {
			split := strings.Split(s, ",")
			return Host{
				name:   split[0],
				target: split[1],
			}, nil
		})
	}
	if strings.EqualFold(rtype, endpoint.RecordTypeA) || strings.EqualFold(rtype, endpoint.RecordTypeAAAA) {
		Map[string, Host](r.Config.DNS.Hosts, &result, func(s string) (Host, error) {
			split := strings.Split(s, " ")
			return Host{
				name:   split[1],
				target: split[0],
			}, nil
		})
	}
	return &result
}

func pathForType(rtype string) (string, error) {
	switch rtype {
	case endpoint.RecordTypeCNAME:
		return "/config/dns/cnameRecords", nil
	case endpoint.RecordTypeA:
		return "/config/dns/hosts", nil
	case endpoint.RecordTypeAAAA:
		return "/config/dns/hosts", nil
	}
	return "", errors.New("Unknown RecordType")
}

func pathForEndpoint(ep *endpoint.Endpoint) (string, error) {
	var p string
	var split string
	switch ep.RecordType {
	case endpoint.RecordTypeCNAME:
		p = "cnameRecords"
		split = "%2C"
	case endpoint.RecordTypeA:
		p = "hosts"
		split = "%20"
	case endpoint.RecordTypeAAAA:
		p = "hosts"
		split = "%20"
	default:
		return "", errors.New("Unknown RecordType")
	}
	return fmt.Sprintf("/config/dns/%s/%s%s%s", p, ep.Targets[0], split, ep.DNSName), nil
}

// newPiholeClient creates a new Pihole API client.
func newPiholeClient(cfg Config) (piholeApi, error) {
	if cfg.Server == "" {
		return nil, ErrNoPiholeServer
	}

	// Set up a persistent cookiejar for storing session information
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}

	// Set up an HTTP client using the cookiejar
	httpClient := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.TLSInsecureSkipVerify,
			},
		},
	}

	p := &piholeClient{
		cfg:        cfg,
		httpClient: httpClient,
	}
	if err := p.retrieveNewToken(context.Background()); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *piholeClient) retrieveNewToken(ctx context.Context) error {
	if p.cfg.Password == "" {
		logger.Debugf("No password was supplied to External DNS")
	}

	var loginResponse LoginResponse
	if _, err := p.callPihole(ctx, http.MethodPost, "/auth", LoginRequest{Password: p.cfg.Password}, &loginResponse); err != nil {
		return err
	}
	p.session = &loginResponse.Session
	return nil
}

func (p *piholeClient) listRecords(ctx context.Context, rtype string) ([]*endpoint.Endpoint, error) {
	path, err := pathForType(rtype)
	if err != nil {
		return nil, err
	}

	var response RecordsResponse
	if _, err := p.callPihole(ctx, http.MethodGet, path, nil, &response); err != nil {
		return nil, err
	}
	var endpoints []*endpoint.Endpoint
	Map[Host, *endpoint.Endpoint](*response.Records(rtype), &endpoints, func(host Host) (*endpoint.Endpoint, error) {
		if !p.cfg.DomainFilter.Match(host.name) {
			logger.Debugf("Skipping record %s that does not match domain filter", host.name)
			return nil, errors.New("Skipping record that does not match domain filter")
		}

		if rtype == endpoint.RecordTypeA && strings.Contains(host.target, ":") {
			return nil, errors.New("Skipping AAAA record type")
		}

		if rtype == endpoint.RecordTypeAAAA && strings.Contains(host.target, ".") {
			return nil, errors.New("Skipping A record type")
		}

		return &endpoint.Endpoint{
			DNSName:    host.name,
			Targets:    []string{host.target},
			RecordType: rtype,
		}, nil
	})
	return endpoints, nil
}

func (p *piholeClient) createRecord(ctx context.Context, ep *endpoint.Endpoint) error {
	return p.manageRecord(ctx, http.MethodPut, ep)
}

func (p *piholeClient) deleteRecord(ctx context.Context, ep *endpoint.Endpoint) error {
	return p.manageRecord(ctx, http.MethodDelete, ep)
}

func (p *piholeClient) manageRecord(ctx context.Context, action string, ep *endpoint.Endpoint) error {
	if !p.cfg.DomainFilter.Match(ep.DNSName) {
		logger.Debugf("Skipping record %s that does not match domain filter", ep.DNSName)
		return nil
	}

	if p.cfg.DryRun {
		logger.Infof("DRY RUN: %s %s IN %s -> %s", action, ep.DNSName, ep.RecordType, ep.Targets[0])
		return nil
	}

	logger.Infof("%s %s IN %s -> %s", action, ep.DNSName, ep.RecordType, ep.Targets[0])

	path, err := pathForEndpoint(ep)
	if err != nil {
		return err
	}

	if _, err = p.callPihole(ctx, action, path, nil, nil); err != nil {
		return err
	}

	return nil
}

func (p *piholeClient) callPihole(ctx context.Context, method string, path string, body interface{}, response interface{}) (*http.Response, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/api%s", p.cfg.Server, path)

	logger.Debugf("Calling pihole %s %s", method, url)

	var req *http.Request
	if body == nil {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(jsonBody))
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if p.session != nil {
		req.Header.Set("sid", p.session.Sid)
	}

	res, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if path != "/auth" && res.StatusCode == http.StatusUnauthorized {
		if err := p.retrieveNewToken(ctx); err != nil {
			return nil, err
		}
		return p.callPihole(ctx, method, path, body, response)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("received non-200 status code from request: %s", res.Status)
	}

	if response != nil {
		err = json.NewDecoder(res.Body).Decode(response)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}
