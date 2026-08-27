package main

import (
	"reflect"
	"testing"
)

func TestInterspersedFlags(t *testing.T) {
	t.Parallel()
	input := []string{"bundle.zip", "--output", "workspace", "--quiet", "--timezone=UTC"}
	want := []string{"--output", "workspace", "--quiet", "--timezone=UTC", "bundle.zip"}
	got := interspersed(input, map[string]bool{"output": true, "quiet": false, "timezone": true})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
