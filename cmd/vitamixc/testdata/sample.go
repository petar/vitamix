package main

func main() {
	go helloWorld()

	go func() {
		fooBar()
		time.Now()
		time.Sleep(1e9)
	}()

	select {
	case <-ch:
	case ch <- y:
		boom()
	default:
		boom()
	}

	ch <- y

	x = <-ch
	x, ok := <-ch
	x, y = <-ch, <-ch
}
