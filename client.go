package oncall

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Client provides functions that access and abstract the Grafana OnCall API.
// URL must be set to the for the client to work.
type Client struct {
	AuthToken string
	URL       *url.URL
	//If Client is nil, http.DefaultClient will be used
	Client *http.Client
	//If Trace is non-nil, information about HTTP requests will be given into the
	//Writer.
	Trace io.Writer
}

type PaginatedResponse[T any] struct {
	//Count is the total number of items. It can be 0 if a request does not return
	//any data.
	Count int `json:"count"`
	//Next is a link to the next page. It can be empty if the next page does not
	//contain any data.
	Next string `json:"next"`
	//Previous is a link to the previous page. It can be empty if the previous
	//page does not contain any data.
	Previous string `json:"previous"`

	Results []T `json:"results"`
}

// URL encoded values can be given as a *url.Values as "input" when performing
// a GET call
func (c *Client) doRequest(
	method, path string,
	input interface{},
	output interface{}) error {

	var query url.Values
	var body io.Reader
	if input != nil {
		if strings.ToUpper(method) == "GET" {
			//Input has to be a url.Values
			query = input.(url.Values)
		} else {
			body = &bytes.Buffer{}
			err := json.NewEncoder(body.(*bytes.Buffer)).Encode(input)
			if err != nil {
				return err
			}
		}
	}

	resp, err := c.Curl(method, path, query, body)
	if err != nil {
		return err
	}
	defer func() {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("non-2xx response code: %s", resp.Status)
	}

	if output != nil {
		err = json.NewDecoder(resp.Body).Decode(output)
		if err != nil {
			return err
		}
	}

	return nil
}

// Curl takes the given path, prepends <URL>/api/v1/ to it, and makes the request
// with the remainder of the given parameters. Errors returned only reflect
// transport errors, not HTTP semantic errors
func (c *Client) Curl(method string, path string, urlQuery url.Values, body io.Reader) (*http.Response, error) {
	//Setup URL
	u := *c.URL
	pathPrefix := strings.Trim(u.Path, "/")
	if pathPrefix != "" {
		pathPrefix = u.Path + "/"
	}
	u.Path = fmt.Sprintf("/%sapi/v1/%s", pathPrefix, strings.Trim(path, "/"))
	if u.Port() == "" {
		u.Host = fmt.Sprintf("%s:443", u.Host)
	}
	u.RawQuery = urlQuery.Encode()

	//Do the request
	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	if c.Trace != nil {
		dump, _ := httputil.DumpRequest(req, true)
		_, _ = c.Trace.Write([]byte(fmt.Sprintf("Request:\n%s\n", dump)))
	}

	req.Header.Set("Authorization", c.AuthToken)

	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}

	if client.CheckRedirect == nil {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			req.Header.Set("Authorization", c.AuthToken)
			return nil
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if c.Trace != nil {
		dump, _ := httputil.DumpResponse(resp, true)
		_, _ = c.Trace.Write([]byte(fmt.Sprintf("Response:\n%s\n", dump)))
	}

	return resp, nil
}
