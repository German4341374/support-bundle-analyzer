package synthetic

import (
	"archive/zip"
	"fmt"
	"os"
	"sort"
)

var outageDemoFiles = map[string]string{
	"config/application.env": "APP_ENV=production\nDATABASE_URL=postgresql://demo:change-me@db.example.test:5432/support\n",
	"logs/api.log":           "2026-08-01T14:31:07Z INFO service=api request_id=req-demo-1 request started\n2026-08-01T14:31:08Z ERROR service=api request_id=req-demo-1 database connection refused user=1457\n2026-08-01T14:31:09Z ERROR service=api request_id=req-demo-2 database connection refused user=9182\n2026-08-01T14:31:12Z WARN service=api request_id=req-demo-3 upstream timeout after 5000ms\n",
	"network/capture.har":    `{"log":{"version":"1.2","creator":{"name":"synthetic-fixture","version":"1"},"entries":[{"startedDateTime":"2026-08-01T14:31:08Z","time":2100,"request":{"method":"GET","url":"https://api.example.test/v1/tickets?token=demo-only","headers":[{"name":"Authorization","value":"Bearer example_token_value_123456"}]},"response":{"status":503,"statusText":"Service Unavailable","headers":[{"name":"Content-Type","value":"application/json"}],"content":{"size":57,"mimeType":"application/json"}},"timings":{"blocked":1,"dns":2,"connect":3,"send":1,"wait":2080,"receive":13}}]}}`,
	"README.txt":             "Synthetic database outage support bundle. All identities and credentials are fictional.\n",
}

var healthyDemoFiles = map[string]string{
	"config/application.env": "APP_ENV=production\nDATABASE_URL=postgresql://demo:change-me@db.example.test:5432/support\n",
	"logs/api.log":           "2026-08-01T14:21:07Z INFO service=api request_id=req-demo-healthy-1 request started\n2026-08-01T14:21:08Z INFO service=api request_id=req-demo-healthy-1 request completed status=200 duration_ms=118\n",
	"network/capture.har":    `{"log":{"version":"1.2","creator":{"name":"synthetic-fixture","version":"1"},"entries":[{"startedDateTime":"2026-08-01T14:21:08Z","time":118,"request":{"method":"GET","url":"https://api.example.test/v1/tickets","headers":[]},"response":{"status":200,"statusText":"OK","headers":[{"name":"Content-Type","value":"application/json"}],"content":{"size":31,"mimeType":"application/json"}},"timings":{"blocked":1,"dns":2,"connect":3,"send":1,"wait":102,"receive":9}}]}}`,
	"README.txt":             "Synthetic healthy support bundle. All identities and credentials are fictional.\n",
}

func WriteBundle(destination string) error {
	return WriteBundleScenario(destination, "database-outage")
}

func WriteBundleScenario(destination, scenario string) error {
	var demoFiles map[string]string
	switch scenario {
	case "database-outage":
		demoFiles = outageDemoFiles
	case "healthy":
		demoFiles = healthyDemoFiles
	default:
		return fmt.Errorf("unsupported demo scenario %q", scenario)
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("destination already exists: %s", destination)
	} else if !os.IsNotExist(err) {
		return err
	}
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	writer := zip.NewWriter(f)
	keys := make([]string, 0, len(demoFiles))
	for name := range demoFiles {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		entry, err := writer.Create(name)
		if err != nil {
			writer.Close()
			f.Close()
			return err
		}
		if _, err := entry.Write([]byte(demoFiles[name])); err != nil {
			writer.Close()
			f.Close()
			return err
		}
	}
	if err := writer.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
