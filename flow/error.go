package flow

import "fmt"

type GraphError struct {
	name string
	err  error
	next *GraphError
}

func (e *GraphError) Error() string {
	if e.name == "" {
		return fmt.Sprintf("graph error: %s", e.err)
	}
	return fmt.Sprintf("node `%s' failed: %s", e.name, e.err)
}

func (e *GraphError) Unwrap() error {
	return e.err
}

func (e *GraphError) NodeName() string {
	return e.name
}

func (e *GraphError) Next() *GraphError {
	return e.next
}
