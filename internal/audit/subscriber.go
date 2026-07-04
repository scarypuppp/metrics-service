package audit

type Subscriber interface {
	Wait()
	Stop()
}
