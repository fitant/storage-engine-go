package storageengine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var ErrObjectNotFound = errors.New("storage engine object not found")

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientConfig struct {
	httpClient HttpClient
	endpoint   string
}

type Object struct {
	cc         *ClientConfig
	id         string
	data       string
	currPass   string
	isUpstream bool
}

func NewClientConfigDefaultClient(endpoint string) (*ClientConfig, error) {
	return NewClientConfig(http.DefaultClient, endpoint)
}

func NewClientConfig(httpClient HttpClient, endpoint string) (*ClientConfig, error) {
	if httpClient == nil {
		return nil, errors.New("http client cannot be nil")
	}
	if endpoint == "" {
		return nil, errors.New("endpoint cannot be empty")
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if endpointURL.Scheme == "" || endpointURL.Host == "" {
		return nil, errors.New("endpoint must be an absolute URL")
	}
	// Do a GET on base endpoint expecting a StatusOK
	req := http.Request{
		URL:    endpointURL,
		Method: http.MethodGet,
	}
	resp, err := httpClient.Do(&req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("response is nil")
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code not ok: %d", resp.StatusCode)
	}
	return &ClientConfig{
		httpClient: httpClient,
		endpoint:   endpoint,
	}, nil
}

func NewObject(clientConfig *ClientConfig) (*Object, error) {
	if clientConfig == nil {
		return nil, errors.New("clientConfig cannot be nil")
	}
	return &Object{cc: clientConfig}, nil
}

func (o *Object) SetID(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}

	if o.id != id {
		o.isUpstream = false
		o.data = ""
	}

	o.id = id
	return nil
}

func (o *Object) GetID() string {
	return o.id
}

func (o *Object) SetData(data string) error {
	if data == "" {
		return errors.New("data cannot be empty")
	}
	o.data = data
	return nil
}

func (o *Object) GetData() string {
	return o.data
}

func (o *Object) SetPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	o.currPass = password
	return nil
}

func (o *Object) GetPassword() string {
	return o.currPass
}

func (o *Object) Refresh() error {
	// Make sure there are no zero-value calls
	if o.id == "" || o.currPass == "" {
		return errors.New("id or pass are empty")
	}
	reqData := ReadRequest{
		ID:   o.id,
		Pass: o.currPass,
	}
	data, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	endpointURL, err := url.Parse(o.cc.endpoint + "/read")
	if err != nil {
		return err
	}
	req := &http.Request{
		URL:    endpointURL,
		Method: http.MethodGet,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(data)),
	}
	res, err := o.cc.httpClient.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("response is nil")
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	// Check status OK before reading data
	// Storage Engine only sends data on OK
	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return ErrObjectNotFound
		}
		return fmt.Errorf("response status not ok: %d", res.StatusCode)
	}
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var response ReadResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}
	o.data = response.Note
	o.isUpstream = true
	return nil
}

func (o *Object) Publish() error {
	// Make sure there are no zero-value calls
	// Storage Engine supports empty ID
	if o.currPass == "" || o.data == "" {
		return errors.New("data or pass are empty")
	}
	reqData := PublishRequest{
		ID:   o.id,
		Pass: o.currPass,
		Note: o.data,
	}
	data, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	var endpointURL *url.URL
	var method string
	if !o.isUpstream {
		endpointURL, err = url.Parse(o.cc.endpoint + "/create")
		method = http.MethodPost
	} else {
		endpointURL, err = url.Parse(o.cc.endpoint + "/update/note")
		method = http.MethodPut
	}
	if err != nil {
		return err
	}
	req := &http.Request{
		URL:    endpointURL,
		Method: method,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewReader(data)),
	}
	res, err := o.cc.httpClient.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("response is nil")
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	// Check status OK before reading data
	// Storage Engine only sends data on OK
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("response status not ok: %d", res.StatusCode)
	}
	data, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var response PublishResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}
	o.id = response.ID
	o.isUpstream = true
	return nil
}
