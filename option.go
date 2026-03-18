package annotate

// Option is a functional option for configuring a [Renderer].
type Option func(*Renderer)

// WithStyle returns an [Option] that sets the [Renderer]'s global [Style].
func WithStyle(s Style) Option {
	return func(r *Renderer) {
		r.Style = s
	}
}
