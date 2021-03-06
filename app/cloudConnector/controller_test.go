/* Apache v2 license
*  Copyright (C) <2019> Intel Corporation
*
*  SPDX-License-Identifier: Apache-2.0
 */

package cloudConnector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"strings"

	"github.com/intel/rsp-sw-toolkit-im-suite-cloud-connector-service/app/config"
)

func GenerateWebhook(testServerURL string, auth bool, methodType string) Webhook {
	n := Webhook{
		Method:  methodType,
		URL:     testServerURL + "/callwebhook",
		Payload: []byte{},
	}
	if auth {
		n.Auth = Auth{AuthType: "oauth2", Endpoint: testServerURL + "/oauth", Data: "testname:testpassword"}
	}

	return n
}

func TestMain(m *testing.M) {

	_ = config.InitConfig(nil)

	os.Exit(m.Run())

}

// nolint: dupl
func TestOAuth2PostWebhookOk(t *testing.T) {
	accessTokens = sync.Map{}
	testJdaMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("hit server")
		if request.Method != "POST" {
			t.Errorf("Expected 'POST' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()
		if escapedPath == "/oauth" {
			data := make(map[string]interface{})
			data["access_token"] = "eyJhbGci0iJSUzI1NiJ9.eyJeHAi0jE0NjUzMzU.eju3894"
			data["token_type"] = "bearer"
			data["expires_in"] = 3599
			data["scope"] = "access"
			data["jti"] = "aceable12-1709-4aae-a289-df8b88c84c95"
			data["x-tenant-id"] = "c54d9ccb-6ddd-4416-85d4-e01f565b1266"

			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else if escapedPath == "/callwebhook" {
			data := make(map[string]interface{})
			data["timestamp"] = 14903136768
			data["skus"] = `["MS122-38"]`
			data["ruleId"] = "SomeRuleId-1234"
			data["notificationId"] = "Out-of-stock-ShoesId123"
			data["stockCount"] = 0.0
			data["sellCount"] = 0.0

			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testJdaMockServer.Close()

	webHook := GenerateWebhook(testJdaMockServer.URL, true, http.MethodPost)
	data := []byte(`{ }`)

	webHook.URL = testJdaMockServer.URL + "/callwebhook"
	webHook.Auth.AuthType = "OAuth2"
	webHook.Auth.Endpoint = testJdaMockServer.URL + "/oauth"
	webHook.Auth.Data = "this is a test"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "")
	if err != nil {
		t.Error(err)
	}

	_, secondErr := ProcessWebhook(webHook, "")
	if secondErr != nil {
		t.Error(secondErr)
	}
}

// nolint: dupl
func TestOAuth2GetWebhookOk(t *testing.T) {
	accessTokens = sync.Map{}
	testJdaMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		escapedPath := request.URL.EscapedPath()
		if escapedPath == "/oauth" {
			if request.Method != "POST" {
				t.Errorf("Expected 'POST' request, received '%s", request.Method)
			}
			data := make(map[string]interface{})
			data["access_token"] = "eyJhbGci0iJSUzI1NiJ9.eyJeHAi0jE0NjUzMzU.eju3894"
			data["token_type"] = "bearer"
			data["expires_in"] = 3599
			data["scope"] = "access"
			data["jti"] = "aceable12-1709-4aae-a289-df8b88c84c95"
			data["x-tenant-id"] = "c54d9ccb-6ddd-4416-85d4-e01f565b1266"

			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else if escapedPath == "/callwebhook" {
			if request.Method != "GET" {
				t.Errorf("Expected 'GET' request, received '%s", request.Method)
			}
			data := make(map[string]interface{})
			data["timestamp"] = 14903136768
			data["skus"] = `["MS122-38"]`
			data["ruleId"] = "SomeRuleId-1234"
			data["notificationId"] = "Out-of-stock-ShoesId123"
			data["stockCount"] = 0.0
			data["sellCount"] = 0.0

			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testJdaMockServer.Close()

	webHook := GenerateWebhook(testJdaMockServer.URL, true, http.MethodGet)
	webHook.URL = testJdaMockServer.URL + "/callwebhook"
	webHook.Auth.AuthType = "oauth2"
	webHook.Auth.Endpoint = testJdaMockServer.URL + "/oauth"
	webHook.Auth.Data = "this is a test"

	_, err := ProcessWebhook(webHook, "")
	if err != nil {
		t.Error(err)
	}
}

func TestOAuth2PostWebhookForbidden(t *testing.T) {
	accessTokens = sync.Map{}
	testJdaMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Errorf("Expected 'POST' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()
		if escapedPath == "/oauth" {
			// authentication failed
			writer.WriteHeader(http.StatusUnauthorized)
		} else if escapedPath == "/callwebhook" {
			writer.WriteHeader(http.StatusForbidden)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testJdaMockServer.Close()

	webHook := GenerateWebhook(testJdaMockServer.URL, true, http.MethodPost)
	data := []byte(`{ }`)

	webHook.URL = testJdaMockServer.URL + "/callwebhook"
	webHook.Auth.AuthType = "OAuth2"
	webHook.Auth.Endpoint = testJdaMockServer.URL + "/oauth"
	webHook.Auth.Data = "this is a test"
	webHook.Payload = data

	// expecting unauthorized
	_, err := ProcessWebhook(webHook, "")
	if err == nil {
		t.Error("Expected authentication error, not 200 status")
	}
}

// nolint: dupl
func TestOAuth2PostWebhookFailNotification(t *testing.T) {
	accessTokens = sync.Map{}
	testJdaMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Errorf("Expected 'POST' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()
		if escapedPath == "/oauth" {
			data := make(map[string]interface{})
			data["access_token"] = "eyJhbGci0iJSUzI1NiJ9.eyJeHAi0jE0NjUzMzU.eju3894"
			data["token_type"] = "bearer"
			data["expires_in"] = 3599
			data["scope"] = "access"
			data["jti"] = "aceable12-1709-4aae-a289-df8b88c84c95"
			data["x-tenant-id"] = "c54d9ccb-6ddd-4416-85d4-e01f565b1266"

			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else if escapedPath == "/callwebhook" {
			writer.WriteHeader(http.StatusForbidden)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testJdaMockServer.Close()

	webHook := GenerateWebhook(testJdaMockServer.URL, true, http.MethodPost)
	data := []byte(`{ }`)

	webHook.URL = testJdaMockServer.URL + "/callwebhook"
	webHook.Auth.AuthType = "OAuth2"
	webHook.Auth.Endpoint = testJdaMockServer.URL + "/oauth"
	webHook.Auth.Data = "this is a test"
	webHook.Payload = data

	// expecting unauthorized
	_, err := ProcessWebhook(webHook, "")
	if err == nil {
		t.Error("Expected POST notification error, not 200 status")
	}
}

// nolint: dupl
func TestOAuth2GetWebhookFailNotification(t *testing.T) {
	accessTokens = sync.Map{}
	testJdaMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		escapedPath := request.URL.EscapedPath()
		if escapedPath == "/oauth" {
			if request.Method != "POST" {
				t.Errorf("Expected 'POST' request, received '%s", request.Method)
			}
			data := make(map[string]interface{})
			data["access_token"] = "eyJhbGci0iJSUzI1NiJ9.eyJeHAi0jE0NjUzMzU.eju3894"
			data["token_type"] = "bearer"
			data["expires_in"] = 3599
			data["scope"] = "access"
			data["jti"] = "aceable12-1709-4aae-a289-df8b88c84c95"
			data["x-tenant-id"] = "c54d9ccb-6ddd-4416-85d4-e01f565b1266"

			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else if escapedPath == "/callwebhook" {
			writer.WriteHeader(http.StatusForbidden)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testJdaMockServer.Close()

	webHook := GenerateWebhook(testJdaMockServer.URL, true, http.MethodGet)
	webHook.URL = testJdaMockServer.URL + "/callwebhook"
	webHook.Auth.AuthType = "oauth2"
	webHook.Auth.Endpoint = testJdaMockServer.URL + "/oauth"
	webHook.Auth.Data = "this is a test"

	_, err := ProcessWebhook(webHook, "")
	if err == nil {
		t.Error("Expected POST notification error, not 200 status")
	}
}

// nolint: dupl
func TestPostWebhookNoAuthenticationOK(t *testing.T) {
	accessTokens = sync.Map{}
	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Errorf("Expected 'POST' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()

		expectedHeaderItem := "application/x-www-form-urlencoded"

		if request.Header["Content-Type"][0] != expectedHeaderItem {
			t.Errorf("Expected request header content to be %s, received %s", expectedHeaderItem, request.Header["Content-Type"][0])
		}

		if escapedPath == "/callwebhook" {
			data := make(map[string]interface{})
			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testMockServer.Close()

	webHook := GenerateWebhook(testMockServer.URL, false, http.MethodPost)
	data := []byte(`{ }`)

	webHook.Method = "POST"
	webHook.Header = http.Header{}
	webHook.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	webHook.URL = testMockServer.URL + "/callwebhook"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "")
	if err != nil {
		t.Error(err)
	}
}

func TestGetWebhookNoAuthenticationOK(t *testing.T) {
	accessTokens = sync.Map{}
	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			t.Errorf("Expected 'GET' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()

		expectedHeaderItem := "application/x-www-form-urlencoded"

		if request.Header["Content-Type"][0] != expectedHeaderItem {
			t.Errorf("Expected request header content to be %s, received %s", expectedHeaderItem, request.Header["Content-Type"][0])
		}

		if escapedPath == "/callwebhook" {
			data := make(map[string]interface{})
			jsonData, _ := json.Marshal(data)
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write(jsonData)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testMockServer.Close()

	webHook := GenerateWebhook(testMockServer.URL, false, http.MethodGet)
	data := []byte(`{ }`)

	webHook.Method = "GET"
	webHook.Header = http.Header{}
	webHook.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	webHook.URL = testMockServer.URL + "/callwebhook"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "")
	if err != nil {
		t.Error(err)
	}
}

func TestPostWebhookNoAuthenticationForbidden(t *testing.T) {
	accessTokens = sync.Map{}
	testJdaMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			t.Errorf("Expected 'POST' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()
		if escapedPath == "/callwebhook" {
			writer.WriteHeader(http.StatusForbidden)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testJdaMockServer.Close()

	webHook := GenerateWebhook(testJdaMockServer.URL, false, http.MethodPost)
	data := []byte(`{ }`)
	webHook.Method = "POST"
	webHook.URL = testJdaMockServer.URL + "/callwebhook"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "")
	if err == nil {
		t.Fatalf("Expected error, but didn't get one")
	}

	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("Received error as expected, but didn't get one with 403 status.")
	}
}

func TestGetWebhookNoAuthenticationForbidden(t *testing.T) {
	accessTokens = sync.Map{}
	testMockServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != "GET" {
			t.Errorf("Expected 'GET' request, received '%s", request.Method)
		}

		escapedPath := request.URL.EscapedPath()

		expectedHeaderItem := "application/x-www-form-urlencoded"

		if request.Header["Content-Type"][0] != expectedHeaderItem {
			t.Errorf("Expected request header content to be %s, received %s", expectedHeaderItem, request.Header["Content-Type"][0])
		}

		if escapedPath == "/callwebhook" {
			writer.WriteHeader(http.StatusForbidden)
		} else {
			t.Errorf("Expected request to '/oauth' or 'notification', received %s", escapedPath)
		}
	}))

	defer testMockServer.Close()

	webHook := GenerateWebhook(testMockServer.URL, false, http.MethodGet)
	data := []byte(`{ }`)

	webHook.Method = "GET"
	webHook.Header = http.Header{}
	webHook.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}
	webHook.URL = testMockServer.URL + "/callwebhook"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "")
	if err == nil {
		t.Fatalf("Expected error, but didn't get one")
	}

	if !strings.Contains(err.Error(), "403") {
		t.Fatalf("Received error as expected, but didn't get one with 403 status.")
	}

}

func TestPostWebhookProxy(t *testing.T) {
	accessTokens = sync.Map{}
	testURL := "testURL.com"
	webHook := GenerateWebhook(testURL, false, http.MethodPost)
	data := []byte(`{ }`)

	webHook.URL = testURL + "/callwebhook"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "test.proxy")
	if err == nil {
		t.Error(err)
	}
}

func TestGetWebhookProxy(t *testing.T) {
	accessTokens = sync.Map{}
	testURL := "testURL.com"
	webHook := GenerateWebhook(testURL, false, http.MethodGet)
	data := []byte(`{ }`)

	webHook.URL = testURL + "/callwebhook"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "test.proxy")
	if err == nil {
		t.Error(err)
	}
}

func TestPostOAuth2WebhookProxy(t *testing.T) {
	accessTokens = sync.Map{}
	testURL := "testURL.com"
	webHook := GenerateWebhook(testURL, false, http.MethodPost)
	data := []byte(`{ }`)

	webHook.URL = testURL + "/callwebhook"
	webHook.Auth.AuthType = "OAuth2"
	webHook.Auth.Endpoint = testURL + "/oauth"
	webHook.Auth.Data = "this is a test"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "test.proxy")
	if err == nil {
		t.Error(err)
	}
}

func TestGetOAuth2WebhookProxy(t *testing.T) {
	accessTokens = sync.Map{}
	testURL := "testURL.com"
	webHook := GenerateWebhook(testURL, false, http.MethodPost)
	data := []byte(`{ }`)

	webHook.URL = testURL + "/callwebhook"
	webHook.Auth.AuthType = "OAuth2"
	webHook.Auth.Endpoint = testURL + "/oauth"
	webHook.Auth.Data = "this is a test"
	webHook.Payload = data

	_, err := ProcessWebhook(webHook, "test.proxy")
	if err == nil {
		t.Error(err)
	}
}
