package redact

import (
	"strings"
	"testing"
)

func TestStrictRedactionUsesStablePseudonyms(t *testing.T) {
	t.Parallel()
	detector := NewDetector()
	input := "owner@example.test connected from 192.0.2.10; owner@example.test retried from 192.0.2.10"
	redacted, counts := detector.Redact(input, "strict")
	if strings.Contains(redacted, "owner@example.test") || strings.Contains(redacted, "192.0.2.10") {
		t.Fatalf("sensitive value survived redaction: %s", redacted)
	}
	if strings.Count(redacted, "EMAIL_001") != 2 || strings.Count(redacted, "IPV4_001") != 2 {
		t.Fatalf("pseudonyms are not stable: %s", redacted)
	}
	if counts["email"] != 2 || counts["ipv4"] != 2 {
		t.Fatalf("unexpected replacement counts: %+v", counts)
	}
}

func TestStandardRedactionKeepsNonSecretPII(t *testing.T) {
	t.Parallel()
	detector := NewDetector()
	input := "email=owner@example.test password=synthetic-secret-value"
	redacted, _ := detector.Redact(input, "standard")
	if !strings.Contains(redacted, "owner@example.test") {
		t.Fatalf("standard profile unexpectedly removed email: %s", redacted)
	}
	if strings.Contains(redacted, "synthetic-secret-value") {
		t.Fatalf("password survived standard redaction: %s", redacted)
	}
}

func TestDetectionDoesNotReturnValues(t *testing.T) {
	t.Parallel()
	matches := NewDetector().Detect("Authorization: Bearer synthetic-token-value-123456", true)
	if len(matches) != 1 || matches[0].Kind != "bearer_token" || matches[0].Count != 1 {
		t.Fatalf("unexpected detection result: %+v", matches)
	}
}
