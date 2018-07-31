# Mañana

Yet another futures and promises library for Go.

Super-dummy example code:

```go
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/mariomac/manana"
)

func TimeConsumingSynchronousFunction() int {
	<-time.After(3 * time.Second)
	return 42
}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)
	future := manana.Do(func() (interface{}, error) {
		return TimeConsumingSynchronousFunction(), nil
	})
	future.OnSuccess(func(result interface{}) {
		fmt.Println("The result is", result)
		wg.Done()
	})
	future2 := manana.Do(func() (interface{}, error) {
		return TimeConsumingSynchronousFunction(), nil
	})
	future2.OnSuccess(func(result interface{}) {
		fmt.Println("The result of the second invocation is", result)
		wg.Done()
	})
	wg.Wait()
}
```


The library is usable right now. Until I get time to write this README, you
can [read the API docs](https://godoc.org/github.com/mariomac/manana)
and [see few examples](examples).

## TODO
* Change API for a more go-ish management of status via channels.
    - Don't do a OOP-like API but something that just saves us some
      repetitive work with channels.
* An API to query the progress of the future
* A `Then` continuation future
* A Fluent API (concat function invocations)
* More coordination functions: `Any`, `All`, `Pipe`...
* Use atomics to enforce values thread-safety.
* `go generate` to provide a generics-like interface
