package event

type Dispatcher interface {
	Dispatch(event Event) error
}
