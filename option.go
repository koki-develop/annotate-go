package annotate

// Option is a functional option for configuring a [Renderer].
type Option func(*Renderer)

// WithStyle returns an [Option] that sets the [Renderer]'s global [Style].
func WithStyle(s Style) Option {
	return func(r *Renderer) {
		r.Style = s
	}
}

// WithBefore returns an [Option] that sets the default number of lines
// to display before each labeled line. Negative values are clamped to 0.
func WithBefore(n int) Option {
	return func(r *Renderer) {
		if n < 0 {
			n = 0
		}
		r.Before = n
	}
}

// WithAfter returns an [Option] that sets the default number of lines
// to display after each labeled line. Negative values are clamped to 0.
func WithAfter(n int) Option {
	return func(r *Renderer) {
		if n < 0 {
			n = 0
		}
		r.After = n
	}
}
