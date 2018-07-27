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

func TestFuture_Success(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Given a future
	f := New()

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
	f.Success("suceed!")

	// Both subscribers eventually receive the success
	assert.NoError(t, eventually(2*time.Second, func() {
		wg.Wait()
		assert.Equal(t, "suceed!", second)
		assert.Equal(t, "suceed!", first)
	}))
}

func TestFuture_Success_OneBeforeOneAfterSubscribing(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Given a future
	f := New()

	// When attaching an "onsuccess" callback
	var first, second string
	f.OnSuccess(func(obj interface{}) {
		first = obj.(string)
		wg.Done()
	})

	// And the future succeeds
	f.Success("suceed!")

	// But another "onsuccess" is subscribed after the future finishes
	f.OnSuccess(func(obj interface{}) {
		second = obj.(string)
		wg.Done()
	})

	// Both subscribers eventually receive the success
	assert.NoError(t, eventually(2*time.Second, func() {
		wg.Wait()
		assert.Equal(t, "suceed!", second)
		assert.Equal(t, "suceed!", first)
	}))
}

func TestFuture_Success_AfterSubscribing(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)
	// Given a future
	f := New()

	// When the future succeeds
	f.Success("suceed!")

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
	assert.NoError(t, eventually(2*time.Second, func() {
		wg.Wait()
		assert.Equal(t, "suceed!", second)
		assert.Equal(t, "suceed!", first)
	}))
}

func TestFuture_Error(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Given a future
	f := New()

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
	assert.NoError(t, eventually(2*time.Second, func() {
		wg.Wait()
		assert.EqualError(t, first, "catapun")
		assert.EqualError(t, second, "catapun")
	}))
}

func TestFuture_Error_OneBeforeOneAfterSubscribing(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Given a future
	f := New()

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
	assert.NoError(t, eventually(2*time.Second, func() {
		wg.Wait()
		assert.EqualError(t, first, "catapun")
		assert.EqualError(t, second, "catapun")
	}))
}

func TestFuture_Error_AfterSubscribing(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// Given a future
	f := New()

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
	assert.NoError(t, eventually(2*time.Second, func() {
		wg.Wait()
		assert.EqualError(t, first, "catapun")
		assert.EqualError(t, second, "catapun")
	}))
}

// Test if trying to complete twice

// Test get with success value
// Test that onsuccess still works

// Test get with error value
// Test that onerror still works

// Testear que no haya stall mientras no se especifica onsuccess y onerror
