package archiveutil

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/German4341374/support-bundle-analyzer/internal/apperror"
	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/ulikunitz/xz"
)

type ExtractResult struct {
	InputSHA256 string
	Files       []string
	TotalBytes  int64
}

type extractionState struct {
	limits     model.Limits
	root       string
	seen       map[string]struct{}
	files      []string
	totalBytes int64
}

func Extract(input, destination string, limits model.Limits) (ExtractResult, error) {
	hash, err := hashFile(input)
	if err != nil {
		return ExtractResult{}, err
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return ExtractResult{}, err
	}
	state := &extractionState{limits: limits, root: destination, seen: make(map[string]struct{})}
	format, err := detectFormat(input)
	if err != nil {
		return ExtractResult{}, err
	}
	switch format {
	case "zip":
		err = extractZIP(input, state)
	case "tar", "tar.gz", "tar.bz2", "tar.xz":
		err = extractTAR(input, format, state)
	case "gzip":
		err = extractSingleGZIP(input, state)
	default:
		err = apperror.New(apperror.ArchiveUnsupported, "unsupported archive format")
	}
	if err != nil {
		return ExtractResult{}, err
	}
	return ExtractResult{InputSHA256: hash, Files: state.files, TotalBytes: state.totalBytes}, nil
}

func detectFormat(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	header := make([]byte, 8)
	n, err := io.ReadFull(f, header)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}
	header = header[:n]
	lower := strings.ToLower(name)
	switch {
	case len(header) >= 4 && string(header[:4]) == "PK\x03\x04":
		return "zip", nil
	case len(header) >= 2 && header[0] == 0x1f && header[1] == 0x8b:
		if strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz") {
			return "tar.gz", nil
		}
		return "gzip", nil
	case len(header) >= 3 && string(header[:3]) == "BZh" && strings.HasSuffix(lower, ".tar.bz2"):
		return "tar.bz2", nil
	case len(header) >= 6 && string(header[:6]) == "\xfd7zXZ\x00" && strings.HasSuffix(lower, ".tar.xz"):
		return "tar.xz", nil
	case strings.HasSuffix(lower, ".tar"):
		return "tar", nil
	default:
		return "", apperror.New(apperror.ArchiveUnsupported, "input is not a recognized supported archive")
	}
}

func (s *extractionState) reserve(name string, size int64) (string, error) {
	safe, err := SafeArchivePath(name, s.limits.MaxFilenameBytes)
	if err != nil {
		return "", err
	}
	key := strings.ToLower(safe)
	if _, ok := s.seen[key]; ok {
		return "", apperror.New(apperror.ArchiveUnsafeEntry, "archive contains duplicate normalized paths")
	}
	if len(s.files)+1 > s.limits.MaxFiles {
		return "", apperror.New(apperror.ArchiveLimitExceeded, "archive exceeds the configured file-count limit")
	}
	if size < 0 || size > s.limits.MaxSingleFileBytes {
		return "", apperror.New(apperror.ArchiveLimitExceeded, "archive entry exceeds the configured single-file limit")
	}
	if s.totalBytes+size > s.limits.MaxTotalBytes {
		return "", apperror.New(apperror.ArchiveLimitExceeded, "archive exceeds the configured total-uncompressed-size limit")
	}
	destination := filepath.Join(s.root, filepath.FromSlash(safe))
	rel, err := filepath.Rel(s.root, destination)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", apperror.New(apperror.ArchivePathTraversal, "archive entry escapes the extraction root")
	}
	s.seen[key] = struct{}{}
	s.files = append(s.files, safe)
	s.totalBytes += size
	return destination, nil
}

func (s *extractionState) reserveUnknown(name string) (string, error) {
	safe, err := SafeArchivePath(name, s.limits.MaxFilenameBytes)
	if err != nil {
		return "", err
	}
	key := strings.ToLower(safe)
	if _, ok := s.seen[key]; ok {
		return "", apperror.New(apperror.ArchiveUnsafeEntry, "archive contains duplicate normalized paths")
	}
	if len(s.files)+1 > s.limits.MaxFiles {
		return "", apperror.New(apperror.ArchiveLimitExceeded, "archive exceeds the configured file-count limit")
	}
	destination := filepath.Join(s.root, filepath.FromSlash(safe))
	rel, err := filepath.Rel(s.root, destination)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", apperror.New(apperror.ArchivePathTraversal, "archive entry escapes the extraction root")
	}
	s.seen[key] = struct{}{}
	s.files = append(s.files, safe)
	return destination, nil
}

func extractZIP(input string, state *extractionState) error {
	reader, err := zip.OpenReader(input)
	if err != nil {
		return apperror.Wrap(apperror.ArtifactMalformed, "cannot read ZIP archive", err)
	}
	defer reader.Close()
	for _, entry := range reader.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		if entry.Mode()&os.ModeSymlink != 0 || !entry.Mode().IsRegular() {
			return apperror.New(apperror.ArchiveUnsafeEntry, "links and special files are not extracted")
		}
		if entry.CompressedSize64 > 0 && float64(entry.UncompressedSize64)/float64(entry.CompressedSize64) > state.limits.MaxCompressionRatio {
			return apperror.New(apperror.ArchiveLimitExceeded, "archive entry exceeds the configured compression-ratio limit")
		}
		destination, err := state.reserve(entry.Name, int64(entry.UncompressedSize64))
		if err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		if err := writeBounded(destination, source, int64(entry.UncompressedSize64)); err != nil {
			source.Close()
			return err
		}
		if err := source.Close(); err != nil {
			return err
		}
	}
	return nil
}

func extractTAR(input, format string, state *extractionState) error {
	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()
	var reader io.Reader = f
	var closer io.Closer
	switch format {
	case "tar.gz":
		gz, err := gzip.NewReader(f)
		if err != nil {
			return apperror.Wrap(apperror.ArtifactMalformed, "cannot read gzip stream", err)
		}
		reader, closer = gz, gz
	case "tar.bz2":
		reader = bzip2.NewReader(f)
	case "tar.xz":
		xzr, err := xz.NewReader(f)
		if err != nil {
			return apperror.Wrap(apperror.ArtifactMalformed, "cannot read xz stream", err)
		}
		reader = xzr
	}
	if closer != nil {
		defer closer.Close()
	}
	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return apperror.Wrap(apperror.ArtifactMalformed, "cannot read TAR archive", err)
		}
		if header.FileInfo().IsDir() {
			continue
		}
		if !header.FileInfo().Mode().IsRegular() {
			return apperror.New(apperror.ArchiveUnsafeEntry, "links, devices, and special TAR entries are not extracted")
		}
		destination, err := state.reserve(header.Name, header.Size)
		if err != nil {
			return err
		}
		if err := writeBounded(destination, tr, header.Size); err != nil {
			return err
		}
	}
	return nil
}

func extractSingleGZIP(input string, state *extractionState) error {
	f, err := os.Open(input)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return apperror.Wrap(apperror.ArtifactMalformed, "cannot read gzip stream", err)
	}
	defer gz.Close()
	name := gz.Name
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	}
	if name == "" {
		name = "payload"
	}
	destination, err := state.reserveUnknown(name)
	if err != nil {
		return err
	}
	written, err := writeUnknownSize(destination, gz, state.limits.MaxSingleFileBytes)
	if err != nil {
		return err
	}
	state.totalBytes += written
	if state.totalBytes > state.limits.MaxTotalBytes {
		return apperror.New(apperror.ArchiveLimitExceeded, "gzip output exceeds the configured total-uncompressed-size limit")
	}
	return nil
}

func writeBounded(destination string, source io.Reader, expected int64) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written, copyErr := io.CopyN(f, source, expected)
	closeErr := f.Close()
	if copyErr != nil {
		return apperror.Wrap(apperror.ArtifactMalformed, "archive entry ended before its declared size", copyErr)
	}
	if written != expected {
		return apperror.New(apperror.ArtifactMalformed, "archive entry size does not match metadata")
	}
	return closeErr
}

func writeUnknownSize(destination string, source io.Reader, limit int64) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return 0, err
	}
	f, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o600)
	if err != nil {
		return 0, err
	}
	written, copyErr := io.Copy(f, io.LimitReader(source, limit+1))
	closeErr := f.Close()
	if copyErr != nil {
		return written, copyErr
	}
	if written > limit {
		return written, apperror.New(apperror.ArchiveLimitExceeded, "gzip output exceeds the configured single-file limit")
	}
	return written, closeErr
}

func hashFile(name string) (string, error) {
	f, err := os.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func DescribeLimits(l model.Limits) string {
	return fmt.Sprintf("max_files=%d max_total_bytes=%d max_single_file_bytes=%d max_compression_ratio=%.1f", l.MaxFiles, l.MaxTotalBytes, l.MaxSingleFileBytes, l.MaxCompressionRatio)
}
