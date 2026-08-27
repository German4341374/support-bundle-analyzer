package io.github.german4341374.sba.jvm;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.nio.file.Path;
import java.util.Map;

public final class PluginMain {
    private PluginMain() {
    }

    public static void main(String[] args) throws Exception {
        var mapper = new ObjectMapper();
        try (var reader = new BufferedReader(new InputStreamReader(System.in, StandardCharsets.UTF_8))) {
            String line;
            while ((line = reader.readLine()) != null) {
                try {
                    JsonNode request = mapper.readTree(line);
                    if (!"1".equals(request.path("protocolVersion").asText())) throw new IllegalArgumentException("Unsupported protocol version");
                    String artifactPath = request.path("artifact").path("path").asText();
                    String artifactType = request.path("artifact").path("type").asText();
                    Path root = Path.of(request.path("context").path("workspaceRoot").asText(".")).toAbsolutePath().normalize();
                    Path target = root.resolve(artifactPath).normalize();
                    if (!target.startsWith(root)) throw new IllegalArgumentException("Artifact path escapes workspace root");
                    for (Finding finding : JvmAnalyzer.analyze(target, artifactType, artifactPath)) {
                        System.out.println(mapper.writeValueAsString(Map.of("type", "finding", "finding", finding)));
                    }
                } catch (Exception exception) {
                    System.out.println(mapper.writeValueAsString(Map.of(
                            "type", "error",
                            "error", Map.of("code", "JVM_ANALYZER_INPUT_INVALID", "message", exception.getMessage()))));
                }
            }
        }
    }
}
