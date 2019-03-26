package words

// Option ...
type Option func(*server)

// Port to listen on.
func Port(p int) Option {
	return func(s *server) {
		s.port = p
	}
}
