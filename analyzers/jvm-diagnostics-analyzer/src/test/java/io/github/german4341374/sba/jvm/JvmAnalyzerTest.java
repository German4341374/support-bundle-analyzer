package io.github.german4341374.sba.jvm;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

class JvmAnalyzerTest {
    @TempDir
    Path temporaryDirectory;

    @Test
    void detectsDeadlockBlockedThreadsAndMemoryEvidence() throws Exception {
        Path dump = temporaryDirectory.resolve("threads.txt");
        Files.writeString(dump, """
                Found one Java-level deadlock:
                java.lang.Thread.State: BLOCKED
                java.lang.Thread.State: BLOCKED
                java.lang.Thread.State: BLOCKED
                java.lang.OutOfMemoryError: Java heap space
                """);
        var findings = JvmAnalyzer.analyze(dump, "jvm-thread-dump", "threads.txt");
        assertEquals(3, findings.size());
        assertTrue(findings.stream().allMatch(finding -> !finding.summary().toLowerCase().contains("definitive root cause")));
    }

    @Test
    void countsLongGarbageCollectionPauses() {
        assertEquals(2, JvmAnalyzer.gcPauses(List.of("Pause 1500ms", "Pause 20ms", "stopped 2200.5ms"), 1000));
    }

    @Test
    void recognizesHeapDumpWithoutParsingIt() throws Exception {
        Path dump = temporaryDirectory.resolve("java.hprof");
        Files.write(dump, new byte[] {0, 1, 2});
        var findings = JvmAnalyzer.analyze(dump, "jvm-heap-dump", "java.hprof");
        assertEquals("JVM_HEAP_DUMP_RECOGNIZED", findings.get(0).ruleId());
    }
}
