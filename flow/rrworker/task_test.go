package rrworker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTaskCallback(t *testing.T) {
	task := new(Task)
	task.setCallback(func(interface{}, error) {})
	assert.Panics(t, func() {
		task.setCallback(func(interface{}, error) {})
	})

	task = new(Task)
	var cnt int
	task.setCallback(func(interface{}, error) {
		cnt++
	})
	assert.Panics(t, func() {
		task.SetError(nil)
	})
	task.SetResultAny(nil)
	assert.Equal(t, 1, cnt)
	assert.Panics(t, func() {
		task.SetError(errors.New(""))
	})
}
