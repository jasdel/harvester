package queue

import (
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/common"
)

// Client for communicating with th eNATS message queue. The publishers
// will publish to the queue asynchronously.  The receiver will block
// until a message has been received on the queue.  Receiver endpoints
// are also configured as Queue Receivers.  Therefore if there are
// multiple receivers on the same topic only a single one will receive
// the message. Messages will be received in the order they were sent.
type client struct {
	// Client for communicating with the NATS message queue. Transmission
	// with the message queue will be pre-(d)encoded. So extra processing
	// is not needed
	ec *nats.EncodedConn

	// Sending channel to publish to a queue. Only initialized by
	// newClient if the sender flag is set.
	sendCh chan *common.URLQueueItem

	// Receiving channel to receive from a queue. Only initialized
	// by newClient if the receiver flag is set.
	recvCh chan *common.URLQueueItem
}

// Interface for publishing to an URLQueueItem topic
type Publisher interface {
	// Closes the Publish channel. No more calls to Send should be made
	// once Close is called. The Publisher's close should be called
	// when finished with the topic or it will leak.
	Close()

	// Sends one or multiple URL items to associated topic's receivers
	Send(item ...*common.URLQueueItem)
}

// Interface for receiving from an URLQueueITem topic
type Receiver interface {
	// Closes the Receiver channel. No more calls to Send should be made
	// once Close is called. The Receiver's close should be called
	// when finished with the topic or it will leak.
	Close()

	// Receive channel to receive items from the associated topic
	Receive() <-chan *common.URLQueueItem
}

// Creates a new Queue Publisher which is only able to send
// to to topic provided. The Close of a
func NewPublisher(cfg QueueConfig) (Publisher, error) {
	return newClient(cfg, true, false)
}

// Creates a new Queue Receiver which is only able to receive from
// the topic provided.
func NewReceiver(cfg QueueConfig) (Receiver, error) {
	return newClient(cfg, false, true)
}

// Creates a new Queue Client. The client can be configured as a sender,
// receiver, or both for the topic provided.
func newClient(cfg QueueConfig, sender, receiver bool) (*client, error) {
	c := &client{}

	nc, err := nats.Connect(cfg.ConnURL)
	if err != nil {
		return nil, err
	}

	c.ec, err = nats.NewEncodedConn(nc, "json")
	if err != nil {
		return nil, err
	}

	if sender {
		c.sendCh = make(chan *common.URLQueueItem)
		c.ec.BindSendChan(cfg.Topic, c.sendCh)
	}

	if receiver {
		c.recvCh = make(chan *common.URLQueueItem)
		c.ec.BindRecvQueueChan(cfg.Topic, cfg.Topic, c.recvCh)
	}

	return c, nil
}

// Closes the Queue. No more attempts send or receive should be made
// once the clients queue connection is closed.
func (c *client) Close() {
	c.ec.Close()
}

// Adds a new URLQueueItem to the queue.  A Single or multiple
// items can be added at once, and they will be sent to the queue
// in order.
func (c *client) Send(items ...*common.URLQueueItem) {
	for i := 0; i < len(items); i++ {
		c.sendCh <- items[i]
	}
}

// Returns a read only channel to send URLQueueItem to
func (c *client) Receive() <-chan *common.URLQueueItem {
	return c.recvCh
}

// Queue configuration states what topic the queue channel should be
// attached to and the connection URL for the messaging service.
type QueueConfig struct {
	// Topic to connect to. In the case of a receiver queue client
	// the topic will also be the queue channel.
	Topic string `json:"topic"`

	// Connection URL to the messaging service.
	ConnURL string `json:"connURL"`
}
