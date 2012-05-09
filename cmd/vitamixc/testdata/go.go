package main
func main() {
	go helloWorld()

	go func() {
		fooBar()
		time.Now()
		time.Sleep(1e9)
	}()
}

