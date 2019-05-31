package controller

import (
	"net/url"
	"strings"
	"testing"

	"github.com/dbsystel/kibana-config-controller/kibana"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	"github.com/stretchr/testify/assert"
)

func TestSearchIDFromJSON(t *testing.T) {
	assert := assert.New(t)

	url, _ := url.Parse("https://example.com")
	logcfg := opslog.Config{LogLevel: "debug", LogFormat: "json"}
	logger, err := opslog.New(logcfg)
	if err != nil {
		t.Errorf("could not create logger: %s", err)
	}
	kibanaAPI := kibana.New(url, 1, logger)
	c := New(*kibanaAPI, logger)

	var tests = []struct {
		input    string
		expected string
	}{
		{`{"id": "abc"}`, "abc"},
		{`{"other": "value","id": "abc"}`, "abc"},
		{`{"other": "value","_id": "abcd"}`, "abcd"},
		{`{"other": "value","foo": "bar"}`, ""},
		{`invalid json`, ""},
	}

	for _, test := range tests {
		json := strings.NewReader(test.input)
		assert.Equal(c.searchIDFromJSON(json), test.expected)
	}
}
