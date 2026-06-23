package broker

import (
	"io"
	"net/http"
	"time"

	"resty.dev/v3"

	"genaiz.com/genaiz/version/env"
)

type requestBridge interface {
	io.Closer

	Cookie(*http.Cookie) requestBridge

	Get(string) (responseBridge, error)

	Form() requestBridge

	FormData(map[string]string) requestBridge

	Json() requestBridge

	QueryParams(map[string]string) requestBridge

	Resulting(any) requestBridge

	Post(string) (responseBridge, error)

	Timeout() time.Duration
}

type restyBridge struct {
	client  *resty.Client
	request *resty.Request
}

func (r *restyBridge) Close() error {
	return r.client.Close()
}

func (r *restyBridge) Cookie(cookie *http.Cookie) requestBridge {
	r.request.SetCookie(cookie)
	return r
}

func (r *restyBridge) Get(url string) (responseBridge, error) {
	var resp, err = r.request.Get(url)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *restyBridge) Form() requestBridge {
	r.request.SetResponseExpectContentType("application/x-www-form-urlencoded")
	return r
}

func (r *restyBridge) FormData(params map[string]string) requestBridge {
	r.request.SetFormData(params)
	return r
}

func (r *restyBridge) Json() requestBridge {
	r.request.SetResponseExpectContentType("application/json")
	return r
}

func (r *restyBridge) Post(url string) (responseBridge, error) {
	var resp, err = r.request.Post(url)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r *restyBridge) QueryParams(args map[string]string) requestBridge {
	r.request.SetQueryParams(args)
	return r
}

func (r *restyBridge) Resulting(ref any) requestBridge {
	r.request.SetResult(ref)
	return r
}

func (r *restyBridge) Timeout() time.Duration {
	return r.client.Timeout()
}

type responseBridge interface {
	Bytes() []byte

	Cookies() []*http.Cookie

	IsStatusSuccess() bool

	Result() any

	Status() string

	StatusCode() int
}

func protocolGateChecker(client *resty.Client, req *resty.Request) error {
	_ = client

	if env.IsAllowedProtocol(req.URL) {
		return nil
	}

	return errorDisallowedProtocol
}
