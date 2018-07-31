// Package manana provides wrapper functionalities to work with futures and promises, and tools to
// easy porting your old synchronous code to asynchronous.
package manana

import (
	"context"
	"errors"
	"time"
)

// ErrorCanceled is an error returned when trying to operate with an already canceled Future
var ErrorCanceled = errors.New("this future is canceled")

// ErrorCompleted is an error returned when trying to complete and already completed Future
var ErrorCompleted = errors.New("this future is already completed")

// ErrorTimeout is an error returned when a timeout has been reached
var ErrorTimeout = errors.New("this operation has timed out")

// Future holds the results of an operation that runs asynchronously, in background.
type Future interface {
	// OnSuccess adds a callback to be run when the Future ends with a Success status. The callback
	// receives the value resulting from the successful operation.
	OnSuccess(callback func(_ interface{}))

	// OnFail adds a callback to be run when the Future fails with an error status, or when the
	// Future is canceled with the Cancel() function. The callback receives the error resulting
	// from the failed operation, or "ErrorCanceled" when the future is canceled.
	OnFail(callback func(_ error))

	// OnComplete adds a callback to be run indistinctly when the Future succeeds or fails (or
	// is cancelled). The callback may receive the value resulting from the successful operation
	// (first argument), or the error resulting from the failed operation (or ErrorCanceled,
	// second argument).
	OnComplete(callback func(_ interface{}, _ error))

	// Get makes the invoker goroutine to wait indefinitely until the Future completes, and may
	// return the value resulting from the successful execution of the Future (first value), or the
	// error resulting from the failed operation (or ErrorCanceled, second value).
	Get() (interface{}, error)

	// Eventually makes the invoker goroutine to wait until the Future completes, or until the
	// specified  timeout is triggered. It may return the value resulting from the successful
	// execution of the Future (first value), or the error resulting from the failed operation. The
	// second error may also (or ErrorCanceled, second value).
	Eventually(timeout time.Duration) (interface{}, error)

	// Cancel cancels the future. Canceling a future does not guarantee the goroutine it holds
	// can be immediately interrupted.
	Cancel() error

	// IsCompleted returns true if the future is finished, whatever is status is failed, success or
	// canceled.
	IsCompleted() bool

	// IsCanceled returns true if the future has been canceled, even if the held goroutine is still
	// being executed.
	IsCanceled() bool
}

// Promise is a Future whose Success/Fail status can be set.
type Promise interface {
	Future
	// Success completes the Promise with a success value passed as argument.
	Success(value interface{}) error
	// Fail completes the Promise with an error passed as argument.
	Fail(err error) error
	// CancelCh returns a channel that is closed when the work held in this Promise has to be
	// canceled.
	CancelCh() <-chan struct{}
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

// NewPromise creates a new, empty promise. This function is useful if you want to directly manage
// the status of a Promise from your code. To transparently wrap synchronous code into an
// asynchronous promise, you may use manana.Do and manana.DoCtx functions.
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

// Do wraps a synchronous function into an asynchronous envelope. The function is run in background
// and Do immediately returns a Future to get subscribed to the status of the function.
//
// If the wrapped function returns any value as first return value, the future succeeds.
// If the wrapped function returns an error as second return value, the future fails with the
// given error.
func Do(syncFunc func() (interface{}, error)) Future {
	p := NewPromise().(*promiseImpl)
	go func() {
		valCh := make(chan interface{})
		errCh := make(chan error)
		go func() {
			val, err := syncFunc()
			if err != nil {
				errCh <- err
			} else {
				valCh <- val
			}
		}()
		select {
		case err := <-errCh:
			p.Fail(err)
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
	if !f.IsCanceled() {
		if f.IsCompleted() {
			go callback(f.value)
			return
		}
		f.successCBs = append(f.successCBs, callback)
	}
}

func (f *promiseImpl) OnFail(callback func(_ error)) {
	if f.IsCompleted() {
		go callback(f.err)
		return
	}
	f.errorCBs = append(f.errorCBs, callback)
}

func (f *promiseImpl) OnComplete(callback func(_ interface{}, _ error)) {
	if !f.IsCanceled() {
		if f.IsCompleted() {
			go callback(f.value, f.err)
			return
		}
		f.completeCBs = append(f.completeCBs, callback)
	}
}

func (f *promiseImpl) Success(value interface{}) error {
	if f.IsCanceled() {
		return ErrorCanceled
	}
	if f.IsCompleted() {
		return ErrorCompleted
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

func (f *promiseImpl) Fail(err error) error {
	if f.IsCanceled() {
		return ErrorCanceled
	}
	if f.IsCompleted() {
		return ErrorCompleted
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
		return nil, ErrorTimeout
	}
}

func (f *promiseImpl) IsCompleted() bool {
	if f.IsCanceled() {
		return true
	}
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
	err := f.Fail(ErrorCanceled)
	if f.IsCanceled() {
		return ErrorCanceled
	}
	f.cancel()
	return err
}

// CancelCh returns a channel that is closed when the work held in this Promise has to be
// canceled.
func (f *promiseImpl) CancelCh() <-chan struct{} {
	return f.context.Done()
}
