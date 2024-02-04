type Client interface {
	Host string,
	Port number,
}

func (client *Client) Dial() err {
	fmt.Println("Dialing", Host, Port)
}

