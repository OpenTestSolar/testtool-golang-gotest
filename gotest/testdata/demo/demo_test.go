package gotest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	cases := []struct {
		Name           string
		A, B, Expected int
	}{
		{"pos", 2, 3, 5},
		{"neg", 2, -3, -1},
		{"zero", 2, 0, 2},
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			if value := Add(c.A, c.B); value != c.Expected {
				t.Errorf("calculate was wrong")
			}
		})
	}
}

func TestAdd1(t *testing.T) {
	assert.Equal(t, 2, Add(1, 1))
}

func TestAdd2(t *testing.T) {
	assert.Equal(t, 1, Add(1, 1))
}

func TestSlowFunc(t *testing.T) {
	ret := SlowFunc()
	assert.Equal(t, ret, 0)
}
