package resthttp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

const (
	// HeaderKeyRequestID request id header key
	headerKeyRequestID = "X-Request-Id"
)

var runOnce sync.Once
var restyClient *resty.Client

// Client resty client
func Client() *resty.Client {
	runOnce.Do(func() {
		restyClient = resty.New().
			SetHeader("Content-Type", "application/json").
			SetHeader("Charset", "utf-8").
			SetTimeout(10 * time.Second)
	})

	return restyClient
}

// Request new resty request
func Request(ctx context.Context) *resty.Request {
	return Client().R().SetContext(ctx)
}

// WithRequestID resty request with request id
func WithRequestID(ctx context.Context, requestID string) *resty.Request {
	return Request(ctx).SetHeader(headerKeyRequestID, requestID)
}

// Execute do network request
func Execute(request *resty.Request, method, url string, body interface{}, resp interface{}) (int, error) {
	logrus.Infoln("url:%s\n", url)

	if body != nil {
		request = request.SetBody(body)
	}

	r, err := request.Execute(strings.ToUpper(method), url)
	if err != nil {
		return r.StatusCode(), err
	}

	logrus.Infoln("resp.status:%s\n", r.Status())

	return r.StatusCode(), ParseResponse(r, resp)
}

// ParseResponse parse response
func ParseResponse(r *resty.Response, obj interface{}) error {
	//fail
	if !r.IsSuccess() {
		err := json.Unmarshal(r.Body(), obj)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(r.Body()))
	}

	//success
	if obj != nil {
		e := json.Unmarshal(r.Body(), obj)
		if e != nil {
			fmt.Printf("parseResponse:%s", e.Error())
			return e
		}
		return nil
	}
	return nil
}
