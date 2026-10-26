package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

// net.Pipe exercises HTTP shutdown without requiring permission to open TCP sockets.
type pipeListener struct {
	connections chan net.Conn
	closed      chan struct{}
	once        sync.Once
}

func (l *pipeListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.connections:
		return conn, nil
	case <-l.closed:
		return nil, net.ErrClosed
	}
}

func (l *pipeListener) Close() error {
	l.once.Do(func() { close(l.closed) })
	return nil
}

func (l *pipeListener) Addr() net.Addr { return &net.TCPAddr{} }

func TestShutdownWaitsForActiveRequest(t *testing.T) {
	deadline, cleanup := context.WithTimeout(context.Background(), 3*time.Second)
	defer cleanup()
	ctx, cancel := context.WithCancel(deadline)
	defer cancel()
	listener := &pipeListener{connections: make(chan net.Conn, 1), closed: make(chan struct{})}
	client, connection := net.Pipe()
	defer client.Close()
	defer connection.Close()
	listener.connections <- connection
	client.SetDeadline(time.Now().Add(3 * time.Second))
	started, release := make(chan struct{}), make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		select {
		case <-release:
			io.WriteString(w, "completed")
		case <-deadline.Done():
		}
	})}
	defer server.Close()
	done := make(chan error, 1)
	go func() { done <- serve(ctx, server, listener) }()
	responseDone := make(chan error, 1)
	go func() {
		if _, err := io.WriteString(client, "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"); err != nil {
			responseDone <- err
			return
		}
		response, err := http.ReadResponse(bufio.NewReader(client), nil)
		if err != nil {
			responseDone <- err
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err == nil && (response.StatusCode != http.StatusOK || string(body) != "completed") {
			err = fmt.Errorf("unexpected response: %d %q", response.StatusCode, body)
		}
		responseDone <- err
	}()
	select {
	case <-started:
	case <-deadline.Done():
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case <-listener.closed:
	case <-deadline.Done():
		t.Fatal("shutdown did not close the listener")
	}
	select {
	case err := <-done:
		t.Fatalf("shutdown returned before request completed: %v", err)
	default:
	}
	close(release)
	for _, result := range []<-chan error{responseDone, done} {
		select {
		case err := <-result:
			if err != nil {
				t.Fatal(err)
			}
		case <-deadline.Done():
			t.Fatal("request or shutdown did not complete")
		}
	}
}
