package foo

import "fmt"

// mockgen -destination=./mocks/mock_Foo.go -package=mocks -source=./foo/foo/go
type Foo interface {
	Do(int) int
}

func Bar(f Foo) {
	fmt.Println("Bar() start....")
	i := f.Do(99)
	fmt.Printf("f.Do(99) = %d \n", i)
	fmt.Println("Bar() end....")
}
