package kibana

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// APIClient wrapper for the api client to kibana
type APIClient struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	ID         int
	logger     log.Logger
}

// FindResp the kibana API response
type FindResp struct {
	Total int `json:"total"`
	Data  []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// CreateObject creates the given object
func (c *APIClient) CreateObject(objType, objID string, dataJSON io.Reader) error {
	return c.doPost(makeURL(c.BaseURL, "api/saved_objects/"+objType+"/"+objID), dataJSON)
}

// UpdateObject updates the given object
func (c *APIClient) UpdateObject(objType, objID string, dataJSON io.Reader) error {
	url := makeURL(c.BaseURL, "api/saved_objects/"+objType+"/"+objID)
	req, err := http.NewRequest("PUT", url, dataJSON)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("kbn-xsrf", "true")

	return c.doRequest(req)
}

// DeleteObject deletes the object with the given ID
func (c *APIClient) DeleteObject(objType, objID string) error {
	url := makeURL(c.BaseURL, "api/saved_objects/"+objType+"/"+objID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("kbn-xsrf", "true")

	return c.doRequest(req)
}

func (c *APIClient) doPost(url string, dataJSON io.Reader) error {
	req, err := http.NewRequest("POST", url, dataJSON)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("kbn-xsrf", "true")

	return c.doRequest(req)
}

func (c *APIClient) doRequest(req *http.Request) error {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		for strings.Contains(err.Error(), "connection refused") {
			//nolint:errcheck
			level.Error(c.logger).Log("err", err.Error())
			//nolint:errcheck
			level.Info(c.logger).Log("msg", "Perhaps Kibana is not ready. Waiting for 8 seconds and retry again...")
			time.Sleep(8 * time.Second)
			resp, err = c.HTTPClient.Do(req)
			if err == nil {
				break
			}
		}
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	response, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"unexpected status code returned from Kibana API (got: %d, expected: 200, msg:%s)",
			resp.StatusCode,
			string(response),
		)
	}
	return nil
}

// Clientset TODO: needed?
type Clientset struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
}

// New creates a kibana api client
func New(baseURL *url.URL, id int, logger log.Logger) *APIClient {
	return &APIClient{
		BaseURL:    baseURL,
		HTTPClient: http.DefaultClient,
		ID:         id,
		logger:     logger,
	}
}

func makeURL(baseURL *url.URL, endpoint string) string {
	result := *baseURL

	result.Path = path.Join(result.Path, endpoint)

	return result.String()
}
