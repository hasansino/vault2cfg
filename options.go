package vault2cfg

type Option func(*settings)

func WithTagName(tagName string) Option {
	return func(s *settings) {
		s.tagName = tagName
	}
}
