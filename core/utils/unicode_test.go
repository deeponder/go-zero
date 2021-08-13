package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGBKToUTF8(t *testing.T) {
	assert.Equal(t, "abc", GBKToUTF8("abc"))
}

func TestUTF8ToGBK(t *testing.T) {
	assert.Equal(t, "abc", UTF8ToGBK("abc"))
}
