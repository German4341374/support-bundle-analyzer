package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/German4341374/support-bundle-analyzer/internal/compare"
	"github.com/German4341374/support-bundle-analyzer/internal/model"
	"github.com/German4341374/support-bundle-analyzer/internal/pipeline"
	"github.com/German4341374/support-bundle-analyzer/internal/report"
	"github.com/German4341374/support-bundle-analyzer/internal/sanitize"
	"github.com/German4341374/support-bundle-analyzer/internal/synthetic"
	"github.com/German4341374/support-bundle-analyzer/internal/workspace"
)

const (
	exitOK         = 0
	exitUsage      = 2
	exitInput      = 3
	exitAnalysis   = 4
	exitValidation = 5
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(arguments []string) int {
	if len(arguments) == 0 {
		usage(os.Stderr)
		return exitUsage
	}
	var err error
	switch arguments[0] {
	case "analyze":
		err = analyzeCommand(arguments[1:])
	case "report":
		err = reportCommand(arguments[1:])
	case "compare":
		err = compareCommand(arguments[1:])
	case "redact":
		err = redactCommand(arguments[1:])
	case "generate-demo":
		err = demoCommand(arguments[1:])
	case "version", "--version", "-v":
		fmt.Printf("support-bundle-analyzer %s (schema %s)\n", model.ToolVersion, model.SchemaVersion)
		return exitOK
	case "help", "--help", "-h":
		usage(os.Stdout)
		return exitOK
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n", arguments[0])
		usage(os.Stderr)
		return exitUsage
	}
	if err == nil {
		return exitOK
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	if errors.Is(err, flag.ErrHelp) {
		return exitOK
	}
	if os.IsNotExist(err) {
		return exitInput
	}
	return exitAnalysis
}

func analyzeCommand(arguments []string) error {
	set := flag.NewFlagSet("analyze", flag.ContinueOnError)
	output := set.String("output", "analysis-workspace", "new workspace directory")
	timezone := set.String("timezone", "UTC", "IANA timezone for timestamps without an offset")
	jsonOutput := set.Bool("json", false, "write the result summary as JSON")
	if err := set.Parse(arguments); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return fmt.Errorf("usage: support-bundle-analyzer analyze <archive> [--output directory]")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	result, err := pipeline.Run(ctx, pipeline.Options{
		Input: set.Arg(0), Output: *output, Timezone: *timezone,
		Progress: func(stage, message string) {
			if !*jsonOutput {
				fmt.Fprintf(os.Stderr, "[%s] %s\n", stage, message)
			}
		},
	})
	if err != nil {
		return err
	}
	summary := map[string]any{"analysisId": result.Manifest.AnalysisID, "workspace": result.Root, "artifacts": result.Manifest.ArtifactCount, "findings": len(result.Findings), "timelineEvents": len(result.Timeline), "sensitiveMatches": len(result.Sensitive)}
	if *jsonOutput {
		return writeJSON(os.Stdout, summary)
	}
	fmt.Printf("Analysis complete: %d artifacts, %d findings, %d timeline events\nReport: %s\n", result.Manifest.ArtifactCount, len(result.Findings), len(result.Timeline), filepath.Join(result.Root, "report", "index.html"))
	return nil
}

func reportCommand(arguments []string) error {
	set := flag.NewFlagSet("report", flag.ContinueOnError)
	output := set.String("output", "", "report output directory")
	if err := set.Parse(arguments); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return fmt.Errorf("usage: support-bundle-analyzer report <workspace> [--output directory]")
	}
	data, err := workspace.Load(set.Arg(0))
	if err != nil {
		return err
	}
	target := *output
	if target == "" {
		target = filepath.Join(set.Arg(0), "report")
	}
	if err := report.Generate(target, report.Data{Manifest: data.Manifest, Findings: data.Findings, Timeline: data.Timeline, Sensitive: data.Sensitive}); err != nil {
		return err
	}
	fmt.Printf("Report written to %s\n", filepath.Join(target, "index.html"))
	return nil
}

func compareCommand(arguments []string) error {
	set := flag.NewFlagSet("compare", flag.ContinueOnError)
	output := set.String("output", "comparison.json", "JSON comparison output")
	if err := set.Parse(arguments); err != nil {
		return err
	}
	if set.NArg() != 2 {
		return fmt.Errorf("usage: support-bundle-analyzer compare <baseline-workspace> <incident-workspace>")
	}
	baseline, err := workspace.Load(set.Arg(0))
	if err != nil {
		return err
	}
	incident, err := workspace.Load(set.Arg(1))
	if err != nil {
		return err
	}
	result := compare.Workspaces(baseline, incident)
	content, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(*output, append(content, '\n'), 0o600); err != nil {
		return err
	}
	fmt.Printf("Comparison written to %s\n", *output)
	return nil
}

func redactCommand(arguments []string) error {
	set := flag.NewFlagSet("redact", flag.ContinueOnError)
	output := set.String("output", "sanitized-workspace", "new sanitized workspace directory")
	profile := set.String("profile", "standard", "standard or strict")
	if err := set.Parse(arguments); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return fmt.Errorf("usage: support-bundle-analyzer redact <workspace> [--profile standard|strict]")
	}
	result, err := sanitize.Workspace(set.Arg(0), *output, *profile)
	if err != nil {
		return err
	}
	fmt.Printf("Sanitized export written to %s (%d source files reviewed)\n", *output, len(result.Files))
	return nil
}

func demoCommand(arguments []string) error {
	set := flag.NewFlagSet("generate-demo", flag.ContinueOnError)
	if err := set.Parse(arguments); err != nil {
		return err
	}
	if set.NArg() != 1 {
		return fmt.Errorf("usage: support-bundle-analyzer generate-demo <output.zip>")
	}
	if err := synthetic.WriteBundle(set.Arg(0)); err != nil {
		return err
	}
	fmt.Printf("Synthetic support bundle written to %s\n", set.Arg(0))
	return nil
}

func writeJSON(target *os.File, value any) error {
	encoder := json.NewEncoder(target)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func usage(target *os.File) {
	fmt.Fprintln(target, "Universal Support Bundle Analyzer")
	fmt.Fprintln(target, "Usage: support-bundle-analyzer <command> [options]")
	fmt.Fprintln(target, "Commands: analyze, report, compare, redact, generate-demo, version")
	fmt.Fprintln(target, "Run 'support-bundle-analyzer <command> -h' for command options.")
	fmt.Fprintln(target, "Security: bundle content is never executed; review sanitized exports before sharing.")
}

var _ = exitValidation
