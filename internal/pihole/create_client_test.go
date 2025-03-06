package pihole

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
)

func (suite *PiholeTestSuite) TestNoServerUrl() {
	t := suite.T()

	_, err := newPiholeClient(Config{})

	assert.NotNil(t, err, "NewPiholeClient should return an error with no server url")
	assert.Error(t, err, "Error should be ErrNoPiholeServer")
}

func (suite *PiholeTestSuite) TestAuthEndpoint() {
	t := suite.T()
	server := suite.newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var bod LoginRequest
		_ = json.NewDecoder(r.Body).Decode(&bod)
		assert.Equal(t, "/api/auth", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, LoginRequest{Password: "password"}, bod)
		return
	})

	defer server.Close()

	_, err := newPiholeClient(Config{
		Server:   server.URL,
		Password: "password",
	})

	assert.NotNil(t, err, "NewPiholeClient should return an error with incorrect password")
}

func (suite *PiholeTestSuite) TestIncorrectPassword() {
	t := suite.T()
	server := suite.newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	})

	defer server.Close()

	_, err := newPiholeClient(Config{
		Server:   server.URL,
		Password: "incorrect",
	})

	assert.NotNil(t, err, "NewPiholeClient should return an error with incorrect password")
}

func (suite *PiholeTestSuite) TestCorrectPassword() {
	t := suite.T()
	server := suite.newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LoginResponse{
			Session: Session{
				Sid:   "sid",
				Valid: true,
			},
			Took: 0.234,
		})
	})

	defer server.Close()

	cl, err := newPiholeClient(Config{
		Server:   server.URL,
		Password: "correct",
	})

	assert.Nil(t, err, "NewPiholeClient should not return an error")
	assert.NotNil(t, cl, "NewPiholeClient should not be null")
	assert.Equal(t, "sid", cl.(*piholeClient).session.Sid, "NewPiholeClient should have correct sid")
}
