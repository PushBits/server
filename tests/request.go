// Package tests provides definitions and functionality related to unit and integration tests.
package tests

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
)

// Request holds information for a HTTP request
type Request struct {
	Name         string
	Method       string
	Endpoint     string
	Data         interface{}
	Headers      map[string]string
	ShouldStatus int
	ShouldReturn interface{}
}

// GetRequest returns a ResponseRecorder and gin context according to the data set in the Request.
// String data is passed as is, all other data types are marshaled before.
func (r *Request) GetRequest() (w *httptest.ResponseRecorder, c *gin.Context, err error) {
	w = httptest.NewRecorder()
	var body io.Reader

	switch data := r.Data.(type) {
	case string:
		body = strings.NewReader(data)
	default:
		dataMarshaled, err := json.Marshal(data)
		if err != nil {
			return nil, nil, err
		}
		body = strings.NewReader(string(dataMarshaled))
	}

	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(r.Method, r.Endpoint, body)

	for name, value := range r.Headers {
		c.Request.Header.Set(name, value)
	}

	return w, c, nil
}
