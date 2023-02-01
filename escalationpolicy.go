package oncall

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"
)

const escPolicyPath = "escalation_policies"

type EscalationPolicy struct {
	ID                string
	EscalationChainID string
	Position          int
	Rule              EscalationPolicyRule
}

type escalationPolicyRawHeaders struct {
	ID                string `json:"id"`
	EscalationChainID string `json:"escalation_chain_id"`
	Position          int    `json:"position"`
	Type              string `json:"type"`
}

func (e *EscalationPolicy) MarshalJSON() ([]byte, error) {
	if e == nil {
		return []byte("null"), nil
	}

	out := map[string]interface{}{}
	if e.Rule != nil {
		inter, err := json.Marshal(&e.Rule)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(inter, &out)
		if err != nil {
			return nil, err
		}

		out["type"] = e.Rule.EscalationPolicyType().String()
	}

	if e.ID != "" {
		out["id"] = e.ID
	}
	out["escalation_chain_id"] = e.EscalationChainID
	out["position"] = e.Position

	return json.Marshal(&out)
}

func (e *EscalationPolicy) UnmarshalJSON(b []byte) error {
	rawHeaders := escalationPolicyRawHeaders{}
	err := json.Unmarshal(b, &rawHeaders)
	if err != nil {
		return err
	}

	*e = EscalationPolicy{
		ID:                rawHeaders.ID,
		EscalationChainID: rawHeaders.EscalationChainID,
		Position:          rawHeaders.Position,
	}

	rule := escalationPolicyRuleFromString(rawHeaders.Type)
	if rule != nil {
		err = json.Unmarshal(b, rule)
		if err != nil {
			return err
		}
	}

	e.Rule = rule
	return nil
}

const (
	EscalationPolicyPositionStart = 0
	EscalationPolicyPositionEnd   = -1
)

type EscalationPolicyType int

func (e EscalationPolicyType) String() string {
	if e < 0 || e >= escalationPolicyTypeLen {
		return "unknown"
	}

	return escalationPolicyTypeStringLookup[e]
}

const (
	EscalationPolicyTypeUnknown EscalationPolicyType = iota
	EscalationPolicyTypeWait
	EscalationPolicyTypeNotifyPersons
	EscalationPolicyTypeNotifyPersonNextEachTime
	EscalationPolicyTypeNotifyOnCallFromSchedule
	EscalationPolicyTypeNotifyUserGroup
	EscalationPolicyTypeTriggerAction
	EscalationPolicyTypeResolve
	EscalationPolicyTypeNotifyWholeChannel
	EscalationPolicyTypeNotifyIfTimeFromTo
	escalationPolicyTypeLen
)

var escalationPolicyTypeStringLookup = [escalationPolicyTypeLen]string{
	"unknown",
	"wait",
	"notify_persons",
	"notify_person_next_each_time",
	"notify_on_call_from_schedule",
	"notify_user_group",
	"trigger_action",
	"resolve",
	"notify_whole_channel",
	"notify_if_time_from_to",
}

func escalationPolicyRuleFromString(s string) EscalationPolicyRule {
	switch strings.ToLower(s) {
	case "wait":
		return &EscalationPolicyRuleWait{}
	case "notify_persons":
		return &EscalationPolicyRuleNotifyPersons{}
	case "notify_person_next_each_time":
		return &EscalationPolicyRuleNotifyPersonNextEachTime{}
	case "notify_on_call_from_schedule":
		return &EscalationPolicyRuleNotifyOnCallFromSchedule{}
	case "notify_user_group":
		return &EscalationPolicyRuleNotifyUserGroup{}
	case "trigger_action":
		return &EscalationPolicyRuleTriggerAction{}
	case "resolve":
		return &EscalationPolicyRuleResolve{}
	case "notify_whole_channel":
		return &EscalationPolicyRuleNotifyWholeChannel{}
	case "notify_if_time_from_to":
		return &EscalationPolicyRuleNotifyIfTimeFromTo{}
	}

	return nil
}

type EscalationPolicyRule interface {
	EscalationPolicyType() EscalationPolicyType
}

type EscalationPolicyRuleWait struct {
	Duration time.Duration `json:"duration"`
}

func (e *EscalationPolicyRuleWait) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeWait
}

type EscalationPolicyRuleNotifyPersons struct {
	Important bool     `json:"important"`
	UserIDs   []string `json:"persons_to_notify"`
}

func (e *EscalationPolicyRuleNotifyPersons) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeNotifyPersons
}

type EscalationPolicyRuleNotifyPersonNextEachTime struct {
	UserIDs []string `json:"persons_to_notify_next_each_time"`
}

func (e *EscalationPolicyRuleNotifyPersonNextEachTime) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeNotifyPersonNextEachTime
}

type EscalationPolicyRuleNotifyOnCallFromSchedule struct {
	Important  bool   `json:"important"`
	ScheduleID string `json:"notify_on_call_from_schedule"`
}

func (e *EscalationPolicyRuleNotifyOnCallFromSchedule) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeNotifyOnCallFromSchedule
}

type EscalationPolicyRuleNotifyUserGroup struct {
	Important   bool   `json:"important"`
	UserGroupID string `json:"group_to_notify"`
}

func (e *EscalationPolicyRuleNotifyUserGroup) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeNotifyUserGroup
}

type EscalationPolicyRuleTriggerAction struct {
	ActionID string `json:"action_to_trigger"`
}

func (e *EscalationPolicyRuleTriggerAction) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeTriggerAction
}

type EscalationPolicyRuleResolve struct{}

func (e *EscalationPolicyRuleResolve) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeResolve
}

type EscalationPolicyRuleNotifyWholeChannel struct{}

func (e *EscalationPolicyRuleNotifyWholeChannel) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeNotifyWholeChannel
}

type EscalationPolicyRuleNotifyIfTimeFromTo struct {
	//only hours, minutes, and seconds will be used
	From time.Time
	//only hours, minutes, and seconds will be used
	To time.Time
}

type escalationPolicyRuleNotifyIfTimeFromToRaw struct {
	From string `json:"notify_if_time_from"`
	To   string `json:"notify_if_time_to"`
}

func (e *EscalationPolicyRuleNotifyIfTimeFromTo) EscalationPolicyType() EscalationPolicyType {
	return EscalationPolicyTypeNotifyIfTimeFromTo
}

func (e *EscalationPolicyRuleNotifyIfTimeFromTo) MarshalJSON() ([]byte, error) {
	return json.Marshal(&escalationPolicyRuleNotifyIfTimeFromToRaw{
		From: timeOfDayToString(e.From),
		To:   timeOfDayToString(e.To),
	})
}

func (e *EscalationPolicyRuleNotifyIfTimeFromTo) UnmarshalJSON(b []byte) error {
	raw := escalationPolicyRuleNotifyIfTimeFromToRaw{}
	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	e.From, err = timeFromString(raw.From)
	if err != nil {
		return err
	}

	e.To, err = timeFromString(raw.To)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CreateEscalationPolicy(
	escChainID string,
	position int,
	rule EscalationPolicyRule,
) (*EscalationPolicy, error) {

	ret := &EscalationPolicy{}
	err := c.doRequest(
		"POST",
		escPolicyPath,
		&EscalationPolicy{
			EscalationChainID: escChainID,
			Position:          position,
			Rule:              rule,
		},
		ret,
	)
	return ret, err
}

type EscalationPolicyFilter struct {
	EscalationChainID string
}

func (c *Client) ListEscalationPoliciesByPage(
	page int,
	filter *EscalationPolicyFilter,
) (*PaginatedResponse[EscalationPolicy], error) {

	values := url.Values{}
	if filter != nil {
		if filter.EscalationChainID != "" {
			values.Set("escalation_chain_id", filter.EscalationChainID)
		}
	}

	return getPage[EscalationPolicy](c, page, escPolicyPath, values)
}

func (c *Client) ListEscalationPolicies(
	filter *EscalationPolicyFilter,
) ([]EscalationPolicy, error) {
	return paginate(c.ListEscalationPoliciesByPage, filter)
}

func (c *Client) GetEscalationPolicy(id string) (*EscalationPolicy, error) {
	ret := &EscalationPolicy{}
	err := c.doRequest("GET", buildPath(escPolicyPath, id), nil, ret)
	return ret, err
}

func (c *Client) DeleteEscalationPolicy(id string) error {
	return c.doRequest("DELETE", buildPath(escPolicyPath, id), nil, nil)
}
