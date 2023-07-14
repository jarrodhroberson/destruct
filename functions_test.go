package main

import (
	"testing"
	"time"
)

type person struct {
	FirstName    string
	LastName     string
	EmailAddress string
	BirthDate    time.Time
	CreatedDate  time.Time `identity:"-"`
}

func mustLocation(ianaName string) *time.Location {
	l, err := time.LoadLocation(ianaName)
	if err != nil {
		panic(err)
	}
	return l
}

var EST = mustLocation("America/New_York")

var me = person{
	FirstName:    "Jarrod",
	LastName:     "Roberson",
	EmailAddress: "jarrod@vertigrated.com",
	BirthDate:    time.Date(1967, time.November, 27, 0, 0, 0, 0, EST),
	CreatedDate:  time.Now().UTC(),
}

func TestHashIdentity(t *testing.T) {
	type args[T any] struct {
		t T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want string
	}
	tests := []testCase[person]{
		{
			name: "person exclude CreatedDate",
			args: args[person]{t: me},
			want: "313936372d31312d32375430303a30303a30302d30353a30306a6172726f644076657274696772617465642e636f6d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashIdentity(tt.args.t); got != tt.want {
				t.Errorf("HashIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHashIdentityString(t *testing.T) {
	type args[T any] struct {
		t T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want string
	}
	tests := []testCase[string]{
		{
			name: "1",
			args: args[string]{"1"},
			want: "?",
		},
		{
			name: "2",
			args: args[string]{"1"},
			want: "?",
		},
		{
			name: "3",
			args: args[string]{"1"},
			want: "?",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HashIdentity(tt.args.t); got != tt.want {
				t.Errorf("HashIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}
