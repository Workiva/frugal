package common

import "github.com/nats-io/nats"

func getNatsConn() *nats.Conn {
	addr := nats.DefaultURL
	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{addr}
	natsOptions.Secure = false
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	return conn
}
