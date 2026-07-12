package pdf

const BrandCSS = `
:root {
  --kore-brand-navy: #1e3a5f;
  --kore-brand-gold: #c9a227;
  --kore-brand-blue: #2b6cb0;
  --kore-text: #1a1f2e;
  --kore-text-muted: #6b7280;
  --kore-border: #e2e6ed;
}
body {
  font-family: Helvetica, Arial, sans-serif;
  font-size: 10pt;
  color: var(--kore-text);
  margin: 0;
  padding: 24px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 2px solid var(--kore-brand-gold);
  padding-bottom: 12px;
  margin-bottom: 24px;
}
.header h1 {
  color: var(--kore-brand-navy);
  font-size: 18pt;
  margin: 0;
}
.header .company {
  text-align: right;
  font-size: 9pt;
  color: var(--kore-text-muted);
}
.title {
  color: var(--kore-brand-navy);
  font-size: 14pt;
  margin: 0 0 16px;
}
table {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
}
th, td {
  border: 1px solid var(--kore-border);
  padding: 8px;
  text-align: left;
}
th {
  background: #f1f3f6;
  color: var(--kore-brand-navy);
  font-size: 9pt;
  text-transform: uppercase;
}
.footer {
  margin-top: 32px;
  padding-top: 12px;
  border-top: 1px solid var(--kore-border);
  font-size: 8pt;
  color: var(--kore-text-muted);
  display: flex;
  justify-content: space-between;
}
.kore-badge {
  color: var(--kore-brand-gold);
  font-weight: bold;
}
`
