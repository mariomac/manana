package manana

import "sync"

// All returns a future that will succeed when all the argument futures suceed. If only one of the
// futures fail, the returned future will fail with the error of the first failed future.
// If the returned future succeeds, the return value is an array containing the success values of
// all the parameter futures, in the same order as they are passed to the All function.
// TODO: test and correct, if apply
func All(futures ...Future) Future {
	var wg sync.WaitGroup
	wg.Add(len(futures))
	results := make([]interface{}, len(futures))

	allFuture := NewPromise()

	for i, f := range futures {
		future := f
		index := i
		go func() {
			defer wg.Done()
			val, err := future.Get()
			if err != nil {
				allFuture.Fail(err)
			}
			results[index] = val
		}()
	}

	go func() {
		waitCh := make(chan interface{})
		go func() {
			wg.Wait()
			close(waitCh)
		}()
		select {
		case <-waitCh:
			allFuture.Success(results)
		case <-allFuture.(*promiseImpl).context.Done(): // if cancelled
			allFuture.Cancel()
		}
	}()

	return allFuture
}
