package main

func main() {

	// Receive
	<-ch
	x, ok := <-ch
	x, y = <-ch, <-ch

	// Send
	ch <- y

	select {
	case <-ch:
		hello.World()
	case v := <-ch:
		hello.World()
	case ch <- 5:
		hello.World()
	default:
		hello.World()
	}
}
