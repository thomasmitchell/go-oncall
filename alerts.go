package oncall

import (
	"encoding/json"
	"net/url"
	"time"
)

const alertPath = "alerts"

type Alert struct {
	ID           string
	AlertGroupID string
	CreatedAt    time.Time
	Payload      map[string]interface{}
}

type alertRaw struct {
	ID           string                 `json:"id"`
	AlertGroupID string                 `json:"alert_group_id"`
	CreatedAt    string                 `json:"created_at"`
	Payload      map[string]interface{} `json:"payload"`
}

func (a *Alert) UnmarshalJSON(b []byte) error {
	raw := alertRaw{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	parsedCreatedAt, err := timeFromString(raw.CreatedAt)
	if err != nil {
		return err
	}

	*a = Alert{
		ID:           raw.ID,
		AlertGroupID: raw.AlertGroupID,
		CreatedAt:    parsedCreatedAt,
		Payload:      raw.Payload,
	}

	return nil
}

func (a *Alert) MarshalJSON() ([]byte, error) {
	if a == nil {
		return []byte("null"), nil
	}

	return json.Marshal(&alertRaw{
		ID:           a.ID,
		AlertGroupID: a.AlertGroupID,
		CreatedAt:    timeToString(a.CreatedAt),
		Payload:      a.Payload,
	})
}

type ListAlertFilter struct {
	AlertGroupID string
	Search       string
}

func (c *Client) ListAlertsByPage(page int, filter *ListAlertFilter) (*PaginatedResponse[Alert], error) {
	values := url.Values{}
	if filter != nil {
		if filter.AlertGroupID != "" {
			values.Set("alert_group_id", filter.AlertGroupID)
		}

		if filter.Search != "" {
			values.Set("search", filter.Search)
		}
	}

	return getPage[Alert](c, page, alertPath, values)
}

func (c *Client) ListAlerts(filter *ListAlertFilter) ([]Alert, error) {
	return paginate(c.ListAlertsByPage, filter)
}
