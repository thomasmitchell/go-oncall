package oncall

import (
	"net/url"
)

const escChainPath = "escalation_chains"

type EscalationChain struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	TeamID string `json:"team_id"`
}

type ListEscalationChainsFilter struct {
	//This is empty, but here for backwards-compatibility if they implement
	//filters in the future
}

func (c *Client) ListEscalationChainsByPage(page int, filter *ListEscalationChainsFilter) (*PaginatedResponse[EscalationChain], error) {
	return getPage[EscalationChain](c, page, escChainPath, url.Values{})
}

func (c *Client) ListEscalationChains(filter *ListEscalationChainsFilter) ([]EscalationChain, error) {
	return paginate(c.ListEscalationChainsByPage, filter)
}

func (c *Client) GetEscalationChain(id string) (*EscalationChain, error) {
	ret := &EscalationChain{}
	err := c.doRequest("GET", buildPath(escChainPath, id), nil, ret)
	return ret, err
}

type CreateEscalationChainOptions struct {
	TeamID string
}

func (c *Client) CreateEscalationChain(
	name string,
	opts *CreateEscalationChainOptions,
) (*EscalationChain, error) {

	requestBody := struct {
		Name   string `json:"name"`
		TeamID string `json:"team_id,omitempty"`
	}{
		Name: name,
	}

	if opts != nil {
		if opts.TeamID != "" {
			requestBody.TeamID = opts.TeamID
		}
	}

	ret := &EscalationChain{}

	err := c.doRequest("POST", "escalation_chains", &requestBody, ret)
	return ret, err
}

func (c *Client) DeleteEscalationChain(id string) error {
	return c.doRequest("DELETE", buildPath(escChainPath, id), nil, nil)
}
