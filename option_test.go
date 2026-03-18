package annotate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithStyle(t *testing.T) {
	r := New(WithStyle(DefaultStyle))
	assert.NotNil(t, r.Style.LineNumber)
	assert.NotNil(t, r.Style.Separator)
}

func TestNewWithoutOptions(t *testing.T) {
	r := New()
	assert.Nil(t, r.Style.LineNumber)
	assert.Nil(t, r.Style.Separator)
}
