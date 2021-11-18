package storageengine_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fitant/storage-engine-go/storageengine"

	"github.com/stretchr/testify/assert"
)

type httpClientDo func(req *http.Request) (*http.Response, error)

type httpMock struct {
	logic httpClientDo
}

func (h httpMock) Do(req *http.Request) (*http.Response, error) {
	return h.logic(req)
}

type mockHttpClient struct {
	testType string
	endpoint string
}

func (m mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	switch m.testType {
	case "local":
		if req.URL.String() == m.endpoint {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}
		return nil, nil
	}
	return nil, nil
}

func newMockHTTPClient(endpoint string) mockHttpClient {
	return mockHttpClient{endpoint: endpoint, testType: "local"}
}

func TestNewClientConfig(t *testing.T) {
	// Fail to create ClientConfig with empty endpoint
	cc_1, err_1 := storageengine.NewClientConfig(http.DefaultClient, "")
	assert.Nil(t, cc_1)
	assert.NotNil(t, err_1)
	// Successfully create ClientConfig with endpoint
	endpoint_2 := "http://localhost:69"
	cc_2, err_2 := storageengine.NewClientConfig(newMockHTTPClient(endpoint_2), endpoint_2)
	assert.NotNil(t, cc_2)
	assert.Nil(t, err_2)
	// Successfully create ClientConfig with endpoint
	endpoint_3 := "https://storage-engine.example.com"
	cc_3, err_3 := storageengine.NewClientConfig(newMockHTTPClient(endpoint_3), endpoint_3)
	assert.NotNil(t, cc_3)
	assert.Nil(t, err_3)
}

func TestNewObject(t *testing.T) {
	endpoint := "https://storage-engine.example.com"
	valid_cc, _ := storageengine.NewClientConfig(newMockHTTPClient(endpoint), endpoint)
	// Create Object with valid config
	obj_1, err_1 := storageengine.NewObject(valid_cc)
	assert.NotNil(t, obj_1)
	assert.Nil(t, err_1)
	// Fail to create Object with invalid (nil) config
	obj_2, err_2 := storageengine.NewObject(nil)
	assert.Nil(t, obj_2)
	assert.NotNil(t, err_2)
}

// NOTE: A just created object has empty zero-value fields this is checked before upstream publish

func TestObjectSetID(t *testing.T) {
	endpoint := "https://storage-engine.example.com"
	valid_cc, _ := storageengine.NewClientConfig(newMockHTTPClient(endpoint), endpoint)
	// Set ID on new object successfully
	obj_1, _ := storageengine.NewObject(valid_cc)
	err_1 := obj_1.SetID("bkrst")
	assert.Nil(t, err_1)
	assert.Equal(t, obj_1.GetID(), "bkrst")
	// Fail to set empty ID on object
	obj_2, _ := storageengine.NewObject(valid_cc)
	err_2 := obj_2.SetID("")
	assert.NotNil(t, err_2)
	// Set special symbol ID on new object successfully
	// StorageEngine accepts special sybols in ID
	obj_3, _ := storageengine.NewObject(valid_cc)
	err_3 := obj_3.SetID("&*##@@!#")
	assert.Nil(t, err_3)
	assert.Equal(t, obj_3.GetID(), "&*##@@!#")
}

func TestObjectSetData(t *testing.T) {
	endpoint := "https://storage-engine.example.com"
	valid_cc, _ := storageengine.NewClientConfig(newMockHTTPClient(endpoint), endpoint)
	// Set data on new object successfully
	obj_1, _ := storageengine.NewObject(valid_cc)
	err_1 := obj_1.SetData("sometimes you just gotta deal with it")
	assert.Nil(t, err_1)
	assert.Equal(t, obj_1.GetData(), "sometimes you just gotta deal with it")
	// Fail to set empty data on object
	obj_2, _ := storageengine.NewObject(valid_cc)
	err_2 := obj_2.SetData("")
	assert.NotNil(t, err_2)
	// Set special symbol data on new object successfully
	obj_3, _ := storageengine.NewObject(valid_cc)
	err_3 := obj_3.SetData("&*freg##@@!#")
	assert.Nil(t, err_3)
	assert.Equal(t, obj_3.GetData(), "&*freg##@@!#")
}

func TestObjectSetPass(t *testing.T) {
	endpoint := "https://storage-engine.example.com"
	valid_cc, _ := storageengine.NewClientConfig(newMockHTTPClient(endpoint), endpoint)
	// Set pass on new object successfully
	obj_1, _ := storageengine.NewObject(valid_cc)
	err_1 := obj_1.SetPassword("u$eD!cew@re")
	assert.Nil(t, err_1)
	assert.Equal(t, obj_1.GetPassword(), "u$eD!cew@re")
	// Fail to set empty pass on object
	obj_2, _ := storageengine.NewObject(valid_cc)
	err_2 := obj_2.SetPassword("")
	assert.NotNil(t, err_2)
	// Set special symbol pass on new object successfully
	obj_3, _ := storageengine.NewObject(valid_cc)
	err_3 := obj_3.SetPassword("&*##@@tgbr!#")
	assert.Nil(t, err_3)
	assert.Equal(t, obj_3.GetPassword(), "&*##@@tgbr!#")
}

func TestObjectRefresh(t *testing.T) {
	endpoint := "https://storage-engine.example.com"

	// Required Types for test
	type readRequest struct {
		ID   string `json:"id" bson:"id"`
		Pass string `json:"pass" bson:"id"`
	}
	type readResponse struct {
		ID   string `json:"id" bson:"id"`
		Note string `json:"note" bson:"note"`
	}

	// Successfully fetch data from storage engine
	note_1 := "hello i am awesome"
	x_1 := func(req *http.Request) (*http.Response, error) {
		// This makes the valid_cc generate else it be nil
		if req.URL.String() == endpoint {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}
		// Test related code
		req_1 := readRequest{
			ID:   "wow",
			Pass: "hello",
		}
		res_1 := readResponse{
			ID:   "wow",
			Note: note_1,
		}
		req_1_bytes, _ := json.Marshal(req_1)
		res_1_bytes, _ := json.Marshal(res_1)
		if req.URL.String() == endpoint+"/read" && req.Method == http.MethodGet {
			req_bytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(req_bytes, req_1_bytes) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(res_1_bytes)),
				}, nil
			}
		}
		return nil, errors.New("something horrible must have happened here")
	}
	valid_cc1, _ := storageengine.NewClientConfig(httpMock{logic: x_1}, endpoint)
	obj_1, _ := storageengine.NewObject(valid_cc1)
	obj_1.SetID("wow")
	obj_1.SetPassword("hello")
	err_1 := obj_1.Refresh()
	assert.Nil(t, err_1)
	assert.Equal(t, note_1, obj_1.GetData())

	// Don't crash even when response error and res are nil
	x_2 := func(req *http.Request) (*http.Response, error) {
		// This makes the valid_cc generate else it be nil
		if req.URL.String() == endpoint {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}
		// Test related code
		return nil, nil
	}
	valid_cc2, _ := storageengine.NewClientConfig(httpMock{logic: x_2}, endpoint)
	obj_2, _ := storageengine.NewObject(valid_cc2)
	obj_2.SetID("y24786*(5")
	obj_2.SetPassword("14/-/*-+")
	err_2 := obj_2.Refresh()
	assert.NotNil(t, err_2)

	// Handle bad request response
	x_3 := func(req *http.Request) (*http.Response, error) {
		// This makes the valid_cc generate else it be nil
		if req.URL.String() == endpoint {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}
		// Test related code
		return &http.Response{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	valid_cc3, _ := storageengine.NewClientConfig(httpMock{logic: x_3}, endpoint)
	obj_3, _ := storageengine.NewObject(valid_cc3)
	obj_3.SetID("234y_+l:w|><>/")
	obj_3.SetPassword("687/-/*32fb&**")
	err_3 := obj_3.Refresh()
	assert.NotNil(t, err_3)
}

func TestObjectPublish(t *testing.T) {
	endpoint := "https://storage-engine.example.com"

	type createRequest struct {
		ID   string `json:"id" bson:"id"`
		Pass string `json:"pass" bson:"pass"`
		Note string `json:"note" bson:"note"`
	}
	type createResponse struct {
		ID string `json:"id" bson:"id"`
	}

	// Successfully create a new note
	var req_checked_out_1 bool
	var req_checked_out_1_2 bool
	x_1 := func(req *http.Request) (*http.Response, error) {
		// This makes the valid_cc generate else it be nil
		if req.URL.String() == endpoint {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}

		req_1 := createRequest{
			ID:   "wow",
			Pass: "02_+",
			Note: "heyo",
		}
		// Simulate StorageEngine sending back a different ID
		res_1 := createResponse{
			ID: "wo2",
		}
		req_1_bytes, _ := json.Marshal(req_1)
		res_1_bytes, _ := json.Marshal(res_1)
		// Test related code
		if req.URL.String() == endpoint+"/create" && req.Method == http.MethodPost {
			req_bytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(req_bytes, req_1_bytes) {
				req_checked_out_1 = true
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(res_1_bytes)),
				}, nil
			}
		}

		req_2 := createRequest{
			ID:   "wo2",
			Pass: "02_+",
			Note: "what's that function",
		}
		req_2_bytes, _ := json.Marshal(req_2)
		// Test related code
		if req.URL.String() == endpoint+"/update/note" && req.Method == http.MethodPut {
			req_bytes, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(req_bytes, req_2_bytes) {
				req_checked_out_1_2 = true
				// Reuse response from previous call
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(res_1_bytes)),
				}, nil
			}
		}
		return nil, errors.New("something horrible must have happened here")
	}
	valid_cc1, _ := storageengine.NewClientConfig(httpMock{logic: x_1}, endpoint)
	obj_1, _ := storageengine.NewObject(valid_cc1)
	obj_1.SetID("wow")
	obj_1.SetPassword("02_+")
	obj_1.SetData("heyo")
	err_1 := obj_1.Publish()
	assert.Nil(t, err_1)
	assert.Equal(t, "wo2", obj_1.GetID())
	assert.Equal(t, true, req_checked_out_1)

	// Try to update the just created note
	obj_1.SetData("what's that function")
	err_1_2 := obj_1.Publish()
	assert.Nil(t, err_1_2)
	assert.Equal(t, "what's that function", obj_1.GetData())
	assert.Equal(t, true, req_checked_out_1_2)
}
