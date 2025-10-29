package net

import (
	"net"
	"time"
)

type StubConn struct {
}

func (s StubConn) Read(b []byte) (n int, err error) {
	return 0, nil
}

func (s StubConn) Write(b []byte) (n int, err error) {
	return 0, nil
}

func (s StubConn) LocalAddr() net.Addr {
	return nil
}

func (s StubConn) RemoteAddr() net.Addr {
	return nil
}

func (s StubConn) SetDeadline(t time.Time) error {
	return nil
}

func (s StubConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (s StubConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (s StubConn) Close() error {
	return nil
}
