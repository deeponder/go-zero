package netx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetLocalIP(t *testing.T) {
	assert.True(t, len(GetLocalIP()) > 0)
}

func TestGetU32LocalIP(t *testing.T) {
	assert.True(t, GetU32LocalIP() > 0)
}
