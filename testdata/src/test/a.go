package test

import (
	"fmt"
	"os"
)

func myCustomPrint(b bool, msg string, format ...any) {
	if b {
		fmt.Printf(msg, format...)
	}
}

func test() {
	name := "John"

	fmt.Printf("This is my name: \"%s\"\n", name) // want `use %q instead of \"%s\" for formatting strings with quotations`
	fmt.Printf("This is my name: %s", name)
	fmt.Printf("This is my name: %q", name)
	fmt.Printf("This is my age: %d", 42)

	fmt.Sprintf("This is my name: \"%s\"\n", name) // want `use %q instead of \"%s\" for formatting strings with quotations`
	fmt.Sprintf("This is my name: %s", name)
	fmt.Sprintf("This is my name: %q", name)
	fmt.Sprintf("This is my age: %d", 42)

	fmt.Fprintf(os.Stdout, "This is my name: \"%s\"\n", name) // want `use %q instead of \"%s\" for formatting strings with quotations`
	fmt.Fprintf(os.Stdout, "This is my name: %s", name)
	fmt.Fprintf(os.Stdout, "This is my name: %q", name)
	fmt.Fprintf(os.Stdout, "This is my age: %d", 42)

	fmt.Errorf("This is my name: \"%s\"\n", name) // want `use %q instead of \"%s\" for formatting strings with quotations`
	fmt.Errorf("This is my name: %s", name)
	fmt.Errorf("This is my name: %q", name)
	fmt.Errorf("This is my age: %d", 42)

	myCustomPrint(true, "This is my name: \"%s\"\n", name) // want `use %q instead of \"%s\" for formatting strings with quotations`
	myCustomPrint(true, "This is my name: %s", name)
	myCustomPrint(true, "This is my name: %q", name)
	myCustomPrint(true, "This is my age: %d", 42)
}
