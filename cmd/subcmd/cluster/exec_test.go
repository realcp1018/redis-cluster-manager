package cluster

import (
	"errors"
	"testing"

	goRedis "github.com/redis/go-redis/v9"
)

func TestFormatExecResult(t *testing.T) {
	tests := []struct {
		name   string
		stdout interface{}
		err    error
		want   interface{}
	}{
		{
			name: "nil redis reply",
			err:  goRedis.Nil,
			want: "(nil)",
		},
		{
			name: "command error",
			err:  errors.New("ERR unknown command"),
			want: "Error executing command: ERR unknown command",
		},
		{
			name:   "successful reply",
			stdout: "OK",
			want:   "OK",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatExecResult(tt.stdout, tt.err); got != tt.want {
				t.Fatalf("formatExecResult() = %v, want %v", got, tt.want)
			}
		})
	}
}
