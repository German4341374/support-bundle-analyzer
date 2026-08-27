package archiveutil

import "testing"

func TestSafeArchivePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "relative", input: "logs/api.log", want: "logs/api.log"},
		{name: "windows separators", input: `logs\api.log`, want: "logs/api.log"},
		{name: "dot normalized", input: "logs/./api.log", want: "logs/api.log"},
		{name: "parent escape", input: "../../etc/passwd", wantErr: true},
		{name: "unix absolute", input: "/etc/passwd", wantErr: true},
		{name: "windows absolute", input: `C:\Windows\win.ini`, wantErr: true},
		{name: "nul", input: "a\x00b", wantErr: true},
		{name: "bidi override", input: "safe\u202etxt.exe", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := SafeArchivePath(test.input, 512)
			if test.wantErr && err == nil {
				t.Fatalf("SafeArchivePath(%q) returned no error", test.input)
			}
			if !test.wantErr && err != nil {
				t.Fatalf("SafeArchivePath(%q): %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("SafeArchivePath(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}

func FuzzSafeArchivePath(f *testing.F) {
	for _, seed := range []string{"logs/api.log", "../../etc/passwd", `C:\Windows\win.ini`, "unicode/данные.log", "a\x00b"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		result, err := SafeArchivePath(input, 512)
		if err == nil && (result == ".." || len(result) == 0 || result[0] == '/') {
			t.Fatalf("unsafe successful result %q for %q", result, input)
		}
	})
}
