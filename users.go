package oncall

import "net/url"

const userPath = "users"

type User struct {
	ID       string            `json:"id"`
	Email    string            `json:"email"`
	Username string            `json:"username"`
	Slack    UserSlackMetadata `json:"slack"`
	Role     UserRole          `json:"role"`
}

type UserSlackMetadata struct {
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
}

type UserRole string

const (
	UserRoleUnknown  UserRole = "unknown"
	UserRoleUser     UserRole = "user"
	UserRoleObserver UserRole = "observer"
	UserRoleAdmin    UserRole = "admin"
)

type UserFilter struct {
	Username string
}

func (c *Client) ListUsersByPage(
	page int,
	filter *UserFilter,
) (*PaginatedResponse[User], error) {

	values := url.Values{}
	if filter != nil {
		if filter.Username != "" {
			values.Set("username", filter.Username)
		}
	}

	return getPage[User](c, page, userPath, values)
}

func (c *Client) ListUsers(
	filter *UserFilter,
) ([]User, error) {
	return paginate(c.ListUsersByPage, filter)
}

func (c *Client) GetUser(id string) (*User, error) {
	ret := &User{}
	err := c.doRequest("GET", buildPath(userPath, id), nil, ret)
	return ret, err
}
