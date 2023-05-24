package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestV1Port(t *testing.T) {
	assert.Equal(t, ":0", port(map[string]string{}))
	assert.Equal(t, ":4242", port(map[string]string{"port": "4242"}))
}
