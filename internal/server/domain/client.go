package domain

type Client struct {
	id   string
	name string
}

func (c *Client) Id() string {
	return c.id
}

// TODO: later check if this is really needed along with the name field
func (c *Client) Name() string {
	return c.name
}

func NewClient(id string, name string) *Client {
	return &Client{
		id,
		name,
	}
}
