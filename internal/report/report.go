package report

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	viewer "github.com/German4341374/support-bundle-analyzer/apps/static-report-viewer"
	"github.com/German4341374/support-bundle-analyzer/internal/apperror"
	"github.com/German4341374/support-bundle-analyzer/internal/model"
)

type Data struct {
	Manifest  model.Manifest         `json:"manifest"`
	Findings  []model.Finding        `json:"findings"`
	Timeline  []model.TimelineEvent  `json:"timeline"`
	Sensitive []model.SensitiveMatch `json:"sensitive"`
}

func Generate(directory string, data Data) error {
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	serialized, err := json.Marshal(data)
	if err != nil {
		return apperror.Wrap(apperror.ReportGenerationFailed, "cannot serialize report data", err)
	}
	encoded := base64.StdEncoding.EncodeToString(serialized)
	files := map[string][]byte{
		"index.html": viewer.IndexHTML,
		"app.js":     viewer.AppJS,
		"styles.css": viewer.StylesCSS,
		"data.js":    []byte(fmt.Sprintf("window.__SBA_DATA_B64__ = %q;\n", encoded)),
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(directory, name), content, 0o600); err != nil {
			return apperror.Wrap(apperror.ReportGenerationFailed, "cannot write report asset", err)
		}
	}
	return nil
}
