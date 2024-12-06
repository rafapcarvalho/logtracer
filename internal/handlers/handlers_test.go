package handlers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStdoutJSON(t *testing.T) {
	handler := StdoutJSON()
	assert.NotNil(t, handler)
}

func TestStdoutTXT(t *testing.T) {
	handler := StdoutTXT()
	assert.NotNil(t, handler)
}
