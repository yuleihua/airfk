package cmd

import (
	"testing"
)

func TestCmds(t *testing.T) {
	tests := []struct {
		input string
		ok    bool
	}{
		{"ls -l ", true},
		{"uname -a", true},
		{"touch /tmp/12222323323999", false},
	}
	for _, test := range tests {
		data, err := execCmdline(test.input)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != 0 && test.ok {
			t.Logf("data= %v, want %v", string(data), test.ok)
		} else {
			t.Fatalf("data= %v, want %v", string(data), test.ok)
		}
	}
}
