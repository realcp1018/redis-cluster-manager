package redis

import (
	"errors"
	"testing"
)

func TestIsLoadingError(t *testing.T) {
	if !IsLoadingError(errors.New("LOADING Redis is loading the dataset in memory")) {
		t.Fatal("expected LOADING error to be detected")
	}

	if IsLoadingError(errors.New("ERR unknown command")) {
		t.Fatal("expected non-LOADING error to be ignored")
	}
}
