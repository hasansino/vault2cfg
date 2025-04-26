package vault2cfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type _testCfg struct {
	ValueA    string `vault:"TEST_VALUE_A"`
	ValueB    string `vault:"TEST_VALUE_B"`
	SubStruct struct {
		ValueC string `vault:"TEST_VALUE_C"`
	}
}

func TestBind(t *testing.T) {
	cfg := &_testCfg{}

	var data = map[string]interface{}{
		"TEST_VALUE_A": "foo",
		"TEST_VALUE_B": "bar",
		"TEST_VALUE_C": "sux",
	}

	Bind(cfg, data)
	assert.Equal(t, "foo", cfg.ValueA)
	assert.Equal(t, "bar", cfg.ValueB)
	assert.Equal(t, "sux", cfg.SubStruct.ValueC)
}
