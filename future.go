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
	OnComplete(callback func(_ interface{}, _ error))
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
	completed   chan interface{}
	successCBs  []func(_ interface{})
	errorCBs    []func(_ error)
	completeCBs []func(_ interface{}, _ error)
	value       interface{}
	err         error
}

func NewPromise() Promise {
	p := &promiseImpl{
		completed:   make(chan interface{}),
		successCBs:  make([]func(_ interface{}), 0),
		errorCBs:    make([]func(_ error), 0),
		completeCBs: make([]func(_ interface{}, _ error), 0),
	}
	return p
}

func Do(asyncFunc func() (interface{}, error)) Future {
	p := NewPromise()

	go func() {
		val, err := asyncFunc()
		if err != nil {
			p.Error(err)
		} else {
			p.Success(val)
		}
	}()

	return p
}

// OnSuccess invokes the statusReceiver function as soon as the future is successfully completed
func (f *promiseImpl) OnSuccess(callback func(_ interface{})) {
	if f.IsCompleted() {
		go callback(f.value)
		return
	}
	f.successCBs = append(f.successCBs, callback)
}

func (f *promiseImpl) OnError(callback func(_ error)) {
	if f.IsCompleted() {
		go callback(f.err)
		return
	}
	f.errorCBs = append(f.errorCBs, callback)
}

func (f *promiseImpl) OnComplete(callback func(_ interface{}, _ error)) {
	if f.IsCompleted() {
		go callback(f.value, f.err)
		return
	}
	f.completeCBs = append(f.completeCBs, callback)
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
	for _, rCallback := range f.completeCBs {
		go rCallback(f.value, nil)
	}
	// callbacks arrays are not needed anymore. Removing
	f.successCBs = nil
	f.completeCBs = nil
	f.errorCBs = nil
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
	for _, rCallback := range f.completeCBs {
		go rCallback(nil, f.err)
	}
	// callbacks arrays are not needed anymore. Removing
	f.successCBs = nil
	f.completeCBs = nil
	f.errorCBs = nil
	return nil
}

// Get should coexist and close onsuccess
func (f *promiseImpl) Get() (interface{}, error) {
	// Wait for completion
	<-f.completed
	return f.value, f.err
}

func (f *promiseImpl) Eventually(timeout time.Duration) (interface{}, error) {
	gotValues := make(chan interface{})
	defer close(gotValues)
	var val interface{}
	var err error
	go func() {
		val, err = f.Get()
		gotValues <- struct{}{}
	}()

	select {
	case <-gotValues:
		return val, err
	case <-time.After(timeout):
		return nil, fmt.Errorf("future has not completed after %v", timeout)
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
