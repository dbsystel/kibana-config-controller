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


type APIClient struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
	Id         int
	logger     log.Logger
}

type KibanaFindResp struct {
	Total int `json:"total"`
	Data []struct{ Id string `json:"id"`} `json:"data"`
}

func (c *APIClient) CreateObject(objType, objId string, dataJSON io.Reader) error {
	return c.doPost(makeUrl(c.BaseUrl, "api/saved_objects/" + objType + "/" + objId), dataJSON)
}

func (c *APIClient) UpdateObject(objType, objId string, dataJSON io.Reader) error {
	url := makeUrl(c.BaseUrl, "api/saved_objects/" + objType + "/" + objId)
	req, err := http.NewRequest("PUT", url, dataJSON)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("kbn-xsrf", "true")

	return c.doRequest(req)
}

func (c *APIClient) DeleteObject(objType, objId string) error {
	url := makeUrl(c.BaseUrl, "api/saved_objects/" + objType + "/" + objId)
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
			level.Error(c.logger).Log("err", err.Error())
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
			        "Unexpected status code returned from Kibana API (got: %d, expected: 200, msg:%s)",
			               resp.StatusCode,
			               string(response),

			               )
	}
	return nil
}

type Clientset struct {
	BaseUrl    *url.URL
	HTTPClient *http.Client
}

func New(baseUrl *url.URL, id int, logger log.Logger) *APIClient {
	return &APIClient{
		BaseUrl:    baseUrl,
		HTTPClient: http.DefaultClient,
		Id:         id,
		logger:     logger,
	}
}

func makeUrl(baseURL *url.URL, endpoint string) string {
	result := *baseURL

	result.Path = path.Join(result.Path, endpoint)

	return result.String()
}


