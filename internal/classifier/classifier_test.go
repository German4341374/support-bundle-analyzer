package classifier

import "testing"

func TestDetect(t *testing.T) {
	t.Parallel()
	tests := []struct{ name, content, want string }{
		{name: "capture.har", content: `{"log":{"entries":[{"startedDateTime":"2026-01-01T00:00:00Z"}]}}`, want: "har"},
		{name: "api.log", content: "2026-01-01 ERROR connection refused", want: "generic-log"},
		{name: "event.xml", content: "<Event><System><EventID>7031</EventID></System></Event>", want: "windows-event"},
		{name: "deployment.yaml", content: "apiVersion: apps/v1\nkind: Deployment", want: "kubernetes-manifest"},
		{name: "thread-dump.txt", content: `"main" java.lang.Thread.State: BLOCKED`, want: "jvm-thread-dump"},
		{name: "heap.hprof", content: "binary", want: "recognized-large-binary-artifact"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := Detect(test.name, []byte(test.content)); got != test.want {
				t.Fatalf("Detect() = %q, want %q", got, test.want)
			}
		})
	}
}
