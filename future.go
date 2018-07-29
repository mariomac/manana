package manana

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrorCanceled is an error returned when trying to operate with an already canceled Future
var ErrorCanceled = errors.New("this future has been already canceled")
// ErrorComplete is an error returned when trying to complete and already completed Future
var ErrorCompleted = errors.New("this promise has been already completed")

// Future holds the results of an operation that runs asynchronously, in background.
type Future interface {
	OnSuccess(callback func(_ interface{}))
	OnFail(callback func(_ error))
	OnComplete(callback func(_ interface{}, _ error))
	IsCompleted() bool
	Get() (interface{}, error)
	Eventually(timeout time.Duration) (interface{}, error)
	Cancel() error
}

type Promise interface {
	Future
	Success(value interface{}) error
	Error(err error) error
}

type promiseImpl struct {
	completed   chan interface{}
	context     context.Context
	cancel      context.CancelFunc
	successCBs  []func(_ interface{})
	errorCBs    []func(_ error)
	completeCBs []func(_ interface{}, _ error)
	value       interface{} // Todo: use atomic
	err         error       // todo: use atomic
}

func NewPromise() Promise {
	ctx, cancelFunc := context.WithCancel(context.Background())
	p := &promiseImpl{
		completed:   make(chan interface{}),
		context:     ctx,
		cancel:      cancelFunc,
		successCBs:  make([]func(_ interface{}), 0),
		errorCBs:    make([]func(_ error), 0),
		completeCBs: make([]func(_ interface{}, _ error), 0),
	}
	return p
}

func Do(asyncFunc func() (interface{}, error)) Future {
	p := NewPromise().(*promiseImpl)
	go func() {
		valCh := make(chan interface{})
		errCh := make(chan error)
		go func() {
			val, err := asyncFunc()
			if err != nil {
				errCh <- err
			} else {
				valCh <- val
			}
		}()
		select {
		case err := <-errCh:
			p.Error(err)
		case val := <-valCh:
			p.Success(val)
		case <-p.context.Done():
			// do nothing
		}
	}()
	return p
}

func DoCtx(asyncFunc func(ctx context.Context) (interface{}, error)) Future {
	p := NewPromise().(*promiseImpl)
	go func() {
		valCh := make(chan interface{})
		errCh := make(chan error)
		go func() {
			val, err := asyncFunc(p.context)
			if err != nil {
				errCh <- err
			} else {
				valCh <- val
			}
		}()
		select {
		case err := <-errCh:
			p.Error(err)
		case val := <-valCh:
			p.Success(val)
		case <-p.context.Done():
			// do nothing
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
	if !f.IsCanceled() {
		f.successCBs = append(f.successCBs, callback)
	}
}

func (f *promiseImpl) OnFail(callback func(_ error)) {
	if f.IsCompleted() {
		go callback(f.err)
		return
	}
	if !f.IsCanceled() {
		f.errorCBs = append(f.errorCBs, callback)
	}
}

func (f *promiseImpl) OnComplete(callback func(_ interface{}, _ error)) {
	if f.IsCompleted() {
		go callback(f.value, f.err)
		return
	}
	if !f.IsCanceled() {
		f.completeCBs = append(f.completeCBs, callback)
	}
}

func (f *promiseImpl) Success(value interface{}) error {
	if f.IsCompleted() {
		return ErrorCompleted
	}
	if f.IsCanceled() {
		return ErrorCanceled
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

func (f *promiseImpl) Error(err error) error {
	if f.IsCompleted() {
		return ErrorCompleted
	}
	if f.IsCanceled() {
		return ErrorCanceled
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
	select {
	case <-f.completed:
		return f.value, f.err
	case <-f.context.Done():
		return nil, ErrorCanceled
	}
}

func (f *promiseImpl) Eventually(timeout time.Duration) (interface{}, error) {
	if f.IsCanceled() {
		return nil, ErrorCanceled
	}
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
	case <-f.context.Done():
		return nil, ErrorCanceled
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

func (f *promiseImpl) IsCanceled() bool {
	select {
	case <-f.context.Done():
		return true
	default:
		return false
	}
}

func (f *promiseImpl) Cancel() error {
	if f.IsCanceled() {
		return ErrorCanceled
	}
	f.cancel()
	return nil
}
