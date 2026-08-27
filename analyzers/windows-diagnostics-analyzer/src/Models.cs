using System.Text.Json.Serialization;

namespace SupportBundleAnalyzer.Windows;

public sealed record WindowsEvent(
    int? EventId,
    string Provider,
    string Channel,
    string Level,
    DateTimeOffset? TimeCreated,
    string Computer,
    string Message);

public sealed record Evidence(
    [property: JsonPropertyName("artifact")] string Artifact,
    [property: JsonPropertyName("jsonPointer")] string JsonPointer,
    [property: JsonPropertyName("excerpt")] string Excerpt);

public sealed record Finding(
    [property: JsonPropertyName("ruleId")] string RuleId,
    [property: JsonPropertyName("severity")] string Severity,
    [property: JsonPropertyName("title")] string Title,
    [property: JsonPropertyName("summary")] string Summary,
    [property: JsonPropertyName("confidence")] string Confidence,
    [property: JsonPropertyName("evidence")] IReadOnlyList<Evidence> Evidence);

public sealed record PluginResponse(
    [property: JsonPropertyName("type")] string Type,
    [property: JsonPropertyName("finding")] Finding Finding);
