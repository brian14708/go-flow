package rrworker

var (
	_ TaskInterface = (*Task)(nil)

	invalidCb = func(interface{}, error) {
		panic("cannot set value twice")
	}
)

type TaskInterface interface {
	// SetResult(T)
	SetResultAny(interface{})
	SetError(error)
	setCallback(ResultCallback)
}

type ResultCallback func(interface{}, error)

type Task ResultCallback

func (p *Task) setCallback(cb ResultCallback) {
	if *p != nil {
		panic("each task can only set promise once")
	}
	*p = Task(cb)
}

func (p *Task) SetResultAny(v interface{}) {
	(*p)(v, nil)
	*p = invalidCb
}

func (p *Task) SetError(err error) {
	if err == nil {
		panic("SetError must provide valid error")
	}
	(*p)(nil, err)
	*p = invalidCb
}
