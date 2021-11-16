package storageengine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type HttpClient interface {
	Do (req *http.Request) (*http.Response, error)
}

type clientConfig struct{
	httpClient HttpClient
	endpoint string
}

type object struct{
	cc *clientConfig
	id string
	data string
	currPass string
}

func NewClientConfig(httpClient HttpClient, endpoint string) (*clientConfig, error) {
	if endpoint == "" {
		return nil, errors.New("endpoint cannot be empty")
	}
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	// Do a GET on base endpoint expecting a StatusOK
	req := http.Request{
		URL: endpointURL,
		Method: http.MethodGet,
	}
	resp, err := httpClient.Do(&req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("response is nil")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response status code not ok: %d", resp.StatusCode)
	}
	return &clientConfig{
		httpClient: httpClient,
		endpoint: endpoint,
	}, nil
}

func NewObject(clientConfig *clientConfig) (*object, error) {
	if clientConfig == nil {
		return nil, errors.New("clientconfig is not nillable")
	}
	return &object{cc: clientConfig}, nil
}

func (o *object) SetID(id string) error {
	if id == "" {
		return errors.New("id cannot be empty")
	}
	o.id = id
	return nil
}

func (o *object) GetID() string {
	return o.id
}

func (o *object) SetData(data string) error {
	if data == "" {
		return errors.New("data cannot be empty")
	}
	o.data = data
	return nil
}

func (o *object) GetData() string {
	return o.data
}

func (o *object) SetPassword(password string) error {
	if password == "" {
		return errors.New("password cannot be empty")
	}
	o.currPass = password
	return nil
}


func (o *object) GetPassword() string {
	return o.currPass
}

func (o *object) Refresh() error {
	if o.id == "" || o.currPass == "" {
		return errors.New("id or pass is empty")
	}
	reqData := ReadRequest{
		ID: o.id,
		Pass: o.currPass,
	}
	data, err := json.Marshal(reqData)
	if err != nil {
		return err
	}
	endpointURL, err := url.Parse(o.cc.endpoint+"/read")
	if err != nil {
		return err
	}
	req := &http.Request{
		URL: endpointURL,
		Method: http.MethodGet,
		Body: io.NopCloser(bytes.NewReader(data)),
	}
	res, err := o.cc.httpClient.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("response is nil")
	}
	// Check status OK before reading data
	// Storage Engine only sends data on OK
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("response status not ok: %d", res.StatusCode)
	}
	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var response ReadResponse
	err = json.Unmarshal(data, &response)
	if err != nil {
		return err
	}
	o.data = response.Note	
	return nil
}
