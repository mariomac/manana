package manana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFutureSuccessHandleOverriding(t *testing.T) {
	// Given a future
	f := NewFuture()
	// When attaching an "onsuccess" receiver
	var first, second string
	f.OnSuccess(func(obj interface{}) {
		first = obj.(string)
	})
	// And adding another before the future succeeds
	f.OnSuccess(func(obj interface{}) {
		second = obj.(string)
	})
	// When the future succeeds
	f.Success("suceeded!")
	// The last receiver gets the success
	assert.Equal(t, "suceeded!", second)
	// but not the first
	assert.Equal(t, "", first)
}


// Test error handle overriding
// Given a future
// When attaching an "on error" receiver
// And adding another before the future fails
// The last receiver gets the error
// but not the first

// Test get with success value

// Test get with error value

// TEstear que no haya stall mientras no se especifica onsuccess y onerror
