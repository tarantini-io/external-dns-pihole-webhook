package pihole

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sigs.k8s.io/external-dns/endpoint"
)

func (suite *PiholeTestSuite) TestAAAARecordsEndpoint() {
	t := suite.T()
	server := suite.authedServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/config/dns/hosts", r.URL.Path)
	})
	defer server.Close()

	client, _ := newPiholeClient(Config{
		Server:   server.URL,
		Password: "password",
	})

	_, _ = client.listRecords(context.Background(), "AAAA")
}

func (suite *PiholeTestSuite) TestListAAAAEndpoints() {
	t := suite.T()
	server := suite.newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(RecordsResponse{
			Took: 0.123,
			Config: RecordsConfig{
				DNS: DNS{
					Hosts: []string{
						"1.1.1.1 test-one.example.io ",
						"1.1.1.1 test-one.example.com",
						"2.2.2.2 test-two.example.io",
						"b29f:3008:3ac4:753e:d124:f276:b92f:5d91 test-three.example.io",
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
			Filters: []string{"test-one.example.io", "test-three.example.io"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	records, err := cl.listRecords(context.Background(), "AAAA")

	assert.Len(t, records, 1)
	assert.Equal(t, records[0].DNSName, "test-three.example.io")
	assert.Equal(t, records[0].Targets[0], "b29f:3008:3ac4:753e:d124:f276:b92f:5d91")
}

func (suite *PiholeTestSuite) TestCreateAAAARecord() {
	t := suite.T()
	server := suite.authedServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/config/dns/hosts/b29f:3008:3ac4:753e:d124:f276:b92f:5d91 test-one.example.io", r.URL.Path)
	})
	defer server.Close()

	client, _ := newPiholeClient(Config{
		Server:   server.URL,
		Password: "password",
	})

	_ = client.createRecord(context.Background(), &endpoint.Endpoint{
		Targets:    []string{"b29f:3008:3ac4:753e:d124:f276:b92f:5d91"},
		DNSName:    "test-one.example.io",
		RecordType: endpoint.RecordTypeAAAA,
	})
}
