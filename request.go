package go_nordigen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Authorization adds the authorization to the http.Request.
// It should change the http.Request to include whatever needs to be added.
type Authorization interface {
	Header(req *http.Request) (*http.Request, error)
}

// Auditor audits the
type Auditor interface {
	Request(req http.Request)
	Response(resp http.Response)
}

type Request struct {
	client      *http.Client
	baseURL     *url.URL
	auth        Authorization
	headers     map[string]string
	queryParams url.Values
	body        interface{}
	bodyBytes   []byte
	result      interface{}
	auditor     Auditor
}

func NewRequest(
	baseURL *url.URL,
	client *http.Client,
	auth Authorization,
	auditor Auditor,
) *Request {
	return &Request{
		baseURL:     baseURL,
		client:      client,
		auth:        auth,
		auditor:     auditor,
		headers:     make(map[string]string),
		queryParams: make(url.Values),
	}
}

func (r *Request) WithHeader(key, value string) *Request {
	if value != "" {
		r.headers[key] = value
	}
	return r
}

func (r *Request) WithHeaders(headers url.Values) *Request {
	for k, v := range headers {
		r.headers[k] = v[0]
	}
	return r
}

func (r *Request) WithResult(result interface{}) *Request {
	r.result = result
	return r
}

func (r *Request) WithJSONBody(obj interface{}) *Request {
	r.body = obj
	r.headers["Content-Type"] = "application/json"
	return r
}

// WithQParam adds a query parameter to the request.
func (r *Request) WithQParam(key string, value interface{}) *Request {
	switch v := value.(type) {
	case []string:
		if v != nil {
			r.queryParams.Add(key, strings.Join(v, ","))
		}
	case *time.Time:
		if v != nil {
			r.queryParams.Add(key, v.Format("2006-01-02"))
		}
	case int:
		r.queryParams.Add(key, fmt.Sprintf("%d", v))
	}
	return r
}

func (r *Request) Get(ctx context.Context, path string) (interface{}, error) {
	return r.do(ctx, http.MethodGet, path)
}

func (r *Request) Post(ctx context.Context, path string) (interface{}, error) {
	if r.body != nil {
		var err error
		r.bodyBytes, err = json.Marshal(r.body)
		if err != nil {
			return nil, err
		}
	}

	return r.do(ctx, http.MethodPost, path)
}

func (r *Request) Delete(ctx context.Context, path string) (interface{}, error) {
	return r.do(ctx, http.MethodDelete, path)
}

func (r *Request) do(ctx context.Context, method, path string) (interface{}, error) {
	ref, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	refURL := r.baseURL.ResolveReference(ref)

	q, err := url.QueryUnescape(r.queryParams.Encode())
	if err != nil {
		return nil, err
	}
	refURL.RawQuery = q

	req, err := http.NewRequestWithContext(ctx, method, refURL.String(), bytes.NewReader(r.bodyBytes))
	if err != nil {
		return nil, err
	}

	for k, v := range r.headers {
		req.Header.Set(k, v)
	}
	if r.auth != nil {
		req, err = r.auth.Header(req)
		if err != nil {
			return nil, err
		}
	}

	if len(r.bodyBytes) > 0 {
		req.Header.Add("Content-Length", strconv.Itoa(len(r.bodyBytes)))
	}

	if r.auditor != nil {
		r.auditor.Request(*req)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	if r.auditor != nil {
		r.auditor.Response(*resp)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)

	if resp.StatusCode >= 300 {
		re := &RequestError{
			StatusCode: resp.StatusCode,
		}
		dec.Decode(&re)
		return nil, re
	}

	return r.decodeResult(dec)
}

func (r *Request) decodeResult(dec *json.Decoder) (interface{}, error) {
	var err error

	switch reflect.ValueOf(r.result).Kind() {
	case reflect.Ptr:
		err = dec.Decode(r.result)
	case reflect.Struct:
		t := reflect.TypeOf(r.result)
		v := reflect.New(t)
		err = dec.Decode(v.Interface())
		r.result = v.Elem().Interface()
	default:
		var v map[string]interface{}
		err = dec.Decode(&v)
		r.result = v
	}

	return r.result, err
}