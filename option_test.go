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

func TestWithBefore(t *testing.T) {
	r := New(WithBefore(3))
	assert.Equal(t, 3, r.Before)
}

func TestWithAfter(t *testing.T) {
	r := New(WithAfter(2))
	assert.Equal(t, 2, r.After)
}

func TestWithBeforeNegative(t *testing.T) {
	r := New(WithBefore(-1))
	assert.Equal(t, 0, r.Before)
}

func TestWithAfterNegative(t *testing.T) {
	r := New(WithAfter(-5))
	assert.Equal(t, 0, r.After)
}
