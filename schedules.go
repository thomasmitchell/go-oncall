package oncall

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const schedulePath = "schedules"

type Schedule struct {
	ID               string
	Name             string
	TeamID           string
	TimeZone         *time.Location
	ICalOverridesURL string
	Slack            ScheduleSlackMetadata
	Calendar         ScheduleCalendar
}

type scheduleRawHeaders struct {
	ID               string                `json:"id"`
	Name             string                `json:"name"`
	TeamID           string                `json:"team_id"`
	TimeZone         string                `json:"time_zone"`
	ICalURLOverrides string                `json:"ical_url_overrides"`
	Slack            ScheduleSlackMetadata `json:"slack"`
	Type             string                `json:"type"`
}

func (s *Schedule) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}

	out := map[string]interface{}{}
	if s.Calendar != nil {
		inter, err := json.Marshal(&s.Calendar)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(inter, &out)
		if err != nil {
			return nil, err
		}

		out["type"] = s.Calendar.ScheduleCalendarType().String()
	}

	if s.ID != "" {
		out["id"] = s.ID
	}
	out["name"] = s.Name
	if s.TeamID != "" {
		out["team_id"] = s.TeamID
	}
	if s.ICalOverridesURL != "" {
		out["ical_url_overrides"] = s.ICalOverridesURL
	}
	if s.TimeZone != nil {
		out["time_zone"] = s.TimeZone.String()
	}
	out["slack"] = s.Slack

	return json.Marshal(&out)
}

func (s *Schedule) UnmarshalJSON(b []byte) error {
	rawHeaders := scheduleRawHeaders{}
	err := json.Unmarshal(b, &rawHeaders)
	if err != nil {
		return err
	}

	*s = Schedule{
		ID:               rawHeaders.ID,
		Name:             rawHeaders.Name,
		TeamID:           rawHeaders.TeamID,
		ICalOverridesURL: rawHeaders.ICalURLOverrides,
		Slack:            rawHeaders.Slack,
	}

	s.TimeZone, err = time.LoadLocation(rawHeaders.TimeZone)
	if err != nil {
		return err
	}

	cal, knownType := scheduleCalendarTypeLookup[strings.ToLower(rawHeaders.Type)]
	if knownType {
		err = json.Unmarshal(b, &cal)
		if err != nil {
			return err
		}
	}

	s.Calendar = cal
	return nil
}

type ScheduleSlackMetadata struct {
	ChannelID   string `json:"channel_id"`
	UserGroupID string `json:"user_group_id"`
}

type ScheduleCalendarType int

const (
	ScheduleCalendarTypeUnknown ScheduleCalendarType = iota
	ScheduleCalendarTypeWeb
	ScheduleCalendarTypeICal
	scheduleCalendarTypeLen
)

var scheduleCalendarTypeStringLookup = [scheduleCalendarTypeLen]string{
	"unknown",
	"calendar",
	"ical",
}

var scheduleCalendarTypeLookup = map[string]ScheduleCalendar{
	"calendar": ScheduleCalendarWeb{},
	"ical":     ScheduleCalendarICal{},
}

func (s ScheduleCalendarType) String() string {
	if s < 0 || s >= scheduleCalendarTypeLen {
		return "unknown"
	}

	return scheduleCalendarTypeStringLookup[s]
}

type ScheduleCalendar interface {
	ScheduleCalendarType() ScheduleCalendarType
}

type ScheduleCalendarWeb struct {
	Shifts []string `json:"shifts"`
}

func (s ScheduleCalendarWeb) ScheduleCalendarType() ScheduleCalendarType {
	return ScheduleCalendarTypeWeb
}

type ScheduleCalendarICal struct {
	PrimaryURL string `json:"ical_url_primary"`
}

func (s ScheduleCalendarICal) ScheduleCalendarType() ScheduleCalendarType {
	return ScheduleCalendarTypeICal
}

type CreateScheduleOptions struct {
	TeamID           string
	Slack            ScheduleSlackMetadata
	ICalOverridesURL string
	//TimeZone defaults to UTC
	TimeZone *time.Location
}

func (c *Client) CreateSchedule(
	name string,
	cal ScheduleCalendar,
	opts *CreateScheduleOptions,
) (*Schedule, error) {

	ret := &Schedule{}

	nonNilOpts := CreateScheduleOptions{}
	if opts != nil {
		nonNilOpts = *opts
	}

	if nonNilOpts.TimeZone == nil {
		nonNilOpts.TimeZone = time.UTC
	}

	schedOut := &Schedule{
		Name:             name,
		Calendar:         cal,
		TeamID:           nonNilOpts.TeamID,
		Slack:            nonNilOpts.Slack,
		ICalOverridesURL: nonNilOpts.ICalOverridesURL,
		TimeZone:         nonNilOpts.TimeZone,
	}

	err := c.doRequest(
		"POST",
		schedulePath,
		schedOut,
		ret,
	)
	return ret, err
}

type ScheduleFilter struct {
	Name string
}

func (c *Client) ListSchedulesByPage(
	page int,
	filter *ScheduleFilter,
) (*PaginatedResponse[Schedule], error) {

	values := url.Values{}
	if filter != nil {
		if filter.Name != "" {
			values.Set("name", filter.Name)
		}
	}

	return getPage[Schedule](c, page, schedulePath, values)
}

func (c *Client) ListSchedules(
	filter *ScheduleFilter,
) ([]Schedule, error) {
	return paginate(c.ListSchedulesByPage, filter)
}

func (c *Client) GetSchedule(id string) (*Schedule, error) {
	ret := &Schedule{}
	err := c.doRequest("GET", buildPath(schedulePath, id), nil, ret)
	return ret, err
}

func (c *Client) DeleteSchedule(id string) error {
	return c.doRequest("DELETE", buildPath(schedulePath, id), nil, nil)
}

//TODO: Write UpdateSchedule
