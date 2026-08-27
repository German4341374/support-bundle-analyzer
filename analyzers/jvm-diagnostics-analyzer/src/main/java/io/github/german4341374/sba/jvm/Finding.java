package io.github.german4341374.sba.jvm;

import java.util.List;
import java.util.Map;

public record Finding(
        String ruleId,
        String severity,
        String title,
        String summary,
        String confidence,
        List<Map<String, Object>> evidence) {
}
