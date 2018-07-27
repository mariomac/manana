package manana

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuture_SuccessHandles(t *testing.T) {
	// Given a future
	f := New()
	// When attaching two "onsuccess" callbacks
	var first, second string
	f.OnSuccess(func(obj interface{}) {
		first = obj.(string)
	})
	f.OnSuccess(func(obj interface{}) {
		second = obj.(string)
	})
	// When the future succeeds
	f.Success("suceeded!")
	// The last receiver gets the success
	assert.Equal(t, "suceeded!", second)
	// but not the first
	assert.Equal(t, "suceeded!", first)
}


// Test error handle overriding
// Given a future
// When attaching an "on error" receiver
// And adding another before the future fails
// The last receiver gets the error
// but not the first

// Test get with success value
// Test that onsuccess still works

// Test get with error value
// Test that onerror still works

// Testear que no haya stall mientras no se especifica onsuccess y onerror
