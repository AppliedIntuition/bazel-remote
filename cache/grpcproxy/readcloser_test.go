package grpcproxy

import (
	"errors"
	"net/http"
	"testing"

	"github.com/buchgr/bazel-remote/v2/cache"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeDataMessage struct{}

func (fakeDataMessage) GetData() []byte {
	return nil
}

type fakeRecvStream struct {
	err error
}

func (s fakeRecvStream) Recv() (fakeDataMessage, error) {
	return fakeDataMessage{}, s.err
}

func (s fakeRecvStream) CloseSend() error {
	return nil
}

func TestStreamReadCloserTranslatesNotFound(t *testing.T) {
	rc := StreamReadCloser[fakeDataMessage]{
		Stream: fakeRecvStream{
			err: status.Error(codes.NotFound, "missing digest"),
		},
	}

	_, err := rc.Read(make([]byte, 1))
	if err == nil {
		t.Fatal("expected Read to return an error")
	}

	var cacheErr *cache.Error
	if !errors.As(err, &cacheErr) {
		t.Fatalf("expected cache.Error, got %T", err)
	}
	if cacheErr.Code != http.StatusNotFound {
		t.Fatalf("expected HTTP 404, got %d", cacheErr.Code)
	}
}
