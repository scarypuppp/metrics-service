package audit

// Subscriber is an audit event consumer that can be stopped and waited on.
type Subscriber interface {
	Wait()
	Stop()
}
