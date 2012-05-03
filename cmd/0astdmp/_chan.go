package main

func main() {

	x, ok := <-ch

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
