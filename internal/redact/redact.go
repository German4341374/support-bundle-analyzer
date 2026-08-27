package redact

import (
	"regexp"
	"sort"
	"strings"
)

type Detector struct {
	patterns []pattern
	seen     map[string]map[string]string
	counters map[string]int
}

type pattern struct {
	kind        string
	re          *regexp.Regexp
	strictOnly  bool
	replacement string
}

type MatchCount struct {
	Kind  string
	Count int
}

func NewDetector() *Detector {
	return &Detector{
		patterns: []pattern{
			{kind: "private_key", re: regexp.MustCompile(`-----BEGIN (?:RSA |EC |OPENSSH )?PRIVATE KEY-----[\s\S]*?-----END (?:RSA |EC |OPENSSH )?PRIVATE KEY-----`)},
			{kind: "bearer_token", re: regexp.MustCompile(`(?i)Bearer\s+[A-Za-z0-9._~+/=-]{12,}`)},
			{kind: "basic_authorization", re: regexp.MustCompile(`(?i)Basic\s+[A-Za-z0-9+/=]{8,}`)},
			{kind: "jwt", re: regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)},
			{kind: "github_token", re: regexp.MustCompile(`\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,})\b`)},
			{kind: "aws_access_key", re: regexp.MustCompile(`\b(?:AKIA|ASIA)[A-Z0-9]{16}\b`)},
			{kind: "password", re: regexp.MustCompile(`(?i)\b(?:password|passwd|pwd|api[_-]?key|secret|access[_-]?token|refresh[_-]?token)\s*[:=]\s*[^\s,;]{4,}`)},
			{kind: "connection_string", re: regexp.MustCompile(`(?i)\b(?:postgres(?:ql)?|mysql|mongodb(?:\+srv)?)://[^\s]+`)},
			{kind: "email", re: regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`), strictOnly: true},
			{kind: "ipv4", re: regexp.MustCompile(`\b(?:25[0-5]|2[0-4]\d|1?\d?\d)(?:\.(?:25[0-5]|2[0-4]\d|1?\d?\d)){3}\b`), strictOnly: true},
			{kind: "phone", re: regexp.MustCompile(`\b\+?[1-9][0-9 ()-]{7,}[0-9]\b`), strictOnly: true},
			{kind: "home_path", re: regexp.MustCompile(`(?i)(?:/home/|/Users/|C:\\Users\\)[A-Za-z0-9._-]+`), strictOnly: true},
		},
		seen:     make(map[string]map[string]string),
		counters: make(map[string]int),
	}
}

func (d *Detector) Detect(text string, strict bool) []MatchCount {
	counts := make(map[string]int)
	for _, candidate := range d.patterns {
		if candidate.strictOnly && !strict {
			continue
		}
		counts[candidate.kind] += len(candidate.re.FindAllStringIndex(text, -1))
	}
	result := make([]MatchCount, 0, len(counts))
	for kind, count := range counts {
		if count > 0 {
			result = append(result, MatchCount{Kind: kind, Count: count})
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Kind < result[j].Kind })
	return result
}

func (d *Detector) Redact(text, profile string) (string, map[string]int) {
	strict := strings.EqualFold(profile, "strict")
	replacements := make(map[string]int)
	redacted := text
	for _, candidate := range d.patterns {
		if candidate.strictOnly && !strict {
			continue
		}
		redacted = candidate.re.ReplaceAllStringFunc(redacted, func(value string) string {
			replacements[candidate.kind]++
			return d.pseudonym(candidate.kind, value)
		})
	}
	return redacted, replacements
}

func (d *Detector) pseudonym(kind, value string) string {
	if d.seen[kind] == nil {
		d.seen[kind] = make(map[string]string)
	}
	if existing, ok := d.seen[kind][value]; ok {
		return existing
	}
	d.counters[kind]++
	label := strings.ToUpper(kind) + "_" + leftPad(d.counters[kind], 3)
	d.seen[kind][value] = label
	return label
}

func leftPad(value, width int) string {
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	if digits == "" {
		digits = "0"
	}
	for len(digits) < width {
		digits = "0" + digits
	}
	return digits
}
