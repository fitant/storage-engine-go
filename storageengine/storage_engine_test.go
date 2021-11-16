package storageengine_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fitant/storage-engine-go/storageengine"

	"github.com/stretchr/testify/assert"
)

type readRequest struct {
	ID   string `json:"id" bson:"id"`
	Pass string `json:"pass" bson:"id"`
}

type readResponse struct {
	ID   string `json:"id" bson:"id"`
	Note string `json:"note" bson:"note"`
}

type mockHttpClient struct {
	testType string
	endpoint string
	reqComp  interface{}
	resComp  interface{}
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
	case "upstream":
		if req.URL.String() == m.endpoint {
			return &http.Response{
				StatusCode: http.StatusOK,
			}, nil
		}
		if req.URL.String() == m.endpoint+"/read" && req.Method == http.MethodGet {
			b, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			var rr readRequest
			err = json.Unmarshal(b, &rr)
			if err != nil {
				return nil, err
			}
			d, _ := json.Marshal(m.resComp.(readResponse))
			if rr == m.reqComp.(readRequest) {
				b := ioutil.NopCloser(bytes.NewReader(d))
				x := &http.Response{
					StatusCode: http.StatusOK,
					Body:       b,
				}
				return x, nil
			}
		}
	}
	return nil, nil
}

func newMockHTTPClient(endpoint string) mockHttpClient {
	return mockHttpClient{endpoint: endpoint, testType: "local"}
}

func newUpstreamMockHTTPClient(endpoint string, reqComp interface{}, resComp interface{}) mockHttpClient {
	return mockHttpClient{endpoint: endpoint, testType: "upstream", reqComp: reqComp, resComp: resComp}
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
	mockReq := readRequest{
		ID:   "wow",
		Pass: "hello",
	}
	mockRes := readResponse{
		ID:   "wow",
		Note: "hello i am awesome",
	}
	// Successfully fetch data from storage engine
	valid_cc1, _ := storageengine.NewClientConfig(newUpstreamMockHTTPClient(endpoint, mockReq, mockRes), endpoint)
	obj_1, _ := storageengine.NewObject(valid_cc1)
	obj_1.SetID("wow")
	obj_1.SetPassword("hello")
	err_1 := obj_1.Refresh()
	assert.Nil(t, err_1)
	assert.Equal(t, mockRes.Note, obj_1.GetData())
}
