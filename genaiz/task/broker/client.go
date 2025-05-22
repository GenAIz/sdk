package broker

import (
	"errors"
	"strconv"
	"strings"

	"resty.dev/v3"
)

const (
	defaultExpiryHours = 12

	apiVersion1 version = "v1"
	pathSession path    = "user/session"
)

type path string

type version string

type Client struct {
	BasicAuth string
	Expiry    int
	HostAddr  string

	factory func() *resty.Client
}

func (p Client) Login(username string, password *[]byte) (*Session, error) {
	var url string
	var err error

	if url, err = p.makeUrl(apiVersion1, pathSession, "create"); err == nil {
		var client = p.factory()
		var resp *resty.Response

		defer func(client *resty.Client) {
			_ = client.Close()
		}(client)

		resp, err = client.R().
			SetExpectResponseContentType("application/json").
			SetResult(&Session{}).
			SetFormData(map[string]string{
				"email":    username,
				"password": string(*password),
				"expiry":   strconv.FormatInt(int64(p.Expiry*60), 10),
			}).
			Post(url)

		if err != nil {
			return nil, err
		}

		if resp != nil {
			if resp.IsSuccess() {
				return resp.Result().(*Session), nil
			} else {
				return nil, errors.New(resp.Status())
			}
		}
	}

	return nil, err
}

func (p Client) makeUrl(version version, path path, rpc ...string) (string, error) {
	var result []string

	if p.HostAddr == "" {
		return "", errors.New("invalid host address")
	}

	if !strings.HasPrefix(p.HostAddr, "http") {
		result = append(result, "http:/")
	}

	result = append(result, p.HostAddr, string(version), string(path))

	if len(rpc) > 0 {
		result = append(result, rpc...)
	}

	return strings.Join(result, "/"), nil
}

func NewClient(addr string) *Client {
	return NewClientWithFactory(addr, defaultExpiryHours, func() *resty.Client {
		return resty.New()
	})
}

func NewClientWithExpiry(addr string, expiry int) *Client {
	return NewClientWithFactory(addr, expiry, func() *resty.Client {
		return resty.New()
	})
}

func NewClientWithFactory(addr string, expiry int, factory func() *resty.Client) *Client {
	return &Client{
		Expiry:   expiry,
		HostAddr: addr,
		factory:  factory,
	}
}
