package io.github.german4341374.sba.jvm;

import java.io.BufferedReader;
import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Locale;
import java.util.Map;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public final class JvmAnalyzer {
    private static final Pattern THREAD_STATE = Pattern.compile("java\\.lang\\.Thread\\.State:\\s+([A-Z_]+)");
    private static final Pattern GC_PAUSE = Pattern.compile("(?i)(?:pause|stopped).*?([0-9]+(?:\\.[0-9]+)?)ms");

    private JvmAnalyzer() {
    }

    public static List<Finding> analyze(Path path, String artifactType, String artifactPath) throws IOException {
        if ("jvm-heap-dump".equals(artifactType)) {
            return List.of(finding(
                    "JVM_HEAP_DUMP_RECOGNIZED",
                    "info",
                    "JVM heap dump recognized",
                    "The large binary was not parsed. Use Eclipse MAT or another specialized offline heap analyzer.",
                    artifactPath,
                    0,
                    "recognized_large_binary_artifact"));
        }
        var lines = streamBounded(path, 2_000_000);
        var output = new ArrayList<Finding>();
        var lower = String.join("\n", lines).toLowerCase(Locale.ROOT);
        if (lower.contains("found one java-level deadlock") || lower.contains("found a total of") && lower.contains("deadlock")) {
            output.add(finding("JVM_DEADLOCK_INDICATOR", "critical", "JVM deadlock indicator observed",
                    "The thread dump explicitly reports a possible Java-level deadlock.", artifactPath, findLine(lines, "deadlock"), "deadlock"));
        }
        if (lower.contains("outofmemoryerror")) {
            output.add(finding("JVM_OUT_OF_MEMORY", "high", "OutOfMemoryError evidence observed",
                    "The artifact contains explicit JVM memory-exhaustion evidence.", artifactPath, findLine(lines, "outofmemoryerror"), "OutOfMemoryError"));
        }
        var states = threadStates(lines);
        var blocked = states.getOrDefault("BLOCKED", 0);
        if (blocked >= 3) {
            output.add(finding("JVM_BLOCKED_THREADS", "medium", "Multiple blocked JVM threads",
                    "Observed " + blocked + " BLOCKED thread states; lock ownership and surrounding stacks need review.", artifactPath, 0, "BLOCKED"));
        }
        var longPauses = gcPauses(lines, 1000.0);
        if (longPauses > 0) {
            output.add(finding("JVM_LONG_GC_PAUSE", "high", "Long JVM garbage-collection pauses",
                    "Observed " + longPauses + " pause(s) over 1000 ms.", artifactPath, 0, "GC pause"));
        }
        return output;
    }

    public static Map<String, Integer> threadStates(List<String> lines) {
        var states = new LinkedHashMap<String, Integer>();
        for (var line : lines) {
            Matcher matcher = THREAD_STATE.matcher(line);
            if (matcher.find()) states.merge(matcher.group(1), 1, Integer::sum);
        }
        return states;
    }

    public static int gcPauses(List<String> lines, double thresholdMs) {
        var count = 0;
        for (var line : lines) {
            Matcher matcher = GC_PAUSE.matcher(line);
            if (matcher.find() && Double.parseDouble(matcher.group(1)) > thresholdMs) count++;
        }
        return count;
    }

    private static List<String> streamBounded(Path path, int maxLines) throws IOException {
        var lines = new ArrayList<String>();
        try (BufferedReader reader = Files.newBufferedReader(path, StandardCharsets.UTF_8)) {
            String line;
            while ((line = reader.readLine()) != null && lines.size() < maxLines) lines.add(line);
        }
        return lines;
    }

    private static int findLine(List<String> lines, String needle) {
        for (var index = 0; index < lines.size(); index++) {
            if (lines.get(index).toLowerCase(Locale.ROOT).contains(needle)) return index + 1;
        }
        return 0;
    }

    private static Finding finding(String rule, String severity, String title, String summary,
                                   String artifact, int line, String excerpt) {
        return new Finding(rule, severity, title, summary, "strong",
                List.of(Map.of("artifact", artifact, "lineStart", line, "excerpt", excerpt)));
    }
}
