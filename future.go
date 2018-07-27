package manana

// futureImpl here should be probably named "Promise", since has a completable status

type Future interface {
	OnSuccess(successReceiver func(_ interface{}))
	OnError(errorReceiver func(_ error))
	Get() (interface{}, error)
}

type CompletableFuture interface {
	Future
	Success(value interface{})
	Error(err error)
}

type futureImpl struct {
	success    chan interface{}
	error      chan error
	successCBs []func(_ interface{})
	errorCBs   []func(_ error)
}

func New() CompletableFuture {
	return &futureImpl{
		success:    make(chan interface{}),
		error:      make(chan error),
		successCBs: make([]func(_ interface{}), 0),
		errorCBs:   make([]func(_ error), 0),
	}
}

// OnSuccess invokes the statusReceiver function as soon as the future is successfully completed
func (f *futureImpl) OnSuccess(callback func(_ interface{})) {
	startListening := len(f.successCBs) == 0
	f.successCBs = append(f.successCBs, callback)
	if startListening {
		go func() {
			status := <-f.success // Wait for success
			f.close()
			for _, rCallback := range f.successCBs {
				go rCallback(status)
			}
		}()
	}
}

func (f *futureImpl) OnError(callback func(_ error)) {
	startListening := len(f.successCBs) == 0
	f.errorCBs = append(f.errorCBs, callback)
	if startListening {
		go func() {
			err := <-f.error // Wait for success
			f.close()
			for _, rCallback := range f.errorCBs {
				go rCallback(err)
			}
		}()
	}
}

// Todo: return error if future has been completed
func (f *futureImpl) Success(value interface{}) {
	f.success <- value
}

// Todo: return error if future has been completed
func (f *futureImpl) Error(err error) {
	f.error <- err
}

// Todo: oncomplete (interface{}, error)

// Get should coexist and close onsuccess
func (f *futureImpl) Get() (interface{}, error) {
	select {
	case successVal := <-f.success:
		f.close()
		return successVal, nil
	case err := <-f.error:
		f.close()
		return nil, err
	}
}

func (f *futureImpl) close() {
	close(f.success)
	close(f.error)
}
