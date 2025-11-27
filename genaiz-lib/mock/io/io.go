package io

type StubWriter struct {
	WriteBytes []byte
	WriteError error
	WriteSize  int
}

func (s *StubWriter) Write(writeBytes []byte) (n int, err error) {
	s.WriteBytes = writeBytes
	return s.WriteSize, s.WriteError
}
