package wiremock

import (
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"resty.dev/v3"
)

type Client struct {
	HostAddr string
}

type AdminMap struct {
	Mappings []AdminMapping `json:"mappings"`
	Meta     *AdminMeta     `json:"meta"`
}

type AdminMapping struct {
	Id           uuid.UUID        `json:"id"`
	Name         string           `json:"name"`
	Request      *MappingRequest  `json:"request"`
	Response     *MappingResponse `json:"response"`
	ScenarioName string           `json:"scenarioName"`
}

type AdminMeta struct {
	Total int `json:"total"`
}

type MappingRequest struct {
	UrlPath         string                           `json:"urlPath"`
	Method          string                           `json:"method"`
	QueryParameters map[string]*QueryParameterFilter `json:"queryParameters"`
}

type MappingResponse struct {
	Status   int                     `json:"status"`
	JsonBody *map[string]interface{} `json:"jsonBody"`
	Headers  *map[string]string      `json:"headers"`
}

type QueryParameterFilter struct {
	EqualTo string `json:"equalTo"`
}

func (c *Client) GetStub(name string) (*AdminMapping, error) {
	var restyClient = resty.New()
	var resp *resty.Response
	var err error

	defer func() { _ = restyClient.Close() }()
	resp, err = restyClient.R().
		SetExpectResponseContentType("application/json").
		SetResult(&AdminMap{}).
		Get(c.HostAddr + "/__admin/mappings")

	if err == nil {
		if resp.IsSuccess() {
			var adminMap = resp.Result().(*AdminMap)

			for _, mapping := range adminMap.Mappings {
				if mapping.Name == name {
					return &mapping, nil
				}
			}
		}

		err = errors.New("could not find mapping")
	}

	return nil, err
}

func (c *Client) GetStubsByPath(urlPath string) ([]*AdminMapping, error) {
	var result []*AdminMapping
	var restyClient = resty.New()
	var resp *resty.Response
	var err error

	defer func() { _ = restyClient.Close() }()
	resp, err = restyClient.R().
		SetExpectResponseContentType("application/json").
		SetResult(&AdminMap{}).
		Get(c.HostAddr + "/__admin/mappings")

	if err == nil {
		if resp.IsSuccess() {
			var adminMap = resp.Result().(*AdminMap)

			for _, mapping := range adminMap.Mappings {
				var mappingRequest = mapping.Request

				if mappingRequest != nil &&
					strings.HasPrefix(mappingRequest.UrlPath, urlPath) {
					result = append(result, &mapping)
				}
			}
		} else {
			return nil, errors.New("could not find mapping")
		}
	}

	return result, nil
}

func (c *Client) Reset() error {
	var restyClient = resty.New()
	var resp *resty.Response
	var err error

	defer func() { _ = restyClient.Close() }()
	resp, err = restyClient.R().
		SetExpectResponseContentType("application/json").
		Post(c.HostAddr + "/__admin/mappings/reset")

	if err == nil {
		if !resp.IsSuccess() {
			return fmt.Errorf("could not reset mappings [%s]", resp.RawResponse.Status)
		}
	}

	return nil
}

func (c *Client) UpdateStub(id string, mapping *AdminMapping) error {
	var restyClient = resty.New()
	var resp *resty.Response
	var err error

	defer func() { _ = restyClient.Close() }()
	resp, err = restyClient.R().
		SetAuthScheme("").
		SetExpectResponseContentType("application/json").
		SetContentType("application/json").
		SetBody(mapping).
		Put(c.HostAddr + "/__admin/mappings/" + id)

	if err == nil {
		if !resp.IsSuccess() {
			return fmt.Errorf("could not update mapping [%s]", resp.RawResponse.Status)
		}
	}

	return nil
}

func NewWiremockClient(hostAddr string) *Client {
	var result []string

	if !strings.HasPrefix(hostAddr, "http") {
		result = append(result, "http:/")
	}

	result = append(result, hostAddr)
	return &Client{
		HostAddr: strings.Join(result, "/"),
	}
}
