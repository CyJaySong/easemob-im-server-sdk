package core

import (
	"errors"
	"github.com/imroc/req/v3"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type Options struct {
	Host                string
	AppKey              string
	ClientID            string
	ClientSecret        string
	TTL                 int64
	unauthorizedHandler func(c *client) error
}

type Client interface {
	// BaseUrl 获取基础url
	BaseUrl() string
	// Get GET请求
	Get(uri string, data interface{}, dataContentType interface{}, resp interface{}) error
	// Post POST请求
	Post(uri string, data interface{}, dataContentType interface{}, resp interface{}) error
	// Put PUT请求
	Put(uri string, data interface{}, dataContentType interface{}, resp interface{}) error
	// Patch PATCH请求
	Patch(uri string, data interface{}, dataContentType interface{}, resp interface{}) error
	// Delete DELETE请求
	Delete(uri string, data interface{}, dataContentType interface{}, resp interface{}) error
}

type client struct {
	opts    *Options
	client  *req.Client
	baseUrl string
}

func NewClient(opts *Options) *client {
	args := strings.Split(opts.AppKey, "#")
	if len(args) != 2 {
		log.Fatal("invalid appKey")
	}

	c := new(client)
	c.opts = opts
	c.baseUrl = "https://" + opts.Host + "/" + args[0] + "/" + args[1]
	c.client = req.C().SetBaseURL(c.baseUrl).
		SetCommonContentType("application/json; charset=utf-8").
		SetCommonHeader("Accept", "application/json; charset=utf-8")
	return c
}

// BaseUrl 获取基础url
func (c *client) BaseUrl() string {
	return c.baseUrl
}

// Get GET请求
func (c *client) Get(uri string, data interface{}, dataContentType interface{}, resp interface{}) error {
	return c.request(http.MethodGet, uri, data, dataContentType, resp)
}

// Post POST请求
func (c *client) Post(uri string, data interface{}, dataContentType interface{}, resp interface{}) error {
	return c.request(http.MethodPost, uri, data, dataContentType, resp)
}

// Put PUT请求
func (c *client) Put(uri string, data interface{}, dataContentType interface{}, resp interface{}) error {
	return c.request(http.MethodPut, uri, data, dataContentType, resp)
}

// Patch PATCH请求
func (c *client) Patch(uri string, data interface{}, dataContentType interface{}, resp interface{}) error {
	return c.request(http.MethodPatch, uri, data, dataContentType, resp)
}

// Delete DELETE请求
func (c *client) Delete(uri string, data interface{}, dataContentType interface{}, resp interface{}) error {
	return c.request(http.MethodDelete, uri, data, dataContentType, resp)
}

// HTTP请求
func (c *client) request(method string, uri string, data interface{}, dataContentType interface{}, resp interface{}) error {
	for i := 0; i < 2; i++ {
		var r = c.client.R()
		if data != nil && dataContentType == "application/x-www-form-urlencoded" {
			r.SetFormDataFromValues(data.(url.Values))
		} else if data != nil {
			r.SetBodyJsonMarshal(data)
		}
		switch v := dataContentType.(type) {
		case string:
			if len(v) > 0 {
				r.SetContentType(v)
			}
		}
		res, err := r.Send(method, uri)
		if err != nil {
			return err
		}

		if res.Response.StatusCode == http.StatusOK {
			if resp == nil || reflect.ValueOf(resp).IsNil() {
				_ = res.Body.Close()
				return nil
			}
			return res.UnmarshalJson(resp)
		}

		if res.Response.StatusCode == http.StatusUnauthorized {
			_ = res.Body.Close()
			if c.opts.unauthorizedHandler != nil && i < 1 {
				if err = c.opts.unauthorizedHandler(c); err != nil {
					return err
				}
				continue
			}
		}

		errResp := &errorResp{}
		if err = res.UnmarshalJson(errResp); err != nil {
			return err
		}

		return errors.New(errResp.ErrorDescription)
	}

	return nil
}
