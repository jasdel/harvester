package storage

type URL struct {
	url    string
	client *Client
}

func (u *URL) AddDecendants(urls []string) error {
	return nil
}
