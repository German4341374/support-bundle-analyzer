package apperror

import "fmt"

type Code string

const (
	ArchivePathTraversal   Code = "ARCHIVE_PATH_TRAVERSAL"
	ArchiveLimitExceeded   Code = "ARCHIVE_LIMIT_EXCEEDED"
	ArchiveUnsupported     Code = "ARCHIVE_UNSUPPORTED_FORMAT"
	ArchiveUnsafeEntry     Code = "ARCHIVE_UNSAFE_ENTRY"
	ArtifactMalformed      Code = "ARTIFACT_MALFORMED"
	ArtifactTooLarge       Code = "ARTIFACT_TOO_LARGE"
	AnalyzerFailed         Code = "ANALYZER_FAILED"
	AnalyzerTimeout        Code = "ANALYZER_TIMEOUT"
	ReportGenerationFailed Code = "REPORT_GENERATION_FAILED"
)

type Error struct {
	Code    Code
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
}

func (e *Error) Unwrap() error { return e.Cause }

func New(code Code, message string) error {
	return &Error{Code: code, Message: message}
}

func Wrap(code Code, message string, cause error) error {
	return &Error{Code: code, Message: message, Cause: cause}
}
