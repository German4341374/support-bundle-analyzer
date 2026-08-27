package archiveutil

import (
	"path"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/German4341374/support-bundle-analyzer/internal/apperror"
	"golang.org/x/text/unicode/norm"
)

var windowsDrive = regexp.MustCompile(`^[A-Za-z]:`)

func SafeArchivePath(raw string, maxBytes int) (string, error) {
	if raw == "" || !utf8.ValidString(raw) || strings.IndexByte(raw, 0) >= 0 {
		return "", apperror.New(apperror.ArchiveUnsafeEntry, "archive entry has an invalid filename")
	}
	name := norm.NFC.String(strings.ReplaceAll(raw, `\`, "/"))
	if len(name) > maxBytes {
		return "", apperror.New(apperror.ArchiveLimitExceeded, "archive entry filename exceeds the configured limit")
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == '\u202a' || r == '\u202b' || r == '\u202d' || r == '\u202e' || r == '\u2066' || r == '\u2067' || r == '\u2068' || r == '\u2069' {
			return "", apperror.New(apperror.ArchiveUnsafeEntry, "archive entry filename contains control or bidirectional formatting characters")
		}
	}
	if strings.HasPrefix(name, "/") || windowsDrive.MatchString(name) {
		return "", apperror.New(apperror.ArchivePathTraversal, "absolute archive paths are not allowed")
	}
	cleaned := path.Clean(name)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", apperror.New(apperror.ArchivePathTraversal, "archive entry escapes the extraction root")
	}
	return cleaned, nil
}
