const report = JSON.parse(new TextDecoder().decode(Uint8Array.from(atob(window.__SBA_DATA_B64__), (character) => character.charCodeAt(0))));

const element = (name, text, className) => {
  const node = document.createElement(name);
  if (text !== undefined) node.textContent = String(text);
  if (className) node.className = className;
  return node;
};

const table = (headers, rows) => {
  const result = element('table');
  const head = element('thead');
  const headRow = element('tr');
  headers.forEach((header) => headRow.append(element('th', header)));
  head.append(headRow);
  result.append(head);
  const body = element('tbody');
  rows.forEach((row) => {
    const tableRow = element('tr');
    row.forEach((value) => tableRow.append(element('td', value)));
    body.append(tableRow);
  });
  result.append(body);
  return result;
};

document.querySelectorAll('.tab').forEach((button) => {
  button.addEventListener('click', () => {
    document.querySelectorAll('.tab, .view').forEach((item) => item.classList.remove('active'));
    button.classList.add('active');
    document.getElementById(button.dataset.view).classList.add('active');
  });
});

document.getElementById('theme-toggle').addEventListener('click', () => document.documentElement.classList.toggle('light'));

const severityCount = (severity) => report.findings.filter((finding) => finding.severity === severity).length;
[
  ['Artifacts', report.manifest.artifactCount],
  ['Findings', report.findings.length],
  ['High / critical', severityCount('high') + severityCount('critical')],
  ['Timeline events', report.timeline.length],
  ['Privacy matches', report.sensitive.reduce((total, item) => total + item.count, 0)],
].forEach(([label, value]) => {
  const card = element('div', undefined, 'summary-card');
  card.append(element('strong', value), element('span', label));
  document.getElementById('summary').append(card);
});

const renderFindings = () => {
  const query = document.getElementById('finding-search').value.toLowerCase();
  const severity = document.getElementById('severity-filter').value;
  const container = document.getElementById('finding-list');
  container.replaceChildren();
  report.findings
    .filter((finding) => !severity || finding.severity === severity)
    .filter((finding) => `${finding.title} ${finding.summary} ${finding.component || ''}`.toLowerCase().includes(query))
    .forEach((finding) => {
      const details = element('details', undefined, 'finding');
      const summary = element('summary');
      summary.append(element('span', finding.severity, `badge ${finding.severity}`), element('h3', finding.title));
      details.append(summary, element('p', finding.summary), element('p', finding.explanation));
      const evidenceTitle = element('h4', 'Evidence');
      details.append(evidenceTitle);
      finding.evidence.forEach((evidence) => {
        const location = [evidence.artifact, evidence.lineStart ? `lines ${evidence.lineStart}-${evidence.lineEnd || evidence.lineStart}` : '', evidence.jsonPointer || ''].filter(Boolean).join(' · ');
        details.append(element('p', `${location}\n${evidence.excerpt || ''}`, 'evidence'));
      });
      const next = element('ul');
      finding.nextSteps.forEach((step) => next.append(element('li', step)));
      details.append(element('h4', 'Next investigation steps'), next);
      container.append(details);
    });
};

document.getElementById('finding-search').addEventListener('input', renderFindings);
document.getElementById('severity-filter').addEventListener('change', renderFindings);
renderFindings();

document.getElementById('timeline-list').append(table(
  ['Time (UTC)', 'Severity', 'Component', 'Event', 'Evidence'],
  report.timeline.slice(0, 1000).map((event) => [event.timestamp, event.severity, event.component, event.message, `${event.artifact}:${event.evidence.line || event.evidence.jsonPointer || ''}`]),
));

const renderArtifacts = () => {
  const query = document.getElementById('artifact-search').value.toLowerCase();
  const rows = report.manifest.artifacts
    .filter((artifact) => `${artifact.path} ${artifact.type}`.toLowerCase().includes(query))
    .map((artifact) => [artifact.path, artifact.type, artifact.size, artifact.sha256]);
  document.getElementById('artifact-list').replaceChildren(table(['Path', 'Type', 'Bytes', 'SHA-256'], rows));
};
document.getElementById('artifact-search').addEventListener('input', renderArtifacts);
renderArtifacts();

document.getElementById('privacy-list').append(table(
  ['Artifact', 'Sensitive data type', 'Matches'],
  report.sensitive.map((item) => [item.artifact, item.kind, item.count]),
));

