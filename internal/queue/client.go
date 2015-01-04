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

// Interface for publishing to an URLQueueItem topic
type Publisher interface {
	// Closes the Publish channel. No more calls to Send should be made
	// once Close is called. The Publisher's close should be called
	// when finished with the topic or it will leak.
	Close()

	// Sends one or multiple URL items to associated topic's receivers
	Send(item ...*types.URLQueueItem)
}

// Interface for receiving from an URLQueueITem topic
type Receiver interface {
	// Closes the Receiver channel. No more calls to Send should be made
	// once Close is called. The Receiver's close should be called
	// when finished with the topic or it will leak.
	Close()

	// Receive channel to receive items from the associated topic
	Receive() <-chan *types.URLQueueItem
}

// Creates a new Queue Publisher which is only able to send
// to to topic provided. The Close of a
func NewPublisher(connURL, topic string) (Publisher, error) {
	return newClient(connURL, topic, true, false)
}

// Creates a new Queue Receiver which is only able to receive from
// the topic provided.
func NewReceiver(connURL, topic string) (Receiver, error) {
	return newClient(connURL, topic, false, true)
}

// Creates a new Queue Client. The client can be configured as a sender,
// receiver, or both for the topic provided.
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

// Closes the Queue. No more attempts send or receive should be made
// once the clients queue connection is closed.
func (c *client) Close() {
	c.ec.Close()
}

// Adds a new URLQueueItem to the queue
func (c *client) Send(items ...*types.URLQueueItem) {
	for i := 0; i < len(items); i++ {
		c.sendCh <- items[i]
	}
}

// Returns a read only channel to send URLQueueItem to
func (c *client) Receive() <-chan *types.URLQueueItem {
	return c.recvCh
}
