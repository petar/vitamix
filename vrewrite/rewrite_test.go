package vrewrite

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"strconv"
	"testing"
)

type testPair struct {
	Source string
	Output string
}

var (
	tests = []testPair{
		testPair{
			Source:
`
package main
import "time"
func A() { time.Sleep(2*time.Second); println("A", time.Now().UnixNano()) }
func B() { 
	// Comment 1a
	time.Sleep(1*time.Second); 
	// Comment 1b
	println("B", time.Now().UnixNano()) 
}
func main() {
	go A()
	// Comment 2
	go B()
	time.Sleep(3*time.Second)
	println("C", time.Now().UnixNano()) 
}
`,
			Output:
`B 1000000000
A 2000000000
C 3000000000
`,
		},
		testPair{
			Source:
`
package main
import "time"
func main() {
	done := make(chan int)
	go func() {
		time.Sleep(2*time.Second)
		println("B")
		println(time.Now().UnixNano())
		done <- 1
	}()
	go func() {
		time.Sleep(1*time.Second)
		println("A")
		println(time.Now().UnixNano())
		goto __JustALabel
	__JustALabel:
		done <- 1
	}()
	<-done
	<-done
	println("OK")
}
`,
			Output:
`A
1000000000
B
2000000000
OK
`,
		},
		testPair{
			Source:
`
package main
import "time"
func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)
	go func() {
		time.Sleep(2*time.Second)
		ch1 <- 1
	}()
	go func() {
		time.Sleep(1*time.Second)
		ch2 <- 1
	}()
	goto __Label
__Label:
	select {
	case <-ch1:
		println("B")
		println(time.Now().UnixNano())
	case <-ch2:
		println("A")
		println(time.Now().UnixNano())
	}
	println("OK")
}
`,
			Output:
`A
1000000000
B
2000000000
OK
`,
		},
	}
)

func TestRewriteFile(t *testing.T) {
	for i, pair := range tests {
		testSnippet(i, pair.Source, pair.Output, t)
	}
}

func testSnippet(i int, src, exp string, t *testing.T) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "", src, 0 /*parser.ParseComments*/)
	if err != nil {
		t.Fatalf("Problem parsing (%s)\n", err)
	}
	RewriteFile(fileSet, file)
	tmp := path.Join(os.TempDir(), strconv.Itoa(i) + ".go")
	fmt.Printf("Using temporary file `%s`", tmp)

	if err := PrintToFile(tmp, fileSet, file); err != nil {
		t.Errorf("print to file (%s)", err)
	}

	out, err := exec.Command("go", "run", tmp).CombinedOutput()
	if err != nil {
		t.Errorf("error inside virtualized test (%s)", err)
	}
	if string(out) != exp {
		t.Errorf("expected %s, got %s", exp, string(out))
	}
}
