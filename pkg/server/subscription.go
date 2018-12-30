package server

import (
	"context"
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"time"

	"airman.com/airfk/pkg/codec"
)

var (
	// ErrNotificationsUnsupported is returned when the connection doesn't support notifications
	ErrNotificationsUnsupported = errors.New("notifications not supported")
	// ErrNotificationNotFound is returned when the notification for the given id is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

// ID defines a pseudo random number that is used to identify RPC subscriptions.
type ID string

// a Subscription is created by a notifier and tight to that notifier. The client can use
// this subscription to wait for an unsubscribe request for the client, see Err().
type Subscription struct {
	ID        ID
	namespace string
	err       chan error // closed on unsubscribe
}

// Err returns a channel that is closed when the client send an unsubscribe request.
func (s *Subscription) Err() <-chan error {
	return s.err
}

// notifierKey is used to store a notifier within the connection context.
type notifierKey struct{}

// Notifier is tight to a RPC connection that supports subscriptions.
// Server callbacks use the notifier to send notifications.
type Notifier struct {
	codec    codec.ServerCodec
	subMu    sync.RWMutex // guards active and inactive maps
	active   map[ID]*Subscription
	inactive map[ID]*Subscription
}

// newNotifier creates a new notifier that can be used to send subscription
// notifications to the client.
func newNotifier(codec codec.ServerCodec) *Notifier {
	return &Notifier{
		codec:    codec,
		active:   make(map[ID]*Subscription),
		inactive: make(map[ID]*Subscription),
	}
}

// NotifierFromContext returns the Notifier value stored in ctx, if any.
func NotifierFromContext(ctx context.Context) (*Notifier, bool) {
	n, ok := ctx.Value(notifierKey{}).(*Notifier)
	return n, ok
}

// CreateSubscription returns a new subscription that is coupled to the
// RPC connection. By default subscriptions are inactive and notifications
// are dropped until the subscription is marked as active. This is done
// by the RPC server after the subscription ID is send to the client.
func (n *Notifier) CreateSubscription() *Subscription {
	s := &Subscription{ID: NewID(), err: make(chan error)}
	n.subMu.Lock()
	n.inactive[s.ID] = s
	n.subMu.Unlock()
	return s
}

// Notify sends a notification to the client with the given data as payload.
// If an error occurs the RPC connection is closed and the error is returned.
func (n *Notifier) Notify(id ID, data interface{}) error {
	n.subMu.RLock()
	defer n.subMu.RUnlock()

	sub, active := n.active[id]
	if active {
		notification := n.codec.CreateNotification(string(id), sub.namespace, data)
		if err := n.codec.Write(notification); err != nil {
			n.codec.Close()
			return err
		}
	}
	return nil
}

// Closed returns a channel that is closed when the RPC connection is closed.
func (n *Notifier) Closed() <-chan interface{} {
	return n.codec.Closed()
}

// unsubscribe a subscription.
// If the subscription could not be found ErrSubscriptionNotFound is returned.
func (n *Notifier) unsubscribe(id ID) error {
	n.subMu.Lock()
	defer n.subMu.Unlock()
	if s, found := n.active[id]; found {
		close(s.err)
		delete(n.active, id)
		return nil
	}
	return ErrSubscriptionNotFound
}

// activate enables a subscription. Until a subscription is enabled all
// notifications are dropped. This method is called by the RPC server after
// the subscription ID was sent to client. This prevents notifications being
// send to the client before the subscription ID is send to the client.
func (n *Notifier) activate(id ID, namespace string) {
	n.subMu.Lock()
	defer n.subMu.Unlock()
	if sub, found := n.inactive[id]; found {
		sub.namespace = namespace
		n.active[id] = sub
		delete(n.inactive, id)
	}
}

// NewID generates a identifier that can be used as an identifier in the RPC interface.
// e.g. filter and subscription identifier.
func NewID() ID {

	id := make([]byte, 16)
	for i := 0; i < len(id); i += 7 {
		val := idGenerator().Int63()
		for j := 0; i+j < len(id) && j < 7; j++ {
			id[i+j] = byte(val)
			val >>= 8
		}
	}

	rpcId := hex.EncodeToString(id)
	// rpc ID's are RPC quantities, no leading zero's and 0 is 0x0
	rpcId = strings.TrimLeft(rpcId, "0")
	if rpcId == "" {
		rpcId = "0"
	}

	return ID("0x" + rpcId)
}

// idGenerator helper utility that generates a (pseudo) random sequence of
// bytes that are used to generate identifiers.
func idGenerator() *rand.Rand {
	return rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
}
