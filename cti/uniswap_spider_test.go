package cti

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTokens(t *testing.T) {
	tokens, err := GetTokens("ETH")
	assert.Equal(t, err, nil)
	var ts []string
	for _, t := range tokens {
		fmt.Printf("%v \n", t)
		ts = append(ts, t.ID)
	}
	pairs, err := GetParis(ts)
	assert.Equal(t, err, nil)
	for _, p := range pairs {
		fmt.Printf("pair:%v \n", p)
	}
}
