package controller

import (
	"net/url"
	"strings"
	"testing"

	"github.com/dbsystel/kibana-config-controller/kibana"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestController(t *testing.T) *Controller {
	url, _ := url.Parse("https://example.com")
	logcfg := opslog.Config{LogLevel: "debug", LogFormat: "json"}
	logger, err := opslog.New(logcfg)
	if err != nil {
		t.Errorf("could not create logger: %s", err)
	}
	kibanaAPI := kibana.New(url, 1, logger)

	return New(*kibanaAPI, logger)
}

func TestSearchIDFromJSON(t *testing.T) {
	assert := assert.New(t)
	c := newTestController(t)

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

func TestSearchTypeFromJSON(t *testing.T) {
	assert := assert.New(t)
	c := newTestController(t)

	var tests = []struct {
		input    string
		expected string
	}{
		{`{"type": "abc"}`, "abc"},
		{`{"other": "value","type": "abc"}`, "abc"},
		{`{"other": "value","_type": "abcd"}`, "abcd"},
		{`{"other": "value","foo": "bar"}`, ""},
		{`invalid json`, ""},
	}

	for _, test := range tests {
		json := strings.NewReader(test.input)
		assert.Equal(c.searchTypeFromJSON(json), test.expected)
	}
}

func TestNoDifference(t *testing.T) {
	assert := assert.New(t)

	var tests = []struct {
		description string
		c1          *v1.ConfigMap
		c2          *v1.ConfigMap
		expected    bool
	}{
		{
			"equal data",
			&v1.ConfigMap{Data: map[string]string{"a": "b"}},
			&v1.ConfigMap{Data: map[string]string{"a": "b"}},
			true,
		},
		{
			"equal data and annotations",
			&v1.ConfigMap{Data: map[string]string{"a": "b"}, ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"c": "d"}}},
			&v1.ConfigMap{Data: map[string]string{"a": "b"}, ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"c": "d"}}},
			true,
		},
		{
			"unequal data",
			&v1.ConfigMap{Data: map[string]string{"a": "b"}},
			&v1.ConfigMap{Data: map[string]string{"a": "g"}},
			false,
		},
		{
			"equal data but unequal annotations",
			&v1.ConfigMap{Data: map[string]string{"a": "b"}, ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"c": "d"}}},
			&v1.ConfigMap{Data: map[string]string{"a": "b"}, ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"c": "e"}}},
			false,
		},
	}

	for _, test := range tests {
		assert.Equal(noDifference(test.c1, test.c2), test.expected, test.description)
	}
}
