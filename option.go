package annotate

type Option func(*Renderer)

func WithStyle(s Style) Option {
	return func(r *Renderer) {
		r.Style = s
	}
}
