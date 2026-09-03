package io

import "bytes"

type StubReader struct {
	ReadBytes []byte
	ReadError error
}

func (s *StubReader) Read(b []byte) (n int, err error) {
	if s.ReadError == nil {
		s.ReadBytes = b
		return len(b), nil
	}

	return 0, s.ReadError
}

type StubWriter struct {
	MatchError []byte
	WriteBytes []byte
	WriteError error
}

func (s *StubWriter) Write(writeBytes []byte) (n int, err error) {
	s.WriteBytes = writeBytes

	if len(s.MatchError) > 0 {
		if bytes.Equal(writeBytes, s.MatchError) {
			return 0, s.WriteError
		}

		return len(writeBytes), nil
	}

	return len(s.WriteBytes), s.WriteError
}
