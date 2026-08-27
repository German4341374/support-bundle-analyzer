using System.Text.Json;
using System.Xml;
using SupportBundleAnalyzer.Windows;

var serializerOptions = new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.CamelCase };
while (Console.ReadLine() is { } line)
{
    try
    {
        using var request = JsonDocument.Parse(line, new JsonDocumentOptions { MaxDepth = 32 });
        var root = request.RootElement;
        if (root.GetProperty("protocolVersion").GetString() != "1")
        {
            throw new InvalidDataException("Unsupported protocol version.");
        }
        var artifact = root.GetProperty("artifact");
        var relativePath = artifact.GetProperty("path").GetString() ?? throw new InvalidDataException("artifact.path is required.");
        var workspaceRoot = root.TryGetProperty("context", out var context) && context.TryGetProperty("workspaceRoot", out var rootProperty)
            ? rootProperty.GetString() ?? "."
            : ".";
        var trustedRoot = Path.GetFullPath(workspaceRoot);
        var fullPath = Path.GetFullPath(Path.Combine(trustedRoot, relativePath));
        if (!fullPath.StartsWith(trustedRoot + Path.DirectorySeparatorChar, StringComparison.OrdinalIgnoreCase))
        {
            throw new InvalidDataException("Artifact path escapes workspace root.");
        }
        await using var stream = File.OpenRead(fullPath);
        foreach (var finding in WindowsAnalyzer.Analyze(WindowsAnalyzer.ParseEventXml(stream), relativePath))
        {
            Console.WriteLine(JsonSerializer.Serialize(finding, serializerOptions));
        }
    }
    catch (Exception exception) when (exception is IOException or JsonException or XmlException or InvalidOperationException)
    {
        Console.WriteLine(JsonSerializer.Serialize(new
        {
            type = "error",
            error = new { code = "WINDOWS_ANALYZER_INPUT_INVALID", message = exception.Message }
        }, serializerOptions));
    }
}
