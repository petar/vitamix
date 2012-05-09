package main
func main() {
	select {
	case <-ch:
	case ch <- y:
		boom()
	default:
		boom()
	}
}
