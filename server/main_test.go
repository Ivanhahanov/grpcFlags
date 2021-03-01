package main

import (
	"context"
	pb "github.com/Ivanhahanov/grpcFlags/flags"
	"testing"
)

type FlagsTests []struct {
	flag   string
	result bool
}

func TestCheckFlag(t *testing.T) {
	s := server{}
	tests := FlagsTests{
		{
			flag:   "Flag{h3110, W0r1d!}",
			result: true,
		},
		{
			flag:   "Flag{h3110, W0r1d!}",
			result: true,
		},
	}

	for _, test := range tests {
		req := &pb.CheckFlagRequest{Flag: []byte(test.flag)}
		resp, err := s.CheckFlag(context.Background(), req)
		if err != nil {
			t.Errorf("CheckFlagTest(%v)=%t, wanted %v", test.flag, resp.Status, test.result)
		}
	}
}
