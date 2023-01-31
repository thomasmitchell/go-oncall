package oncall

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

const isoTimeLayout = "2006-01-02T15:04:05Z"
const isoTimeOfDayLayout = "15:04:05Z"

func timeToString(t time.Time) string            { return t.Format(isoTimeLayout) }
func timeOfDayToString(t time.Time) string       { return t.Format(isoTimeOfDayLayout) }
func timeFromString(s string) (time.Time, error) { return time.Parse(isoTimeLayout, s) }

func buildPath(segments ...string) string {
	sanitized := make([]string, len(segments))
	for i, segment := range segments {
		sanitized[i] = url.PathEscape(segment)
	}

	return strings.Join(sanitized, "/")
}

func paginate[T any, F any](
	fn func(page int, filter *F) (*PaginatedResponse[T], error),
	filter *F,
) ([]T, error) {

	var page int
	var ret []T
	pageResp, err := fn(page, filter)
	if err != nil {
		return nil, err
	}

	ret = make([]T, 0, pageResp.Count)
	ret = append(ret, pageResp.Results...)

	for pageResp.Next != "" {
		page++
		pageResp, err = fn(page, filter)
		if err != nil {
			return nil, err
		}

		ret = append(ret, pageResp.Results...)
	}

	return ret, nil
}

func getPage[T any](c *Client, page int, path string, vals url.Values) (*PaginatedResponse[T], error) {
	if page > 0 {
		vals.Set("page", strconv.Itoa(page))
	}

	ret := &PaginatedResponse[T]{}
	err := c.doRequest("GET", path, vals, &ret)
	return ret, err
}
