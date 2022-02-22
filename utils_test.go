// Copyright 2021 The webserver Authors. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package webserver

import (
	"os"
	"testing"
)

func Test_isTimeoutError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Should work",
			args: args{
				err: os.ErrDeadlineExceeded,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTimeoutError(tt.args.err); got != tt.want {
				t.Errorf("isTimeoutError() = %v, want %v", got, tt.want)
			}
		})
	}
}
