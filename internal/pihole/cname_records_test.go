package pihole

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sigs.k8s.io/external-dns/endpoint"
)

func (suite *PiholeTestSuite) TestCnameRecordsEndpoint() {
	t := suite.T()
	server := suite.authedServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/config/dns/cnameRecords", r.URL.Path)
	})
	defer server.Close()

	client, _ := newPiholeClient(Config{
		Server:   server.URL,
		Password: "password",
	})

	_, _ = client.listRecords(context.Background(), "CNAME")
}

func (suite *PiholeTestSuite) TestListCnameEndpoints() {
	t := suite.T()
	server := suite.newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(RecordsResponse{
			Took: 0.123,
			Config: RecordsConfig{
				DNS: DNS{
					CnameRecords: []string{
						"test-one.example.io,test-one.proxy.io",
						"test-one.example.com,test-one.proxy.com",
						"test-two.example.io,test-two.proxy.io",
						"test-three.example.io,test-three.proxy.io",
					},
				},
			},
		})
	})

	defer server.Close()

	cl, err := newPiholeClient(Config{
		Server:   server.URL,
		Password: "correct",
		DomainFilter: endpoint.DomainFilter{
			Filters: []string{"test-one.example.io", "test-two.example.io"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	records, err := cl.listRecords(context.Background(), "CNAME")

	assert.Len(t, records, 2)
	assert.Equal(t, records[0].DNSName, "test-one.example.io")
	assert.Equal(t, records[0].Targets[0], "test-one.proxy.io")
	assert.Equal(t, records[1].DNSName, "test-two.example.io")
	assert.Equal(t, records[1].Targets[0], "test-two.proxy.io")
}

func (suite *PiholeTestSuite) TestCreateCnameRecord() {
	t := suite.T()
	server := suite.authedServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/config/dns/cnameRecords/test-one.example.io,proxy-one.example.io", r.URL.Path)
	})
	defer server.Close()

	client, _ := newPiholeClient(Config{
		Server:   server.URL,
		Password: "password",
	})

	_ = client.createRecord(context.Background(), &endpoint.Endpoint{
		Targets:    []string{"proxy-one.example.io"},
		DNSName:    "test-one.example.io",
		RecordType: endpoint.RecordTypeCNAME,
	})
}
