// Copyright (c) 2015-2021 Jeevanandam M (jeeva@myjeeva.com), All rights reserved.
// resty source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package resty

import (
	"bytes"
	"crypto/tls"
	"encoding/xml"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bingoohuang/gg/pkg/iox"
	"github.com/stretchr/testify/assert"
)

type AuthSuccess struct {
	ID, Message string
}

type AuthError struct {
	ID, Message string
}

func TestGet(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dc().R().
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10)).
		Get(ts.URL + "/")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "HTTP/1.1", resp.Proto())
	assert.Equal(t, "200 OK", resp.Status())
	assert.NotNil(t, resp.Body())
	assert.Equal(t, "TestGet: text response", resp.String())

	logResponse(t, resp)
}

func TestIllegalRetryCount(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dc().SetRetryCount(-1).R().Get(ts.URL + "/")

	assert.Nil(t, err)
	assert.Nil(t, resp)
}

func TestGetCustomUserAgent(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dcr().
		SetHeader(hdrUserAgentKey, "Test Custom User agent").
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10)).
		Get(ts.URL + "/")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "HTTP/1.1", resp.Proto())
	assert.Equal(t, "200 OK", resp.Status())
	assert.Equal(t, "TestGet: text response", resp.String())

	logResponse(t, resp)
}

func TestGetClientParamRequestParam(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := dc()
	c.SetQueryParam("client_param", "true").
		SetQueryParams(map[string]string{"req_1": "jeeva", "req_3": "jeeva3"}).
		SetDebug(true)
	c.outputLogTo(io.Discard)

	resp, err := c.R().
		SetQueryParams(map[string]string{"req_1": "req 1 value", "req_2": "req 2 value"}).
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10)).
		SetHeader(hdrUserAgentKey, "Test Custom User agent").
		Get(ts.URL + "/")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "HTTP/1.1", resp.Proto())
	assert.Equal(t, "200 OK", resp.Status())
	assert.Equal(t, "TestGet: text response", resp.String())

	logResponse(t, resp)
}

func TestGetRelativePath(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := dc()
	c.SetBaseURL(ts.URL)

	resp, err := c.R().Get("mypage2")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "TestGet: text response from mypage2", resp.String())

	logResponse(t, resp)
}

func TestGet400Error(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dcr().Get(ts.URL + "/mypage")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
	assert.Equal(t, "", resp.String())

	logResponse(t, resp)
}

func TestPostJSONStringSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetHeader(hdrContentTypeKey, "application/json; charset=utf-8").
		SetHeaders(map[string]string{hdrUserAgentKey: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) go-resty v0.1", hdrAcceptKey: "application/json; charset=utf-8"})

	resp, err := c.R().
		SetBody(`{"username":"testuser", "password":"testpass"}`).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	logResponse(t, resp)

	// PostJSONStringError
	resp, err = c.R().
		SetBody(`{"username":"testuser" "password":"testpass"}`).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

	logResponse(t, resp)
}

func TestPostJSONBytesSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetHeader(hdrContentTypeKey, "application/json; charset=utf-8").
		SetHeaders(map[string]string{hdrUserAgentKey: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) go-resty v0.7", hdrAcceptKey: "application/json; charset=utf-8"})

	resp, err := c.R().
		SetBody([]byte(`{"username":"testuser", "password":"testpass"}`)).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	logResponse(t, resp)
}

func TestPostJSONBytesIoReader(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetHeader(hdrContentTypeKey, "application/json; charset=utf-8")

	bodyBytes := []byte(`{"username":"testuser", "password":"testpass"}`)

	resp, err := c.R().
		SetBody(bytes.NewReader(bodyBytes)).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	logResponse(t, resp)
}

func TestPostJSONStructSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	user := &User{Username: "testuser", Password: "testpass"}

	c := dc().SetJSONEscapeHTML(false)
	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/json; charset=utf-8").
		SetBody(user).
		SetResult(&AuthSuccess{}).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

func TestPostJSONRPCStructSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	user := &User{Username: "testuser", Password: "testpass"}

	c := dc().SetJSONEscapeHTML(false)
	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/json-rpc").
		SetBody(user).
		SetResult(&AuthSuccess{}).
		SetQueryParam("ct", "rpc").
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

func TestPostJSONStructInvalidLogin(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetDebug(false)

	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/json; charset=utf-8").
		SetBody(User{Username: "testuser", Password: "testpass1"}).
		SetError(AuthError{}).
		SetJSONEscapeHTML(false).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())

	authError := resp.Error().(*AuthError)
	assert.Equal(t, "unauthorized", authError.ID)
	assert.Equal(t, "Invalid credentials", authError.Message)
	t.Logf("Result Error: %q", resp.Error().(*AuthError))

	logResponse(t, resp)
}

func TestPostJSONErrorRFC7807(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/json; charset=utf-8").
		SetBody(User{Username: "testuser", Password: "testpass1"}).
		SetError(AuthError{}).
		Post(ts.URL + "/login?ct=problem")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())

	authError := resp.Error().(*AuthError)
	assert.Equal(t, "unauthorized", authError.ID)
	assert.Equal(t, "Invalid credentials", authError.Message)
	t.Logf("Result Error: %q", resp.Error().(*AuthError))

	logResponse(t, resp)
}

func TestPostJSONMapSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetDebug(false)

	resp, err := c.R().
		SetBody(map[string]interface{}{"username": "testuser", "password": "testpass"}).
		SetResult(AuthSuccess{}).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

func TestPostJSONMapInvalidResponseJson(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	resp, err := dclr().
		SetBody(map[string]interface{}{"username": "testuser", "password": "invalidjson"}).
		SetResult(&AuthSuccess{}).
		Post(ts.URL + "/login")

	assert.Equal(t, "invalid character '}' looking for beginning of object key string", err.Error())
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	authSuccess := resp.Result().(*AuthSuccess)
	assert.Equal(t, "", authSuccess.ID)
	assert.Equal(t, "", authSuccess.Message)

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

type brokenMarshalJSON struct{}

func (b brokenMarshalJSON) MarshalJSON() ([]byte, error) {
	return nil, errors.New("b0rk3d")
}

func TestPostJSONMarshalError(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	b := brokenMarshalJSON{}
	exp := "b0rk3d"

	_, err := dclr().
		SetHeader(hdrContentTypeKey, "application/json").
		SetBody(b).
		Post(ts.URL + "/login")
	if err == nil {
		t.Fatalf("expected error but got %v", err)
	}

	if !strings.Contains(err.Error(), exp) {
		t.Errorf("expected error string %q to contain %q", err, exp)
	}
}

func TestForceContentTypeForGH276andGH240(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	retried := 0
	c := dc()
	c.SetDebug(false)
	c.SetRetryCount(3)
	c.SetRetryAfter(func(*Client, *Response) (time.Duration, error) {
		retried++
		return 0, nil
	})

	resp, err := c.R().
		SetBody(map[string]interface{}{"username": "testuser", "password": "testpass"}).
		SetResult(AuthSuccess{}).
		ForceContentType("application/json").
		Post(ts.URL + "/login-json-html")

	assert.NotNil(t, err) // expecting error due to incorrect content type from server end
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, 0, retried)

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

func TestPostXMLStringSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetDebug(false)

	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(`<?xml version="1.0" encoding="UTF-8"?><User><Username>testuser</Username><Password>testpass</Password></User>`).
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10)).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	logResponse(t, resp)
}

type brokenMarshalXML struct{}

func (b brokenMarshalXML) MarshalXML(_ *xml.Encoder, _ xml.StartElement) error {
	return errors.New("b0rk3d")
}

func TestPostXMLMarshalError(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	b := brokenMarshalXML{}
	exp := "b0rk3d"

	_, err := dclr().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(b).
		Post(ts.URL + "/login")
	if err == nil {
		t.Fatalf("expected error but got %v", err)
	}

	if !strings.Contains(err.Error(), exp) {
		t.Errorf("expected error string %q to contain %q", err, exp)
	}
}

func TestPostXMLStringError(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	resp, err := dclr().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(`<?xml version="1.0" encoding="UTF-8"?><User><Username>testuser</Username>testpass</Password></User>`).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?><AuthError><Id>bad_request</Id><Message>Unable to read user info</Message></AuthError>`, resp.String())

	logResponse(t, resp)
}

func TestPostXMLBytesSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetDebug(false)

	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody([]byte(`<?xml version="1.0" encoding="UTF-8"?><User><Username>testuser</Username><Password>testpass</Password></User>`)).
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10)).
		SetContentLength(true).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	logResponse(t, resp)
}

func TestPostXMLStructSuccess(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	resp, err := dclr().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(User{Username: "testuser", Password: "testpass"}).
		SetContentLength(true).
		SetResult(&AuthSuccess{}).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

func TestPostXMLStructInvalidLogin(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetError(&AuthError{})

	resp, err := c.R().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(User{Username: "testuser", Password: "testpass1"}).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())
	assert.Equal(t, resp.Header().Get("Www-Authenticate"), "Protected Realm")

	t.Logf("Result Error: %q", resp.Error().(*AuthError))

	logResponse(t, resp)
}

func TestPostXMLStructInvalidResponseXml(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	resp, err := dclr().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(User{Username: "testuser", Password: "invalidxml"}).
		SetResult(&AuthSuccess{}).
		Post(ts.URL + "/login")

	assert.Equal(t, "XML syntax error on line 1: element <Message> closed by </AuthSuccess>", err.Error())
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

func TestPostXMLMapNotSupported(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	_, err := dclr().
		SetHeader(hdrContentTypeKey, "application/xml").
		SetBody(map[string]interface{}{"Username": "testuser", "Password": "testpass"}).
		Post(ts.URL + "/login")

	assert.Equal(t, "unsupported 'Body' type/value", err.Error())
}

func TestRequestBasicAuth(t *testing.T) {
	ts := createAuthServer(t)
	defer ts.Close()

	c := dc()
	c.SetBaseURL(ts.URL).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	resp, err := c.R().
		SetBasicAuth("myuser", "basicauth").
		SetResult(&AuthSuccess{}).
		Post("/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))
	logResponse(t, resp)
}

func TestRequestBasicAuthFail(t *testing.T) {
	ts := createAuthServer(t)
	defer ts.Close()

	c := dc()
	c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetError(AuthError{})

	resp, err := c.R().
		SetBasicAuth("myuser", "basicauth1").
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode())

	t.Logf("Result Error: %q", resp.Error().(*AuthError))
	logResponse(t, resp)
}

func TestRequestAuthToken(t *testing.T) {
	ts := createAuthServer(t)
	defer ts.Close()

	c := dc()
	c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAuthToken("004DDB79-6801-4587-B976-F093E6AC44FF")

	resp, err := c.R().
		SetAuthToken("004DDB79-6801-4587-B976-F093E6AC44FF-Request").
		Get(ts.URL + "/profile")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestRequestAuthScheme(t *testing.T) {
	ts := createAuthServer(t)
	defer ts.Close()

	c := dc()
	c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetAuthScheme("OAuth").
		SetAuthToken("004DDB79-6801-4587-B976-F093E6AC44FF")

	resp, err := c.R().
		SetAuthScheme("Bearer").
		SetAuthToken("004DDB79-6801-4587-B976-F093E6AC44FF-Request").
		Get(ts.URL + "/profile")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestFormData(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetFormData(map[string]string{"zip_code": "00000", "city": "Los Angeles"}).
		SetContentLength(true).
		SetDebug(true)
	c.outputLogTo(io.Discard)

	resp, err := c.R().
		SetFormData(map[string]string{"first_name": "Jeevanandam", "last_name": "M", "zip_code": "00001"}).
		SetBasicAuth("myuser", "mypass").
		Post(ts.URL + "/profile")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "Success", resp.String())
}

func TestMultiValueFormData(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()

	v := url.Values{
		"search_criteria": []string{"book", "glass", "pencil"},
	}

	c := dc()
	c.SetContentLength(true).SetDebug(true)
	c.outputLogTo(io.Discard)

	resp, err := c.R().
		SetQueryParamsFromValues(v).
		Post(ts.URL + "/search")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "Success", resp.String())
}

func TestFormDataDisableWarn(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()

	c := dc()
	c.SetFormData(map[string]string{"zip_code": "00000", "city": "Los Angeles"}).
		SetContentLength(true).
		SetDebug(true).
		SetDisableWarn(true)
	c.outputLogTo(io.Discard)

	resp, err := c.R().
		SetFormData(map[string]string{"first_name": "Jeevanandam", "last_name": "M", "zip_code": "00001"}).
		SetBasicAuth("myuser", "mypass").
		Post(ts.URL + "/profile")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "Success", resp.String())
}

func TestMultiPartUploadFile(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	basePath := getTestDataPath()

	c := dc()
	c.SetFormData(map[string]string{"zip_code": "00001", "city": "Los Angeles"})

	resp, err := c.R().
		SetFile("profile_img", filepath.Join(basePath, "test-img.png")).
		SetContentLength(true).
		Post(ts.URL + "/upload")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestMultiPartUploadFileError(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	basePath := getTestDataPath()

	c := dc()
	c.SetFormData(map[string]string{"zip_code": "00001", "city": "Los Angeles"})

	resp, err := c.R().
		SetFile("profile_img", filepath.Join(basePath, "test-img-not-exists.png")).
		Post(ts.URL + "/upload")

	if err == nil {
		t.Errorf("Expected [%v], got [%v]", nil, err)
	}
	if resp != nil {
		t.Errorf("Expected [%v], got [%v]", nil, resp)
	}
}

func TestMultiPartUploadFiles(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	basePath := getTestDataPath()

	resp, err := dclr().
		SetFormDataFromValues(url.Values{
			"first_name": []string{"Jeevanandam"},
			"last_name":  []string{"M"},
		}).
		SetFiles(map[string]string{"profile_img": filepath.Join(basePath, "test-img.png"), "notes": filepath.Join(basePath, "text-file.txt")}).
		Post(ts.URL + "/upload")

	responseStr := resp.String()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(responseStr, "test-img.png"))
	assert.True(t, strings.Contains(responseStr, "text-file.txt"))
}

func TestMultiPartIoReaderFiles(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	basePath := getTestDataPath()
	profileImgBytes, _ := os.ReadFile(filepath.Join(basePath, "test-img.png"))
	notesBytes, _ := os.ReadFile(filepath.Join(basePath, "text-file.txt"))

	// Just info values
	file := File{
		Name:      "test_file_name.jpg",
		ParamName: "test_param",
		Reader:    bytes.NewBuffer([]byte("test bytes")),
	}
	t.Logf("File Info: %v", file.String())

	resp, err := dclr().
		SetFormData(map[string]string{"first_name": "Jeevanandam", "last_name": "M"}).
		SetFileReader("profile_img", "test-img.png", bytes.NewReader(profileImgBytes)).
		SetFileReader("notes", "text-file.txt", bytes.NewReader(notesBytes)).
		Post(ts.URL + "/upload")

	responseStr := resp.String()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(responseStr, "test-img.png"))
	assert.True(t, strings.Contains(responseStr, "text-file.txt"))
}

func TestMultiPartUploadFileNotOnGetOrDelete(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	basePath := getTestDataPath()

	_, err := dclr().
		SetFile("profile_img", filepath.Join(basePath, "test-img.png")).
		Get(ts.URL + "/upload")

	assert.Equal(t, "multipart content is not allowed in HTTP verb [GET]", err.Error())

	_, err = dclr().
		SetFile("profile_img", filepath.Join(basePath, "test-img.png")).
		Delete(ts.URL + "/upload")

	assert.Equal(t, "multipart content is not allowed in HTTP verb [DELETE]", err.Error())
}

func TestMultiPartFormData(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	resp, err := dclr().
		SetMultipartFormData(map[string]string{"first_name": "Jeevanandam", "last_name": "M", "zip_code": "00001"}).
		SetBasicAuth("myuser", "mypass").
		Post(ts.URL + "/profile")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "Success", resp.String())
}

func TestMultiPartMultipartField(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	jsonBytes := []byte(`{"input": {"name": "Uploaded document", "_filename" : ["file.txt"]}}`)

	resp, err := dclr().
		SetFormDataFromValues(url.Values{
			"first_name": []string{"Jeevanandam"},
			"last_name":  []string{"M"},
		}).
		SetMultipartField("uploadManifest", "upload-file.json", "application/json", bytes.NewReader(jsonBytes)).
		Post(ts.URL + "/upload")

	responseStr := resp.String()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(responseStr, "upload-file.json"))
}

func TestMultiPartMultipartFields(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/upload")

	jsonStr1 := `{"input": {"name": "Uploaded document 1", "_filename" : ["file1.txt"]}}`
	jsonStr2 := `{"input": {"name": "Uploaded document 2", "_filename" : ["file2.txt"]}}`

	fields := []*MultipartField{
		{
			Param:       "uploadManifest1",
			FileName:    "upload-file-1.json",
			ContentType: "application/json",
			Reader:      strings.NewReader(jsonStr1),
		},
		{
			Param:       "uploadManifest2",
			FileName:    "upload-file-2.json",
			ContentType: "application/json",
			Reader:      strings.NewReader(jsonStr2),
		},
		{
			Param:       "uploadManifest3",
			ContentType: "application/json",
			Reader:      strings.NewReader(jsonStr2),
		},
	}

	resp, err := dclr().
		SetFormData(map[string]string{"first_name": "Jeevanandam", "last_name": "M"}).
		SetMultipartFields(fields...).
		Post(ts.URL + "/upload")

	responseStr := resp.String()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(responseStr, "upload-file-1.json"))
	assert.True(t, strings.Contains(responseStr, "upload-file-2.json"))
}

func TestGetWithCookie(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := dcl()
	c.SetBaseURL(ts.URL)
	c.SetCookies(&http.Cookie{
		Name:  "go-resty-1",
		Value: "This is cookie 1 value",
	})

	resp, err := c.R().
		SetCookie(&http.Cookie{
			Name:  "go-resty-2",
			Value: "This is cookie 2 value",
		}).
		SetCookies([]*http.Cookie{
			{
				Name:  "go-resty-1",
				Value: "This is cookie 1 value additional append",
			},
		}).
		Get("mypage2")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "TestGet: text response from mypage2", resp.String())

	logResponse(t, resp)
}

func TestGetWithCookies(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := dc()
	c.SetBaseURL(ts.URL).SetDebug(true)

	tu, _ := url.Parse(ts.URL)
	c.GetClient().Jar.SetCookies(tu, []*http.Cookie{
		{
			Name:  "jar-go-resty-1",
			Value: "From Jar - This is cookie 1 value",
		},
		{
			Name:  "jar-go-resty-2",
			Value: "From Jar - This is cookie 2 value",
		},
	})

	resp, err := c.R().SetHeader("Cookie", "").Get("mypage2")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	// Client cookies
	c.SetCookies(
		&http.Cookie{Name: "go-resty-1", Value: "This is cookie 1 value"},
		&http.Cookie{Name: "go-resty-2", Value: "This is cookie 2 value"})

	resp, err = c.R().
		SetCookie(&http.Cookie{
			Name:  "req-go-resty-1",
			Value: "This is request cookie 1 value additional append",
		}).
		Get("mypage2")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "TestGet: text response from mypage2", resp.String())

	logResponse(t, resp)
}

func TestPutPlainString(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	resp, err := dc().R().
		SetBody("This is plain text body to server").
		Put(ts.URL + "/plaintext")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "TestPut: plain text response", resp.String())
}

func TestPutJSONString(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	client := dc()

	client.OnBeforeRequest(func(c *Client, r *Request) error {
		r.SetHeader("X-Custom-Request-Middleware", "OnBeforeRequest middleware")
		return nil
	})
	client.OnBeforeRequest(func(c *Client, r *Request) error {
		c.SetContentLength(true)
		r.SetHeader("X-ContentLength", "OnBeforeRequest ContentLength set")
		return nil
	})

	client.SetDebug(true)
	client.outputLogTo(io.Discard)

	resp, err := client.R().
		SetHeaders(map[string]string{hdrContentTypeKey: "application/json; charset=utf-8", hdrAcceptKey: "application/json; charset=utf-8"}).
		SetBody(`{"content":"json content sending to server"}`).
		Put(ts.URL + "/json")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, `{"response":"json response"}`, resp.String())
}

func TestPutXMLString(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	resp, err := dc().R().
		SetHeaders(map[string]string{hdrContentTypeKey: "application/xml", hdrAcceptKey: "application/xml"}).
		SetBody(`<?xml version="1.0" encoding="UTF-8"?><Request>XML Content sending to server</Request>`).
		Put(ts.URL + "/xml")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?><Response>XML response</Response>`, resp.String())
}

func TestOnBeforeMiddleware(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	c := dc()
	c.OnBeforeRequest(func(c *Client, r *Request) error {
		r.SetHeader("X-Custom-Request-Middleware", "OnBeforeRequest middleware")
		return nil
	})
	c.OnBeforeRequest(func(c *Client, r *Request) error {
		c.SetContentLength(true)
		r.SetHeader("X-ContentLength", "OnBeforeRequest ContentLength set")
		return nil
	})

	resp, err := c.R().
		SetBody("OnBeforeRequest: This is plain text body to server").
		Put(ts.URL + "/plaintext")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "TestPut: plain text response", resp.String())
}

func TestHTTPAutoRedirectUpTo10(t *testing.T) {
	ts := createRedirectServer(t)
	defer ts.Close()

	_, err := dc().R().Get(ts.URL + "/redirect-1")

	assert.True(t, "Get /redirect-11: stopped after 10 redirects" == err.Error() ||
		"Get \"/redirect-11\": stopped after 10 redirects" == err.Error())
}

func TestHostCheckRedirectPolicy(t *testing.T) {
	ts := createRedirectServer(t)
	defer ts.Close()

	c := dc().
		SetRedirectPolicy(DomainCheckRedirectPolicy("127.0.0.1"))

	_, err := c.R().Get(ts.URL + "/redirect-host-check-1")

	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "redirect is not allowed as per DomainCheckRedirectPolicy"))
}

func TestHeadMethod(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dclr().Head(ts.URL + "/")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestOptionsMethod(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	resp, err := dclr().Options(ts.URL + "/options")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, resp.Header().Get("Access-Control-Expose-Headers"), "x-go-resty-id")
}

func TestPatchMethod(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	resp, err := dclr().Patch(ts.URL + "/patch")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	resp.body = nil
	assert.Equal(t, "", resp.String())
}

func TestSendMethod(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	t.Run("send-get", func(t *testing.T) {
		req := dclr()
		req.Method = http.MethodGet
		req.URL = ts.URL + "/gzip-test"

		resp, err := req.Send()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		assert.Equal(t, "This is Gzip response testing", resp.String())
	})

	t.Run("send-options", func(t *testing.T) {
		req := dclr()
		req.Method = http.MethodOptions
		req.URL = ts.URL + "/options"

		resp, err := req.Send()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		assert.Equal(t, "", resp.String())
		assert.Equal(t, "x-go-resty-id", resp.Header().Get("Access-Control-Expose-Headers"))
	})

	t.Run("send-patch", func(t *testing.T) {
		req := dclr()
		req.Method = http.MethodPatch
		req.URL = ts.URL + "/patch"

		resp, err := req.Send()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		assert.Equal(t, "", resp.String())
	})

	t.Run("send-put", func(t *testing.T) {
		req := dclr()
		req.Method = http.MethodPut
		req.URL = ts.URL + "/plaintext"

		resp, err := req.Send()

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode())

		assert.Equal(t, "TestPut: plain text response", resp.String())
	})
}

func TestRawFileUploadByBody(t *testing.T) {
	ts := createFormPostServer(t)
	defer ts.Close()

	file, err := os.Open(filepath.Join(getTestDataPath(), "test-img.png"))
	assert.Nil(t, err)
	fileBytes, err := io.ReadAll(file)
	assert.Nil(t, err)

	resp, err := dclr().
		SetBody(fileBytes).
		SetContentLength(true).
		SetAuthToken("004DDB79-6801-4587-B976-F093E6AC44FF").
		Put(ts.URL + "/raw-upload")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "image/png", resp.Request.Header.Get(hdrContentTypeKey))
}

func TestProxySetting(t *testing.T) {
	c := dc()

	transport, err := c.transport()

	assert.Nil(t, err)

	assert.Equal(t, false, c.IsProxySet())
	assert.NotNil(t, transport.Proxy)

	c.SetProxy("http://sampleproxy:8888")
	assert.True(t, c.IsProxySet())
	assert.NotNil(t, transport.Proxy)

	c.SetProxy("//not.a.user@%66%6f%6f.com:8888")
	assert.True(t, c.IsProxySet())
	assert.NotNil(t, transport.Proxy)

	c.SetProxy("http://sampleproxy:8888")
	assert.True(t, c.IsProxySet())
	c.RemoveProxy()
	assert.Nil(t, c.proxyURL)
	assert.Nil(t, transport.Proxy)
}

func TestGetClient(t *testing.T) {
	client := New()
	custom := New()
	customClient := custom.GetClient()

	assert.NotNil(t, customClient)
	assert.NotEqual(t, client, http.DefaultClient)
	assert.NotEqual(t, customClient, http.DefaultClient)
	assert.NotEqual(t, client, customClient)
}

func TestIncorrectURL(t *testing.T) {
	c := dc()
	_, err := c.R().Get("//not.a.user@%66%6f%6f.com/just/a/path/also")
	assert.True(t, strings.Contains(err.Error(), "parse //not.a.user@%66%6f%6f.com/just/a/path/also") ||
		strings.Contains(err.Error(), "parse \"//not.a.user@%66%6f%6f.com/just/a/path/also\""))

	c.SetBaseURL("//not.a.user@%66%6f%6f.com")
	_, err1 := c.R().Get("/just/a/path/also")
	assert.True(t, strings.Contains(err1.Error(), "parse //not.a.user@%66%6f%6f.com/just/a/path/also") ||
		strings.Contains(err1.Error(), "parse \"//not.a.user@%66%6f%6f.com/just/a/path/also\""))
}

func TestDetectContentTypeForPointer(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	user := &User{Username: "testuser", Password: "testpass"}

	resp, err := dclr().
		SetBody(user).
		SetResult(AuthSuccess{}).
		Post(ts.URL + "/login")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())

	t.Logf("Result Success: %q", resp.Result().(*AuthSuccess))

	logResponse(t, resp)
}

type ExampleUser struct {
	FirstName string `json:"frist_name"`
	LastName  string `json:"last_name"`
	ZipCode   string `json:"zip_code"`
}

func TestDetectContentTypeForPointerWithSlice(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	users := &[]ExampleUser{
		{FirstName: "firstname1", LastName: "lastname1", ZipCode: "10001"},
		{FirstName: "firstname2", LastName: "lastname3", ZipCode: "10002"},
		{FirstName: "firstname3", LastName: "lastname3", ZipCode: "10003"},
	}

	resp, err := dclr().
		SetBody(users).
		Post(ts.URL + "/users")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode())

	t.Logf("Result Success: %q", resp)

	logResponse(t, resp)
}

func TestDetectContentTypeForPointerWithSliceMap(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	usersmap := map[string]interface{}{
		"user1": ExampleUser{FirstName: "firstname1", LastName: "lastname1", ZipCode: "10001"},
		"user2": &ExampleUser{FirstName: "firstname2", LastName: "lastname3", ZipCode: "10002"},
		"user3": ExampleUser{FirstName: "firstname3", LastName: "lastname3", ZipCode: "10003"},
	}

	var users []map[string]interface{}
	users = append(users, usersmap)

	resp, err := dclr().
		SetBody(&users).
		Post(ts.URL + "/usersmap")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode())

	t.Logf("Result Success: %q", resp)

	logResponse(t, resp)
}

func TestDetectContentTypeForSlice(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	users := []ExampleUser{
		{FirstName: "firstname1", LastName: "lastname1", ZipCode: "10001"},
		{FirstName: "firstname2", LastName: "lastname3", ZipCode: "10002"},
		{FirstName: "firstname3", LastName: "lastname3", ZipCode: "10003"},
	}

	resp, err := dclr().
		SetBody(users).
		Post(ts.URL + "/users")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode())

	t.Logf("Result Success: %q", resp)

	logResponse(t, resp)
}

func TestMultiParamsQueryString(t *testing.T) {
	ts1 := createGetServer(t)
	defer ts1.Close()

	client := dc()
	req1 := client.R()

	client.SetQueryParam("status", "open")

	_, _ = req1.SetQueryParam("status", "pending").
		Get(ts1.URL)

	assert.True(t, strings.Contains(req1.URL, "status=pending"))
	// pending overrides open
	assert.Equal(t, false, strings.Contains(req1.URL, "status=open"))

	_, _ = req1.SetQueryParam("status", "approved").
		Get(ts1.URL)

	assert.True(t, strings.Contains(req1.URL, "status=approved"))
	// approved overrides pending
	assert.Equal(t, false, strings.Contains(req1.URL, "status=pending"))

	ts2 := createGetServer(t)
	defer ts2.Close()

	req2 := client.R()

	v := url.Values{
		"status": []string{"pending", "approved", "reject"},
	}

	_, _ = req2.SetQueryParamsFromValues(v).Get(ts2.URL)

	assert.True(t, strings.Contains(req2.URL, "status=pending"))
	assert.True(t, strings.Contains(req2.URL, "status=approved"))
	assert.True(t, strings.Contains(req2.URL, "status=reject"))

	// because it's removed by key
	assert.Equal(t, false, strings.Contains(req2.URL, "status=open"))
}

func TestSetQueryStringTypical(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dclr().
		SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more").
		Get(ts.URL)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "200 OK", resp.Status())
	assert.Equal(t, "TestGet: text response", resp.String())

	resp, err = dclr().
		SetQueryString("&%%amp;").
		Get(ts.URL)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "200 OK", resp.Status())
	assert.Equal(t, "TestGet: text response", resp.String())
}

func TestSetHeaderVerbatim(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	r := dclr().
		SetHeaderVerbatim("header-lowercase", "value_lowercase").
		SetHeader("header-lowercase", "value_standard")

	assert.Equal(t, "value_lowercase", strings.Join(r.Header["header-lowercase"], "")) //nolint
	assert.Equal(t, "value_standard", r.Header.Get("Header-Lowercase"))
}

func TestSetHeaderMultipleValue(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	r := dclr().
		SetHeaderMultiValues(map[string][]string{
			"Content":       {"text/*", "text/html", "*"},
			"Authorization": {"Bearer xyz"},
		})
	assert.Equal(t, "text/*, text/html, *", r.Header.Get("content"))
	assert.Equal(t, "Bearer xyz", r.Header.Get("authorization"))
}

func TestOutputFileWithBaseDirAndRelativePath(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()
	defer cleanupFiles("testdata/dir-sample")

	client := dc().
		SetRedirectPolicy(FlexibleRedirectPolicy(10)).
		SetOutputDirectory(filepath.Join(getTestDataPath(), "dir-sample")).
		SetDebug(true)
	client.outputLogTo(io.Discard)

	resp, err := client.R().
		SetOutput("go-resty/test-img-success.png").
		Get(ts.URL + "/my-image.png")

	assert.Nil(t, err)
	assert.True(t, resp.Size() != 0)
	assert.True(t, resp.Time() > 0)
}

func TestOutputFileWithBaseDirError(t *testing.T) {
	c := dc().SetRedirectPolicy(FlexibleRedirectPolicy(10)).
		SetOutputDirectory(filepath.Join(getTestDataPath(), `go-resty\0`))

	_ = c
}

func TestOutputPathDirNotExists(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()
	defer cleanupFiles(filepath.Join("testdata", "not-exists-dir"))

	client := dc().
		SetRedirectPolicy(FlexibleRedirectPolicy(10)).
		SetOutputDirectory(filepath.Join(getTestDataPath(), "not-exists-dir"))

	resp, err := client.R().
		SetOutput("test-img-success.png").
		Get(ts.URL + "/my-image.png")

	assert.Nil(t, err)
	assert.True(t, resp.Size() != 0)
	assert.True(t, resp.Time() > 0)
}

func TestOutputFileAbsPath(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()
	defer cleanupFiles(filepath.Join("testdata", "go-resty"))

	_, err := dcr().
		SetOutput(filepath.Join(getTestDataPath(), "go-resty", "test-img-success-2.png")).
		Get(ts.URL + "/my-image.png")

	assert.Nil(t, err)
}

func TestContextInternal(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	r := dc().R().
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10))

	resp, err := r.Get(ts.URL + "/")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestSRV(t *testing.T) {
	c := dc().
		SetRedirectPolicy(FlexibleRedirectPolicy(20)).
		SetScheme("http")

	r := c.R().SetSRV(&SRVRecord{Service: "xmpp-server", Domain: "google.com"})

	assert.Equal(t, "xmpp-server", r.SRV.Service)
	assert.Equal(t, "google.com", r.SRV.Domain)

	resp, err := r.Get("/")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestSRVInvalidService(t *testing.T) {
	_, err := dc().R().
		SetSRV(&SRVRecord{"nonexistantservice", "sampledomain"}).
		Get("/")

	assert.NotNil(t, err)
	_, ok := err.(*net.DNSError)
	assert.True(t, ok)
}

func TestRequestDoNotParseResponse(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	client := dc().SetDoNotParseResponse(true)
	resp, err := client.R().
		SetQueryParam("request_no", strconv.FormatInt(time.Now().Unix(), 10)).
		Get(ts.URL + "/")

	assert.Nil(t, err)

	buf := acquireBuffer()
	defer releaseBuffer(buf)
	_, _ = io.Copy(buf, resp.RawBody())

	assert.Equal(t, "TestGet: text response", buf.String())
	_ = resp.RawBody().Close()

	// Manually setting RawResponse as nil
	resp, err = dc().R().
		SetDoNotParseResponse(true).
		Get(ts.URL + "/")

	assert.Nil(t, err)

	resp.RawResponse = nil
	assert.Nil(t, resp.RawBody())
}

type noCtTest struct {
	Response string `json:"response"`
}

func TestRequestExpectContentTypeTest(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	c := dc()
	resp, err := c.R().
		SetResult(noCtTest{}).
		ExpectContentType("application/json").
		Get(ts.URL + "/json-no-set")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.NotNil(t, resp.Result())
	assert.Equal(t, "json response no content type set", resp.Result().(*noCtTest).Response)

	assert.Equal(t, "", firstNonEmpty("", ""))
}

func TestGetPathParamAndPathParams(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := dc().
		SetBaseURL(ts.URL).
		SetPathParam("userId", "sample@sample.com")

	resp, err := c.R().SetPathParam("subAccountId", "100002").
		Get("/v1/users/{userId}/{subAccountId}/details")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(resp.String(), "TestGetPathParams: text response"))
	assert.True(t, strings.Contains(resp.String(), "/v1/users/sample@sample.com/100002/details"))

	logResponse(t, resp)
}

func TestReportMethodSupportsPayload(t *testing.T) {
	ts := createGenServer(t)
	defer ts.Close()

	c := dc()
	resp, err := c.R().
		SetBody("body").
		Execute("REPORT", ts.URL+"/report")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestRequestQueryStringOrder(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := New().R().
		SetQueryString("productId=232&template=fresh-sample&cat=resty&source=google&kw=buy a lot more").
		Get(ts.URL + "/?UniqueId=ead1d0ed-XXX-XXX-XXX-abb7612b3146&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLC...HZEhwVnJ1d0NSUGVLaUpSaVNLRG5scz0&ApiVersion=2.0")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "200 OK", resp.Status())
	assert.NotNil(t, resp.Body())
	assert.Equal(t, "TestGet: text response", resp.String())

	logResponse(t, resp)
}

func TestRequestOverridesClientAuthorizationHeader(t *testing.T) {
	ts := createAuthServer(t)
	defer ts.Close()

	c := dc()
	c.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}).
		SetHeader("Authorization", "some token").
		SetBaseURL(ts.URL + "/")

	resp, err := c.R().
		SetHeader("Authorization", "Bearer 004DDB79-6801-4587-B976-F093E6AC44FF").
		Get("/profile")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestRequestFileUploadAsReader(t *testing.T) {
	ts := createFilePostServer(t)
	defer ts.Close()

	file, _ := os.Open(filepath.Join(getTestDataPath(), "test-img.png"))
	defer iox.Close(file)

	resp, err := dclr().
		SetBody(file).
		SetHeader("Content-Type", "image/png").
		Post(ts.URL + "/upload")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(resp.String(), "File Uploaded successfully"))

	file, _ = os.Open(filepath.Join(getTestDataPath(), "test-img.png"))
	defer iox.Close(file)

	resp, err = dclr().
		SetBody(file).
		SetHeader("Content-Type", "image/png").
		SetContentLength(true).
		Post(ts.URL + "/upload")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(resp.String(), "File Uploaded successfully"))
}

func TestHostHeaderOverride(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	resp, err := dc().R().
		SetHeader("Host", "myhostname").
		Get(ts.URL + "/host-header")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, "200 OK", resp.Status())
	assert.NotNil(t, resp.Body())
	assert.Equal(t, "myhostname", resp.String())

	logResponse(t, resp)
}

func TestPathParamURLInput(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	c := dc().SetDebug(true).
		SetBaseURL(ts.URL).
		SetPathParams(map[string]string{
			"userId": "sample@sample.com",
		})

	resp, err := c.R().
		SetPathParams(map[string]string{
			"subAccountId": "100002",
			"website":      "https://example.com",
		}).Get("/v1/users/{userId}/{subAccountId}/{website}")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.True(t, strings.Contains(resp.String(), "TestPathParamURLInput: text response"))
	assert.True(t, strings.Contains(resp.String(), "/v1/users/sample@sample.com/100002/https:%2F%2Fexample.com"))

	logResponse(t, resp)
}

// This test case is kind of pass always
func TestTraceInfo(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	serverAddr := ts.URL[strings.LastIndex(ts.URL, "/")+1:]

	client := dc()
	client.SetBaseURL(ts.URL).SetTrace(true)
	for _, u := range []string{"/", "/json", "/long-text", "/long-json"} {
		resp, err := client.R().Get(u)
		assert.Nil(t, err)
		assert.NotNil(t, resp)

		tr := resp.Request.TraceInfo()
		assert.True(t, tr.DNSLookup >= 0)
		assert.True(t, tr.ConnTime >= 0)
		assert.True(t, tr.TLSHandshake >= 0)
		assert.True(t, tr.ServerTime >= 0)
		assert.True(t, tr.ResponseTime >= 0)
		assert.True(t, tr.TotalTime >= 0)
		assert.True(t, tr.TotalTime < time.Hour)
		assert.True(t, tr.TotalTime == resp.Time())
		assert.Equal(t, tr.RemoteAddr.String(), serverAddr)
	}

	client.SetTrace(false)

	for _, u := range []string{"/", "/json", "/long-text", "/long-json"} {
		resp, err := client.R().SetTrace(true).Get(u)
		assert.Nil(t, err)
		assert.NotNil(t, resp)

		tr := resp.Request.TraceInfo()
		assert.True(t, tr.DNSLookup >= 0)
		assert.True(t, tr.ConnTime >= 0)
		assert.True(t, tr.TLSHandshake >= 0)
		assert.True(t, tr.ServerTime >= 0)
		assert.True(t, tr.ResponseTime >= 0)
		assert.True(t, tr.TotalTime >= 0)
		assert.True(t, tr.TotalTime == resp.Time())
		assert.Equal(t, tr.RemoteAddr.String(), serverAddr)
	}

	// for sake of hook funcs
	_, _ = client.R().SetTrace(true).Get("https://httpbin.org/get")
}

func TestTraceInfoWithoutEnableTrace(t *testing.T) {
	ts := createGetServer(t)
	defer ts.Close()

	client := dc()
	client.SetBaseURL(ts.URL)
	for _, u := range []string{"/", "/json", "/long-text", "/long-json"} {
		resp, err := client.R().Get(u)
		assert.Nil(t, err)
		assert.NotNil(t, resp)

		tr := resp.Request.TraceInfo()
		assert.True(t, tr.DNSLookup == 0)
		assert.True(t, tr.ConnTime == 0)
		assert.True(t, tr.TLSHandshake == 0)
		assert.True(t, tr.ServerTime == 0)
		assert.True(t, tr.ResponseTime == 0)
		assert.True(t, tr.TotalTime == 0)
	}
}

func TestTraceInfoOnTimeout(t *testing.T) {
	client := dc()
	client.SetBaseURL("http://resty-nowhere.local").SetTrace(true)

	resp, err := client.R().Get("/")
	assert.NotNil(t, err)
	assert.NotNil(t, resp)

	tr := resp.Request.TraceInfo()
	assert.True(t, tr.DNSLookup >= 0)
	assert.True(t, tr.ConnTime == 0)
	assert.True(t, tr.TLSHandshake == 0)
	assert.True(t, tr.TCPConnTime == 0)
	assert.True(t, tr.ServerTime == 0)
	assert.True(t, tr.ResponseTime == 0)
	assert.True(t, tr.TotalTime > 0)
	assert.True(t, tr.TotalTime == resp.Time())
}

func TestDebugLoggerRequestBodyTooLarge(t *testing.T) {
	ts := createFilePostServer(t)
	defer ts.Close()

	debugBodySizeLimit := int64(512)

	// upload an image with more than 512 bytes
	output := bytes.NewBufferString("")
	resp, err := New().SetDebug(true).outputLogTo(output).SetDebugBodyLimit(debugBodySizeLimit).R().
		SetFile("file", filepath.Join(getTestDataPath(), "test-img.png")).
		SetHeader("Content-Type", "image/png").
		Post(ts.URL + "/upload")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, strings.Contains(output.String(), "REQUEST TOO LARGE"))

	// upload a text file with no more than 512 bytes
	output = bytes.NewBufferString("")
	resp, err = New().SetDebug(true).outputLogTo(output).SetDebugBodyLimit(debugBodySizeLimit).R().
		SetFile("file", filepath.Join(getTestDataPath(), "text-file.txt")).
		SetHeader("Content-Type", "text/plain").
		Post(ts.URL + "/upload")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, strings.Contains(output.String(), " THIS IS TEXT FILE FOR MULTIPART UPLOAD TEST "))

	formTs := createFormPostServer(t)
	defer formTs.Close()

	// post form with more than 512 bytes data
	output = bytes.NewBufferString("")
	resp, err = New().SetDebug(true).outputLogTo(output).SetDebugBodyLimit(debugBodySizeLimit).R().
		SetFormData(map[string]string{
			"first_name": "Alex",
			"last_name":  strings.Repeat("C", int(debugBodySizeLimit)),
			"zip_code":   "00001",
		}).
		SetBasicAuth("myuser", "mypass").
		Post(formTs.URL + "/profile")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, strings.Contains(output.String(), "REQUEST TOO LARGE"))

	// post form with no more than 512 bytes data
	output = bytes.NewBufferString("")
	resp, err = New().SetDebug(true).outputLogTo(output).SetDebugBodyLimit(debugBodySizeLimit).R().
		SetFormData(map[string]string{
			"first_name": "Alex",
			"last_name":  "C",
			"zip_code":   "00001",
		}).
		SetBasicAuth("myuser", "mypass").
		Post(formTs.URL + "/profile")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, strings.Contains(output.String(), "Alex"))

	// post string with more than 512 bytes data
	output = bytes.NewBufferString("")
	resp, err = New().SetDebug(true).outputLogTo(output).SetDebugBodyLimit(debugBodySizeLimit).R().
		SetBody(`{
			"first_name": "Alex",
			"last_name": "`+strings.Repeat("C", int(debugBodySizeLimit))+`C",
			"zip_code": "00001"}`).
		SetBasicAuth("myuser", "mypass").
		Post(formTs.URL + "/profile")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, strings.Contains(output.String(), "REQUEST TOO LARGE"))

	// post slice with more than 512 bytes data
	output = bytes.NewBufferString("")
	resp, err = New().SetDebug(true).outputLogTo(output).SetDebugBodyLimit(debugBodySizeLimit).R().
		SetBody([]string{strings.Repeat("C", int(debugBodySizeLimit))}).
		SetBasicAuth("myuser", "mypass").
		Post(formTs.URL + "/profile")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.True(t, strings.Contains(output.String(), "REQUEST TOO LARGE"))
}

func TestPostMapTemporaryRedirect(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	resp, err := c.R().SetBody(map[string]string{"username": "testuser", "password": "testpass"}).
		Post(ts.URL + "/redirect")

	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

type brokenReadCloser struct{}

func (b brokenReadCloser) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (b brokenReadCloser) Close() error {
	return nil
}

func TestPostBodyError(t *testing.T) {
	ts := createPostServer(t)
	defer ts.Close()

	c := dc()
	resp, err := c.R().SetBody(brokenReadCloser{}).Post(ts.URL + "/redirect")
	assert.NotNil(t, err)
	assert.Equal(t, "read error", err.Error())
	assert.Nil(t, resp)
}
