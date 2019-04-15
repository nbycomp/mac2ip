package main

import (
	"testing"
)

func TestIntToIP(t *testing.T) {
	type args struct {
		ip uint32
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			"starts at 0.0.0.1",
			args{1},
			"0.0.0.1",
			false,
		},
		{
			"ends at 127.255.255.255",
			args{maxValue},
			"127.255.255.255",
			false,
		},
		{
			"returns error when number is out of range",
			args{maxValue + 1},
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IntToIP(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("IntToIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IntToIP() = %v, want %v", got, tt.want)
			}
		})
	}
}
