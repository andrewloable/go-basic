package integration

import (
	"fmt"
	"os"
	"testing"
)

func TestDumpGorilla(t *testing.T) {
	src, err := os.ReadFile("../../examples/gorilla.bas")
	if err != nil {
		t.Fatal(err)
	}
	goSrc, _, _, genErr := transpile(string(src))
	if genErr != nil {
		t.Fatalf("codegen error: %v", genErr)
	}
	fmt.Println(goSrc)
}
