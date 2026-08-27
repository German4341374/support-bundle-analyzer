using System.Globalization;
using System.Xml;
using System.Xml.Linq;

namespace SupportBundleAnalyzer.Windows;

public static class WindowsAnalyzer
{
    private static readonly IReadOnlyDictionary<int, (string Rule, string Severity, string Title)> KnownEvents =
        new Dictionary<int, (string, string, string)>
        {
            [4625] = ("WIN_AUTH_FAILURE", "medium", "Windows authentication failure recorded"),
            [6008] = ("WIN_UNEXPECTED_REBOOT", "high", "Unexpected Windows shutdown recorded"),
            [7031] = ("WIN_SERVICE_CRASH", "high", "Windows service terminated unexpectedly"),
            [1000] = ("WIN_APPLICATION_CRASH", "high", "Application crash event recorded"),
            [7] = ("WIN_DISK_WARNING", "high", "Disk I/O warning recorded"),
            [1014] = ("WIN_DNS_WARNING", "medium", "DNS resolution warning recorded")
        };

    public static IReadOnlyList<WindowsEvent> ParseEventXml(Stream stream)
    {
        var settings = new XmlReaderSettings { DtdProcessing = DtdProcessing.Prohibit, XmlResolver = null };
        using var reader = XmlReader.Create(stream, settings);
        var document = XDocument.Load(reader, LoadOptions.None);
        XNamespace ns = "http://schemas.microsoft.com/win/2004/08/events/event";
        IEnumerable<XElement> events = document.Root?.Name.LocalName == "Event"
            ? new[] { document.Root }
            : document.Descendants(ns + "Event");
        return events.Select((element) => ParseEvent(element, ns)).ToArray();
    }

    public static IReadOnlyList<PluginResponse> Analyze(IEnumerable<WindowsEvent> events, string artifact)
    {
        var findings = new List<PluginResponse>();
        foreach (var group in events.Where(item => item.EventId.HasValue && KnownEvents.ContainsKey(item.EventId.Value))
                     .GroupBy(item => item.EventId!.Value).OrderBy(item => item.Key))
        {
            var definition = KnownEvents[group.Key];
            var first = group.First();
            var evidence = new Evidence(artifact, "/Events", SafeExcerpt(first.Message));
            var finding = new Finding(
                definition.Rule,
                definition.Severity,
                definition.Title,
                $"Observed {group.Count()} matching event(s). Event IDs are supporting evidence and do not establish root cause.",
                "moderate",
                new[] { evidence });
            findings.Add(new PluginResponse("finding", finding));
        }
        return findings;
    }

    private static WindowsEvent ParseEvent(XElement element, XNamespace ns)
    {
        var system = element.Element(ns + "System");
        var eventData = element.Element(ns + "EventData");
        var provider = system?.Element(ns + "Provider")?.Attribute("Name")?.Value ?? "unknown";
        var idText = system?.Element(ns + "EventID")?.Value;
        var level = system?.Element(ns + "Level")?.Value ?? "unknown";
        var created = system?.Element(ns + "TimeCreated")?.Attribute("SystemTime")?.Value;
        var message = string.Join("; ", eventData?.Elements(ns + "Data").Select(item => item.Value) ?? []);
        return new WindowsEvent(
            int.TryParse(idText, NumberStyles.None, CultureInfo.InvariantCulture, out var id) ? id : null,
            provider,
            system?.Element(ns + "Channel")?.Value ?? "unknown",
            level,
            DateTimeOffset.TryParse(created, CultureInfo.InvariantCulture, DateTimeStyles.AssumeUniversal, out var time) ? time : null,
            system?.Element(ns + "Computer")?.Value ?? "unknown",
            message);
    }

    private static string SafeExcerpt(string value) => value.Length <= 300 ? value : value[..300] + "…";
}
