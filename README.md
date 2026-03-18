# annotate-go

A Go library for rendering annotated source code with line numbers and labeled markers.

## Installation

```sh
go get github.com/koki-develop/annotate-go
```

## Usage

```go
src := []byte(`name = "Alice"
age = 30
`)
labels := []annotate.Label{
	{Span: annotate.Span{Start: 0, End: 4}, Marker: annotate.MarkerDash, Text: "string field"},
	{Span: annotate.Span{Start: 15, End: 18}, Marker: annotate.MarkerTilde, Text: "integer field"},
}
r := annotate.New()
output, _ := r.Render(src, labels)
fmt.Print(output)
```

```
1 | name = "Alice"
  | ---- string field
2 | age = 30
  | ~~~ integer field
```

See [pkg.go.dev](https://pkg.go.dev/github.com/koki-develop/annotate-go) for full API documentation.

## License

[MIT](./LICENSE)
