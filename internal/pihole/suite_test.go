package pihole

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type PiholeTestSuite struct {
	suite.Suite
	ctx    context.Context
	server *httptest.Server
}

func (suite *PiholeTestSuite) SetupSubTest() {
	suite.ctx = context.Background()
	suite.server = &httptest.Server{}
}

func (suite *PiholeTestSuite) newTestServer(hdlr http.HandlerFunc) *httptest.Server {
	suite.T().Helper()
	server := httptest.NewServer(hdlr)
	return server
}

func (suite *PiholeTestSuite) authedServer(hndlr http.HandlerFunc) *httptest.Server {
	return suite.newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth" {
			_ = json.NewEncoder(w).Encode(LoginResponse{
				Session: Session{
					Sid:   "sid",
					Valid: true,
				},
				Took: 0.123,
			})
			return
		}
		hndlr(w, r)
		return
	})
}

func TestPiholeTestSuite(t *testing.T) {
	suite.Run(t, new(PiholeTestSuite))
}
