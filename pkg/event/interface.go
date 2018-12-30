// Subscription represents a stream of events. The carrier of the events is typically a
// channel, but isn't part of the interface.
//
// Subscriptions can fail while established. Failures are reported through an error
// channel. It receives a value if there is an issue with the subscription (e.g. the
// network connection delivering the events has been closed). Only one value will ever be
// sent.
//
// The error channel is closed when the subscription ends successfully (i.e. when the
// source of events is closed). It is also closed when Unsubscribe is called.
//
// The Unsubscribe method cancels the sending of events. You must call Unsubscribe in all
// cases to ensure that resources related to the subscription are released. It can be
// called any number of times.

package event

type Subscription interface {
	Err() <-chan error // returns the error channel
	Unsubscribe()      // cancels sending of events, closing the error channel
}
