package manana

import (
	"testing"

	"sync"
	"time"

	"fmt"

	"errors"

	"github.com/stretchr/testify/assert"
)

func eventually(timeout time.Duration, f func()) error {
	testFinish := make(chan interface{})
	defer close(testFinish)
	go func() {
		f()
		testFinish <- struct{}{}
	}()

	select {
	case <-testFinish:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("function hasn't completed after %v", timeout)
	}

}

func Test_IsCompleted_True_Success(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		// Given a future
		f := NewPromise()

		// When it succeeds
		f.Success("hi!")

		// The IsCompleted function is true
		assert.True(t, f.IsCompleted())
	}))
}

func Test_IsCompleted_True_Error(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		// Given a future
		f := NewPromise()

		// When it fails
		f.Error(errors.New("pum"))

		// The IsCompleted function is true
		assert.True(t, f.IsCompleted())
	}))
}

func Test_IsCompleted_False(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		// Given an uncomplete future
		f := NewPromise()

		// The IsCompleted function is false
		assert.False(t, f.IsCompleted())
	}))
}

func TestFuture_Success(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		// Given a future
		f := NewPromise()

		// When attaching two "onsuccess" callbacks
		var first, second string
		f.OnSuccess(func(obj interface{}) {
			first = obj.(string)
			wg.Done()
		})
		f.OnSuccess(func(obj interface{}) {
			second = obj.(string)
			wg.Done()
		})

		// When the future succeeds
		assert.NoError(t, f.Success("suceed!"))

		// Both subscribers eventually receive the success
		wg.Wait()
		assert.Equal(t, "suceed!", second)
		assert.Equal(t, "suceed!", first)
	}))
}

func TestFuture_Success_OneBeforeOneAfterSubscribing(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		// Given a future
		f := NewPromise()

		// When attaching an "onsuccess" callback
		var first, second string
		f.OnSuccess(func(obj interface{}) {
			first = obj.(string)
			wg.Done()
		})

		// And the future succeeds
		assert.NoError(t, f.Success("suceed!"))

		// But another "onsuccess" is subscribed after the future finishes
		f.OnSuccess(func(obj interface{}) {
			second = obj.(string)
			wg.Done()
		})

		// Both subscribers eventually receive the success
		wg.Wait()
		assert.Equal(t, "suceed!", second)
		assert.Equal(t, "suceed!", first)
	}))
}

func TestFuture_Success_AfterSubscribing(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(2)
		// Given a future
		f := NewPromise()

		// When the future succeeds
		assert.NoError(t, f.Success("suceed!"))

		// And the callbacks have been subscribed after succeeding
		var first, second string
		f.OnSuccess(func(obj interface{}) {
			first = obj.(string)
			wg.Done()
		})
		f.OnSuccess(func(obj interface{}) {
			second = obj.(string)
			wg.Done()
		})

		// Both subscribers eventually receive the success
		wg.Wait()
		assert.Equal(t, "suceed!", second)
		assert.Equal(t, "suceed!", first)
	}))
}

func TestFuture_Error(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		// Given a future
		f := NewPromise()

		// When attaching two "OnError" callbacks
		var first, second error
		f.OnError(func(err error) {
			first = err
			wg.Done()
		})

		f.OnError(func(err error) {
			second = err
			wg.Done()
		})

		// When the future fails with error
		f.Error(errors.New("catapun"))

		// Both subscribers eventually receive the error
		wg.Wait()
		assert.EqualError(t, first, "catapun")
		assert.EqualError(t, second, "catapun")
	}))
}

func TestFuture_Error_OneBeforeOneAfterSubscribing(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(2)

		// Given a future
		f := NewPromise()

		// that already has an "OnError" callbacks
		var first, second error
		f.OnError(func(err error) {
			first = err
			wg.Done()
		})

		// When the future fails with error
		f.Error(errors.New("catapun"))

		// And a new error callback is attached
		f.OnError(func(err error) {
			second = err
			wg.Done()
		})

		// Both subscribers eventually receive the error
		wg.Wait()
		assert.EqualError(t, first, "catapun")
		assert.EqualError(t, second, "catapun")
	}))
}

func TestFuture_Error_AfterSubscribing(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {

		wg := sync.WaitGroup{}
		wg.Add(2)

		// Given a future
		f := NewPromise()

		// That is completed with error before any callback is added
		f.Error(errors.New("catapun"))

		// And error callbacks are added later
		var first, second error
		f.OnError(func(err error) {
			first = err
			wg.Done()
		})
		f.OnError(func(err error) {
			second = err
			wg.Done()
		})

		// Both subscribers eventually receive the error
		wg.Wait()
		assert.EqualError(t, first, "catapun")
		assert.EqualError(t, second, "catapun")
	}))
}

func TestFuture_Success_Twice(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		var val string
		f.OnSuccess(func(obj interface{}) {
			val = obj.(string)
			wg.Done()
		})

		// When the future succeeds
		assert.NoError(t, f.Success("suceed!"))

		// And try to succeed a second time
		assert.Error(t, f.Success("another succeed!"))

		// The subscriber eventually receives the success
		wg.Wait()
		assert.Equal(t, "suceed!", val)

		// And the process works for future things
		wg.Add(1)
		var val2 string
		f.OnSuccess(func(obj interface{}) {
			val2 = obj.(string)
			wg.Done()
		})
		wg.Wait()
		assert.Equal(t, "suceed!", val2)

	}))
}

func TestFuture_Error_Twice(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		var rcvErr error
		f.OnError(func(err error) {
			rcvErr = err
			wg.Done()
		})

		// When the future fails
		assert.NoError(t, f.Error(errors.New("catapun")))

		// And try to fail a second time
		assert.Error(t, f.Error(errors.New("catapunchinpun")))

		// The subscriber eventually receives the error
		wg.Wait()
		assert.EqualError(t, rcvErr, "catapun")

		// And the process works for future things
		wg.Add(1)
		var rcvErr2 error
		f.OnError(func(err error) {
			rcvErr2 = err
			wg.Done()
		})
		wg.Wait()
		assert.EqualError(t, rcvErr2, "catapun")

	}))
}

func TestFuture_ErrorAfterSuccess(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		var rcvErr error
		f.OnError(func(err error) {
			rcvErr = err
			assert.Fail(t, "Error should never happen")
		})
		var val string
		f.OnSuccess(func(obj interface{}) {
			val = obj.(string)
			wg.Done()
		})

		// When the future succeeds
		assert.NoError(t, f.Success("success!"))

		// And later try to fail
		assert.Error(t, f.Error(errors.New("catapun")))

		// The subscriber eventually receives the success value
		wg.Wait()
		assert.Equal(t, "success!", val)

		// But not the error
		assert.NoError(t, rcvErr)
	}))
}

func TestFuture_SuccessAfterError(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		var val string
		f.OnSuccess(func(obj interface{}) {
			val = obj.(string)
			assert.Fail(t, "Success should never happen")
		})

		var rcvErr error
		f.OnError(func(err error) {
			rcvErr = err
			wg.Done()
		})

		// When the future fails
		assert.NoError(t, f.Error(errors.New("catapun")))

		// And later try to succeed
		assert.Error(t, f.Success("success!"))

		// The subscriber eventually receives the error
		wg.Wait()
		assert.EqualError(t, rcvErr, "catapun")

		// But not the success
		assert.Equal(t, "", val)
	}))
}

func TestFuture_Get_Success(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {

		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		var val interface{}
		var err error
		// That synchronously waits for a success
		go func() {
			val, err = f.Get()
			wg.Done()
		}()

		// When the future succeeds
		f.Success("success!")

		wg.Wait()
		// The value has been correctly obtained
		assert.Equal(t, "success!", val)
		assert.NoError(t, err)
	}))
}

func TestFuture_Get_AfterSuccess(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		// Given a future
		f := NewPromise()

		// that succeeded
		f.Success("success!")

		var val interface{}
		var err error
		// When synchronously waiting for a success
		val, err = f.Get()

		// The value has been correctly obtained
		assert.Equal(t, "success!", val)
		assert.NoError(t, err)
	}))
}

func TestFuture_Get_Waiting(t *testing.T) {
	// Given a future
	f := NewPromise()

	err := eventually(200*time.Millisecond, func() {
		// When synchronously waiting for a completion
		f.Get()
	})

	// The Get function doesn't return if the future does not complete
	assert.Error(t, err)
}

func TestFuture_Get_OnSuccess_After(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {

		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		// That synchronously waits for a success
		go func() {
			f.Get()
		}()

		// When the future succeeds
		f.Success("success!")

		// And success callbacks are registered

		var val interface{}
		f.OnSuccess(func(ret interface{}) {
			val = ret
			wg.Done()
		})

		wg.Wait()
		// Then the success value is immediately assigned
		assert.Equal(t, "success!", val)
	}))
}

func TestFuture_Get_Error(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {

		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		var val interface{}
		var err error
		// That synchronously waits for a completion
		go func() {
			val, err = f.Get()
			wg.Done()
		}()

		// When the future fails
		f.Error(fmt.Errorf("pumchimpun"))

		wg.Wait()
		// The error has been received
		assert.EqualError(t, err, "pumchimpun")
		assert.Nil(t, val)
	}))
}

func TestFuture_Get_AfterError(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {
		// Given a future
		f := NewPromise()

		// that fails
		f.Error(fmt.Errorf("pumchimpun"))

		var val interface{}
		var err error
		// When synchronously waiting for a completion
		val, err = f.Get()

		// The error has been received
		assert.EqualError(t, err, "pumchimpun")
		assert.Nil(t, val)
	}))
}

func TestFuture_Get_OnError_After(t *testing.T) {
	assert.NoError(t, eventually(2*time.Second, func() {

		wg := sync.WaitGroup{}
		wg.Add(1)

		// Given a future
		f := NewPromise()

		// That synchronously waits for a success
		go func() {
			f.Get()
		}()

		// When the future fails
		f.Error(fmt.Errorf("pumchimpun"))

		// And error callbacks are registered
		var err error
		f.OnError(func(recv error) {
			err = recv
			wg.Done()
		})

		wg.Wait()
		// The error has been received
		assert.EqualError(t, err, "pumchimpun")
	}))
}

