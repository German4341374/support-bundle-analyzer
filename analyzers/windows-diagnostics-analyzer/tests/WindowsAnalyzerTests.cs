using System.Text;
using SupportBundleAnalyzer.Windows;
using Xunit;

namespace WindowsDiagnosticsAnalyzer.Tests;

public sealed class WindowsAnalyzerTests
{
    [Fact]
    public void ParsesAndClassifiesKnownEventWithoutClaimingRootCause()
    {
        const string xml = """
            <Events xmlns="http://schemas.microsoft.com/win/2004/08/events/event">
              <Event><System><Provider Name="Service Control Manager"/><EventID>7031</EventID><Level>2</Level><TimeCreated SystemTime="2026-01-01T00:00:00Z"/><Channel>System</Channel><Computer>demo-host</Computer></System><EventData><Data>Demo service exited</Data></EventData></Event>
            </Events>
            """;
        using var stream = new MemoryStream(Encoding.UTF8.GetBytes(xml));
        var events = WindowsAnalyzer.ParseEventXml(stream);
        var findings = WindowsAnalyzer.Analyze(events, "events.xml");
        Assert.Single(events);
        Assert.Equal(7031, events[0].EventId);
        Assert.Single(findings);
        Assert.Contains("do not establish root cause", findings[0].Finding.Summary, StringComparison.Ordinal);
    }

    [Fact]
    public void ProhibitsDocumentTypeDeclarations()
    {
        const string xml = "<!DOCTYPE foo [<!ENTITY xxe SYSTEM 'file:///etc/passwd'>]><Event>&xxe;</Event>";
        using var stream = new MemoryStream(Encoding.UTF8.GetBytes(xml));
        Assert.ThrowsAny<Exception>(() => WindowsAnalyzer.ParseEventXml(stream));
    }

    [Fact]
    public void IgnoresUnknownEventIdentifiers()
    {
        var findings = WindowsAnalyzer.Analyze(
            new[] { new WindowsEvent(99999, "Demo", "System", "4", null, "host", "Synthetic") },
            "events.xml");
        Assert.Empty(findings);
    }
}
