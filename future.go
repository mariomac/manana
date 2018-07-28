package manana

import (
	"errors"
	"fmt"
	"time"
)

var errorCompleted = errors.New("this promise has been already completed")

// promiseImpl here should be probably named "Promise", since has a completable status

type Future interface {
	OnSuccess(callback func(_ interface{}))
	OnError(callback func(_ error))
	IsCompleted() bool
	Get() (interface{}, error)
	Eventually(timeout time.Duration) (interface{}, error)
}

type Promise interface {
	Future
	Success(value interface{}) error
	Error(err error) error
}

type promiseImpl struct {
	completed  chan interface{}
	successCBs []func(_ interface{})
	errorCBs   []func(_ error)
	value      interface{}
	err        error
}

func NewPromise() Promise {
	p := &promiseImpl{
		completed:  make(chan interface{}),
		successCBs: make([]func(_ interface{}), 0),
		errorCBs:   make([]func(_ error), 0),
	}
	return p
}

// OnSuccess invokes the statusReceiver function as soon as the future is successfully completed
func (f *promiseImpl) OnSuccess(callback func(_ interface{})) {
	if f.value != nil {
		go callback(f.value)
		return
	}
	f.successCBs = append(f.successCBs, callback)
}

func (f *promiseImpl) OnError(callback func(_ error)) {
	if f.err != nil {
		go callback(f.err)
		return
	}
	f.errorCBs = append(f.errorCBs, callback)
}

func (f *promiseImpl) Success(value interface{}) error {
	if f.IsCompleted() {
		return errorCompleted
	}
	f.value = value
	close(f.completed)
	for _, rCallback := range f.successCBs {
		go rCallback(f.value)
	}
	return nil
}

func (f *promiseImpl) Error(err error) error { // TODO: error should work as Success
	if f.IsCompleted() {
		return errorCompleted
	}
	f.err = err
	close(f.completed)
	for _, rCallback := range f.errorCBs {
		go rCallback(f.err)
	}
	return nil
}

// Todo: oncomplete (interface{}, error)

// Get should coexist and close onsuccess
func (f *promiseImpl) Get() (interface{}, error) {
	// Wait for completion
	<-f.completed
	return f.value, f.err
}

func (f *promiseImpl) Eventually(timeout time.Duration) (interface{}, error) {
	finishCh := make(chan interface{})
	errCh := make(chan error)
	defer close(finishCh)
	defer close(errCh)
	f.OnSuccess(func(val interface{}) {
		finishCh <- val
	})
	f.OnError(func(err error) {
		errCh <- err
	})
	select {
	case val := <-finishCh:
		return val, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("future did not completed after %v", timeout)
	}
}

func (f *promiseImpl) IsCompleted() bool {
	select {
	case <-f.completed:
		return true
	default:
		return false
	}
}
