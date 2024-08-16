package p2p

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// TimeoutStream wraps a network.Stream with read and write timeouts
type TimeoutStream struct {
	stream       network.Stream
	readTimeout  time.Duration
	writeTimeout time.Duration
	bufferSize   uint
}

// NewTimeoutStream creates a network.Stream with read and write timeouts
func NewTimeoutStream(ctx context.Context, host host.Host, p peer.ID, readTimeout, writeTimeout time.Duration, bufferSize uint, pids ...protocol.ID) (*TimeoutStream, error) {
	stream, err := host.NewStream(ctx, p, pids...)
	if err != nil {
		return nil, fmt.Errorf("error enabling stream to %s: %w", p.String(), err)
	}

	return AddTimeoutToStream(stream, readTimeout, writeTimeout, bufferSize), nil
}

// AddTimeoutToStream wraps a network.Stream with TimeoutStream
func AddTimeoutToStream(s network.Stream, readTimeout, writeTimeout time.Duration, bufferSize uint) *TimeoutStream {
	return &TimeoutStream{
		stream:       s,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
		bufferSize:   bufferSize,
	}
}

// ReadWithTimeout reads from the stream with a timeout
func (t *TimeoutStream) ReadWithTimeout() ([]byte, error) {
	var data []byte
	if t.readTimeout > 0 {
		err := t.stream.SetReadDeadline(time.Now().Add(t.readTimeout))
		if err != nil {
			return []byte{}, err
		}
	}

	buf := make([]byte, t.bufferSize)
	// read until EOF
	for {
		n, err := t.stream.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}

		if err == io.EOF {
			// end of stream, break the loop
			break
		}

		if err != nil {
			return []byte{}, err
		}
	}

	return data, nil
}

// WriteWithTimeout writes to the stream with a timeout
func (t *TimeoutStream) WriteWithTimeout(buf []byte) (int, error) {
	if t.writeTimeout > 0 {
		err := t.stream.SetWriteDeadline(time.Now().Add(t.writeTimeout))
		if err != nil {
			return 0, err
		}
	}
	return t.stream.Write(buf)
}

// Close closes the stream
func (t *TimeoutStream) Close() error {
	return t.stream.Close()
}
