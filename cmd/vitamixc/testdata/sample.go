package main

func main() {
	// Comment1
	go helloWorld()

	// Comment2
	go func() {
		fooBar()
		// This call will be rewritten too
		time.Now()
		time.Sleep(1e9)
	}()

	// Select
	select {
	case <-ch:
	case ch <- y:
		boom()
	default:
		boom()
	}

	// Send
	ch <- y

	// Receive
	x = <-ch
	x, ok := <-ch
	x, y = <-ch, <-ch
}
