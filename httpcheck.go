package httpcheck

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ivpusic/golog"
	"github.com/stretchr/testify/assert"
)

type (
	Checker struct {
		t        *testing.T
		handler  http.Handler
		request  *http.Request
		response *http.Response
		cookies  map[string]string
	}

	Callback func(*http.Response)
)

var (
	logger = golog.GetLogger("github.com/ivpusic/httpcheck")
)

func New(t *testing.T, handler http.Handler) *Checker {
	logger.Level = golog.INFO

	instance := &Checker{
		t:       t,
		handler: handler,
		cookies: map[string]string{},
	}

	return instance
}

// make request /////////////////////////////////////////////////

// If you want to provide you custom http.Request instance, you can do it using this method
// In this case internal http.Request instance won't be created, and passed instane will be used
// for making request
func (c *Checker) TestRequest(request *http.Request) *Checker {
	assert.NotNil(c.t, request, "Request nil")

	c.request = request
	return c
}

// Prepare for testing some part of code which lives on provided path and method.
func (c *Checker) Test(method, path string) *Checker {
	method = strings.ToUpper(method)
	request, err := http.NewRequest(method, path, nil)

	assert.Nil(c.t, err, "Failed to make new request")

	c.request = request
	return c
}

// headers ///////////////////////////////////////////////////////

// Will put header on request
func (c *Checker) WithHeader(key, value string) *Checker {
	c.request.Header.Set(key, value)
	return c
}

// Will check if response contains header on provided key with provided value
func (c *Checker) HasHeader(key, expectedValue string) *Checker {
	value := c.response.Header.Get(key)
	assert.Exactly(c.t, expectedValue, value)

	return c
}

// cookies ///////////////////////////////////////////////////////

// Will put cookie on request
func (c *Checker) HasCookie(key, expectedValue string) *Checker {
	value, exists := c.cookies[key]
	assert.True(c.t, exists && expectedValue == value)
	return c
}

// Will ckeck if response contains cookie with provided key and value
func (c *Checker) WithCookie(key, value string) *Checker {
	c.request.AddCookie(&http.Cookie{
		Name:  key,
		Value: value,
	})

	return c
}

// status ////////////////////////////////////////////////////////

// Will ckeck if response status is equal to provided
func (c *Checker) HasStatus(status int) *Checker {
	assert.Exactly(c.t, status, c.response.StatusCode)
	return c
}

// json body /////////////////////////////////////////////////////

// Will add the json-encoded struct to the body
func (c *Checker) WithJson(value interface{}) *Checker {
	encoded, err := json.Marshal(value)
	assert.Nil(c.t, err)
	return c.WithBody(encoded)
}

// Will ckeck if body contains json with provided value
func (c *Checker) HasJson(value interface{}) *Checker {
	body, err := ioutil.ReadAll(c.response.Body)
	assert.Nil(c.t, err)

	valueBytes, err := json.Marshal(value)
	assert.Nil(c.t, err)
	assert.Equal(c.t, string(valueBytes), string(body))

	return c
}

// xml //////////////////////////////////////////////////////////

// Adds a XML encoded body to the request
func (c *Checker) WithXml(value interface{}) *Checker {
	encoded, err := xml.Marshal(value)
	assert.Nil(c.t, err)
	return c.WithBody(encoded)
}

// Will ckeck if body contains xml with provided value
func (c *Checker) HasXml(value interface{}) *Checker {
	body, err := ioutil.ReadAll(c.response.Body)
	assert.Nil(c.t, err)

	valueBytes, err := xml.Marshal(value)
	assert.Nil(c.t, err)
	assert.Equal(c.t, string(valueBytes), string(body))

	return c
}

// body //////////////////////////////////////////////////////////

// Adds the []byte data to the body
func (c *Checker) WithBody(body []byte) *Checker {
	c.request.Body = newClosingBuffer(body)
	return c
}

// Will check if body contains provided []byte data
func (c *Checker) HasBody(body []byte) *Checker {
	responseBody, err := ioutil.ReadAll(c.response.Body)

	assert.Nil(c.t, err)
	assert.Equal(c.t, body, responseBody)

	return c
}

// Adds the string to the body
func (c *Checker) WithString(body string) *Checker {
	c.request.Body = newClosingBufferString(body)
	return c
}

// Convenience wrapper for HasBody
// Checks if body is equal to the given string
func (c *Checker) HasString(body string) *Checker {
	return c.HasBody([]byte(body))
}

func (c *Checker) handleCookies(r *http.Response) {
	if header, exist := r.Header["Set-Cookie"]; exist {
		for _, str := range header {
			if ind := strings.Index(str, "="); ind > 0 {
				c.cookies[str[0:ind]] = str[ind+1 : len(str)]
			} else {
				panic("did not find = in cookie string")
			}
		}
	}
}

func (c *Checker) generateCookieString() string {
	str := ""
	for name, val := range c.cookies {
		str += fmt.Sprintf("%s=%s;", name, val)
	}
	return str
}

// Will make reqeust to built request object.
// After request is made, it will save response object for future assertions
// Responsibility of this method is also to start and stop HTTP server
func (c *Checker) Check() *Checker {

	// set cookies
	c.request.Header.Set("Cookie", c.generateCookieString())

	recorder := httptest.NewRecorder()
	c.handler.ServeHTTP(recorder, c.request)

	resp := &http.Response{
		StatusCode: recorder.Code,
		Body:       NewReadCloser(recorder.Body),
		Header:     recorder.Header(),
	}
	c.handleCookies(resp)
	c.response = resp

	return c
}

// Will call provided callback function with current response
func (c *Checker) Cb(cb Callback) {
	cb(c.response)
}
