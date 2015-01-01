package queue

import (
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/types"
)

type client struct {
	ec     *nats.EncodedConn
	sendCh chan *types.URLQueueItem
	recvCh chan *types.URLQueueItem
}

type Publisher interface {
	Close()
	Send(item *types.URLQueueItem)
}

type Receiver interface {
	Close()
	Receive() <-chan *types.URLQueueItem
}

func NewPublisher(connURL, topic string) (Publisher, error) {
	return newClient(connURL, topic, true, false)
}

func NewReceiver(connURL, topic string) (Receiver, error) {
	return newClient(connURL, topic, false, true)
}

func newClient(connURL, topic string, sender, receiver bool) (*client, error) {
	c := &client{}

	nc, err := nats.Connect(connURL)
	if err != nil {
		return nil, err
	}

	c.ec, err = nats.NewEncodedConn(nc, "json")
	if err != nil {
		return nil, err
	}

	if sender {
		c.sendCh = make(chan *types.URLQueueItem)
		c.ec.BindSendChan(topic, c.sendCh)
	}

	if receiver {
		c.recvCh = make(chan *types.URLQueueItem)
		c.ec.BindRecvQueueChan(topic, topic, c.recvCh)
	}

	return c, nil
}

func (c *client) Close() {
	c.ec.Close()
}

func (c *client) Send(item *types.URLQueueItem) {
	c.sendCh <- item
}

func (c *client) Receive() <-chan *types.URLQueueItem {
	return c.recvCh
}
