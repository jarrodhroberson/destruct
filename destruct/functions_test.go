package destruct

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
			want: "23d2c569889575cf9d3fc77c3e6b458a7c64e40f9f33adc6661c4f1b5a9c17cdea43120d0ee6f051dcb0d945e97b56db0e329bef4a305206c31d833d35a5c724",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustHashIdentity(tt.args.t); got != tt.want {
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
			if got := MustHashIdentity(tt.args.t); got != tt.want {
				t.Errorf("HashIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}
