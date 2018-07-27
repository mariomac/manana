package manana

import (
	"fmt"
	"time"
)

// futureImpl here should be probably named "Promise", since has a completable status

type Future interface {
	OnSuccess(callback func(_ interface{}))
	OnError(callback func(_ error))
	Get() (interface{}, error)
	Eventually(timeout time.Duration) (interface{}, error)
}

type Promise interface {
	Future
	Success(value interface{})
	Error(err error)
}

type futureImpl struct {
	success      chan interface{}
	error        chan error
	successCBs   []func(_ interface{})
	errorCBs     []func(_ error)
	completedVal interface{}
	completedErr error
}

func New() Promise {
	return &futureImpl{
		success:    make(chan interface{}),
		error:      make(chan error),
		successCBs: make([]func(_ interface{}), 0),
		errorCBs:   make([]func(_ error), 0),
	}
}

// OnSuccess invokes the statusReceiver function as soon as the future is successfully completed
func (f *futureImpl) OnSuccess(callback func(_ interface{})) {
	if f.completedVal != nil {
		go callback(f.completedVal)
		return
	}
	startListening := len(f.successCBs) == 0
	f.successCBs = append(f.successCBs, callback)
	if startListening {
		go func() {
			f.completedVal = <-f.success // Wait for success
			for _, rCallback := range f.successCBs {
				go rCallback(f.completedVal)
			}
		}()
	}
}

func (f *futureImpl) OnError(callback func(_ error)) {
	if f.completedErr != nil {
		go callback(f.completedErr)
		return
	}
	startListening := len(f.errorCBs) == 0
	f.errorCBs = append(f.errorCBs, callback)
	if startListening {
		go func() {
			f.completedErr = <-f.error // Wait for success
			for _, rCallback := range f.errorCBs {
				go rCallback(f.completedErr)
			}
		}()
	}
}

// Todo: return error if future has been completed
func (f *futureImpl) Success(value interface{}) {
	go func() {
		f.success <- value
		f.close()
	}()
}

// Todo: return error if future has been completed
func (f *futureImpl) Error(err error) {
	go func() {
		f.error <- err
		f.close()
	}()
}

// Todo: oncomplete (interface{}, error)

// Get should coexist and close onsuccess
func (f *futureImpl) Get() (interface{}, error) {
	select {
	case successVal := <-f.success:
		return successVal, nil
	case err := <-f.error:
		return nil, err
	}
}

func (f *futureImpl) Eventually(timeout time.Duration) (interface{}, error) {
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

func (f *futureImpl) close() {
	close(f.success)
	close(f.error)
}
