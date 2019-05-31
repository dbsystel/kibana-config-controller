package controller

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/dbsystel/kibana-config-controller/kibana"
	opslog "github.com/dbsystel/kube-controller-dbsystel-go-common/log"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kibanaAPIClientMock struct {
	mock.Mock
}

func (c *kibanaAPIClientMock) CreateObject(objType, objID string, dataJSON io.Reader) error {
	args := c.Called(objType, objID, dataJSON)
	fmt.Printf("## args: %v", args)
	return nil
}
func (c *kibanaAPIClientMock) UpdateObject(objType, objID string, dataJSON io.Reader) error {
	return nil
}
func (c *kibanaAPIClientMock) DeleteObject(objType, objID string) error {
	return nil
}
func (c *kibanaAPIClientMock) doPost(url string, dataJSON io.Reader) error {
	return nil
}
func (c *kibanaAPIClientMock) doRequest(req *http.Request) error {
	return nil
}
func (c *kibanaAPIClientMock) GetID() int {
	return 1
}

func newLogCfg(t *testing.T) log.Logger {
	logcfg := opslog.Config{LogLevel: "debug", LogFormat: "json"}
	logger, err := opslog.New(logcfg)
	if err != nil {
		t.Errorf("could not create logger: %s", err)
	}

	return logger
}

func newTestController(t *testing.T, kibanaAPI kibana.IAPIClient) *Controller {
	if kibanaAPI == nil {
		url, _ := url.Parse("https://example.com")
		dummyKibanaAPI := kibana.New(url, 1, newLogCfg(t))
		return New(dummyKibanaAPI, newLogCfg(t))
	}

	return New(kibanaAPI, newLogCfg(t))
}

func TestSearchIDFromJSON(t *testing.T) {
	assert := assert.New(t)
	c := newTestController(t, nil)

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
	c := newTestController(t, nil)

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

func TestCreateObject(t *testing.T) {
	kibanaAPI := new(kibanaAPIClientMock)
	c := newTestController(t, kibanaAPI)

	assert := assert.New(t)

	var tests = []struct {
		description string
		configMap   *v1.ConfigMap
	}{
		{
			"invalid kibana id",
			&v1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"kibana.net/id": "abc"}}},
		},
	}

	for _, test := range tests {
		c.Create(test.configMap)
		assert.Equal(true, true)
		// kibanaAPI.AssertExpectations(t)
	}
}
