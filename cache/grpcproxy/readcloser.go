package grpcproxy

import (
	"io"
	"net/http"

	"github.com/buchgr/bazel-remote/v2/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StreamReadCloser[M DataMessage] struct {
	Stream RecvStream[M]
	buf    []byte
}

type DataMessage interface {
	GetData() []byte
}

type RecvStream[M DataMessage] interface {
	Recv() (M, error)
	CloseSend() error
}

func translateStreamError(err error) error {
	if err == nil || err == io.EOF {
		return err
	}

	if status.Code(err) == codes.NotFound {
		return &cache.Error{
			Code: http.StatusNotFound,
			Text: err.Error(),
		}
	}

	return err
}

func (s *StreamReadCloser[M]) readFromBuf(p []byte) int {
	n := len(s.buf)
	if len(p) < n {
		n = len(p)
	}
	copy(p, s.buf[:n])
	if n == len(s.buf) {
		s.buf = nil
	} else {
		s.buf = s.buf[n:]
	}
	return n
}

func (s *StreamReadCloser[M]) Read(p []byte) (int, error) {
	n := 0
	if s.buf != nil {
		n = s.readFromBuf(p)
	}
	if n == len(p) {
		return n, nil
	}
	msg, err := s.Stream.Recv()
	if err == io.EOF {
		err := s.Stream.CloseSend()
		if err != nil {
			return -1, err
		}
	} else if err != nil {
		return -1, translateStreamError(err)
	}
	s.buf = msg.GetData()
	n += s.readFromBuf(p[n:])
	return n, err
}

func (s *StreamReadCloser[M]) Close() error {
	return s.Stream.CloseSend()
}
