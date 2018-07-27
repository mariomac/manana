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
	success         chan interface{}
	error           chan error
	successReceiver func(_ interface{})
	errorReceiver   func(_ error)
}

func NewFuture() CompletableFuture {
	return &futureImpl{
		success: make(chan interface{}),
		error:   make(chan error),
	}
}

// OnSuccess invokes the statusReceiver function as soon as the future is successfully completed
func (f *futureImpl) OnSuccess(successReceiver func(_ interface{})) {
	startListening := f.successReceiver == nil
	f.successReceiver = successReceiver
	if startListening {
		go func() {
			status := <-f.success // Wait for success
			f.close()
			if f.successReceiver != nil {
				f.successReceiver(status) // Invoke status receiver with the result
			}
		}()
	}
}

func (f *futureImpl) OnError(errorReceiver func(_ error)) {
	startListening := f.successReceiver == nil
	f.errorReceiver = errorReceiver
	if startListening {
		go func() {
			err := <-f.error // Wait for success
			f.close()
			if f.errorReceiver != nil {
				f.errorReceiver(err) // Invoke status receiver with the result
			}
		}()
	}
}

func (f *futureImpl) Success(value interface{}) {
	f.success <- value
}

func (f *futureImpl) Error(err error) {
	f.error <- err
}

// Get invalidates OnSuccess and OnError
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
