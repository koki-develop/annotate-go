package annotate_test

import (
	"fmt"

	"github.com/koki-develop/annotate-go"
)

func Example() {
	src := []byte(`name = "Alice"
age = 30
`)

	labels := []annotate.Label{
		{Span: annotate.Span{Start: 0, End: 14}, Marker: annotate.MarkerDash, Text: "string field"},
		{Span: annotate.Span{Start: 15, End: 23}, Marker: annotate.MarkerTilde, Text: "integer field"},
	}

	r := annotate.New()
	output, err := r.Render(src, labels)
	if err != nil {
		panic(err)
	}
	fmt.Print(output)
	// Output:
	// 1 | name = "Alice"
	//   | -------------- string field
	// 2 | age = 30
	//   | ~~~~~~~~ integer field
}
