package manana

import "sync"

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
