package definition

import (
	"fmt"
	"testing"
)

func TestFoo(t *testing.T) {
	definitions, err := StitchDefinitions()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("> # definitions = %d\n", len(definitions.All()))
}
