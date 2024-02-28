package frontend

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBasicAuth(t *testing.T) {
	username := "user"
	password := "pass"
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))

	user, pass, ok := parseBasicAuth(auth)
	assert.Equal(t, username, user)
	assert.Equal(t, password, pass)
	assert.True(t, ok)
}
