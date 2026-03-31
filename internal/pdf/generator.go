package pdf

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kira-app/kira-server/internal/config"
	"github.com/kira-app/kira-server/internal/domain"
	"github.com/kira-app/kira-server/pkg/currency"
)

type Generator struct {
	cfg    config.StorageConfig
	appURL string
}

func NewGenerator(cfg config.StorageConfig, appURL string) *Generator {
	return &Generator{cfg: cfg, appURL: appURL}
}

type invoiceData struct {
	Invoice  *domain.Invoice
	Project  *domain.Project
	User     *domain.User
	AppURL   string
	Margin   int
	Tax      int
	Formatted formattedAmounts
}

type formattedAmounts struct {
	Subtotal string
	Margin   string
	Tax      string
	Total    string
	Items    []formattedItem
}

type formattedItem struct {
	Name           string
	Severity       string
	GitBranch      string
	HoursFormatted string
	Rate           string
	Amount         string
}

func (g *Generator) GenerateInvoice(inv *domain.Invoice, proj *domain.Project, user *domain.User) (string, error) {
	if err := os.MkdirAll(g.cfg.LocalPath, 0755); err != nil {
		return "", fmt.Errorf("pdf: mkdir: %w", err)
	}

	format := func(cents int) string {
		if inv.Currency == domain.CurrencyUSD {
			return currency.FormatUSD(cents)
		}
		return currency.FormatIDR(cents)
	}

	_, margin, tax := currency.CalcTotal(inv.Subtotal, inv.MarginPct, inv.TaxPct)

	items := make([]formattedItem, len(inv.LineItems))
	for i, li := range inv.LineItems {
		hours := float64(li.ElapsedSeconds) / 3600
		items[i] = formattedItem{
			Name:           li.FeatureName,
			Severity:       li.Severity,
			GitBranch:      li.GitBranch,
			HoursFormatted: fmt.Sprintf("%.2f h", hours),
			Rate:           format(li.HourlyRate) + "/hr",
			Amount:         format(li.Amount),
		}
	}

	data := invoiceData{
		Invoice: inv,
		Project: proj,
		User:    user,
		AppURL:  g.appURL,
		Margin:  margin,
		Tax:     tax,
		Formatted: formattedAmounts{
			Subtotal: format(inv.Subtotal),
			Margin:   format(margin),
			Tax:      format(tax),
			Total:    format(inv.Total),
			Items:    items,
		},
	}

	tmpl, err := template.New("invoice").Funcs(template.FuncMap{
		"formatDate": func(t *time.Time) string {
			if t == nil {
				return "-"
			}
			return t.Format("02 Jan 2006")
		},
	}).Parse(invoiceHTMLTemplate)
	if err != nil {
		return "", fmt.Errorf("pdf: parse template: %w", err)
	}

	filename := fmt.Sprintf("invoice-%s.html", strings.ReplaceAll(inv.InvoiceNumber, "/", "-"))
	filePath := filepath.Join(g.cfg.LocalPath, filename)

	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("pdf: create file: %w", err)
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return "", fmt.Errorf("pdf: execute template: %w", err)
	}

	return "/storage/pdfs/" + filename, nil
}

const invoiceHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Invoice {{.Invoice.InvoiceNumber}}</title>
<style>
  body { font-family: Arial, sans-serif; margin: 40px; color: #1a1a1a; }
  h1 { color: #6366f1; }
  table { width: 100%; border-collapse: collapse; margin-top: 20px; }
  th { background: #6366f1; color: white; padding: 10px; text-align: left; }
  td { padding: 8px 10px; border-bottom: 1px solid #e5e7eb; }
  .totals { margin-top: 20px; text-align: right; }
  .totals table { width: 300px; margin-left: auto; }
  .total-row td { font-weight: bold; font-size: 1.1em; color: #6366f1; }
  .meta { display: flex; justify-content: space-between; margin-bottom: 30px; }
  .badge { display: inline-block; padding: 2px 8px; border-radius: 4px; font-size: 0.8em; }
  .badge-critical { background: #fee2e2; color: #991b1b; }
  .badge-high { background: #fef3c7; color: #92400e; }
  .badge-medium { background: #dbeafe; color: #1e40af; }
  .badge-low { background: #d1fae5; color: #065f46; }
</style>
</head>
<body>
<h1>INVOICE</h1>
<div class="meta">
  <div>
    <strong>{{.Invoice.InvoiceNumber}}</strong><br>
    Status: <strong>{{.Invoice.Status}}</strong><br>
    Due: <strong>{{formatDate .Invoice.DueDate}}</strong>
  </div>
  <div style="text-align:right">
    <strong>{{.Project.Name}}</strong><br>
    {{if .Project.ClientName}}Client: {{.Project.ClientName}}<br>{{end}}
    {{if .Project.ClientEmail}}{{.Project.ClientEmail}}<br>{{end}}
    Issued by: {{.User.Name}}
  </div>
</div>

<table>
  <thead>
    <tr>
      <th>Feature</th>
      <th>Severity</th>
      <th>Branch</th>
      <th>Hours</th>
      <th>Rate</th>
      <th>Amount</th>
    </tr>
  </thead>
  <tbody>
    {{range .Formatted.Items}}
    <tr>
      <td>{{.Name}}</td>
      <td><span class="badge badge-{{.Severity}}">{{.Severity}}</span></td>
      <td>{{.GitBranch}}</td>
      <td>{{.HoursFormatted}}</td>
      <td>{{.Rate}}</td>
      <td>{{.Amount}}</td>
    </tr>
    {{end}}
  </tbody>
</table>

<div class="totals">
  <table>
    <tr><td>Subtotal</td><td>{{.Formatted.Subtotal}}</td></tr>
    {{if .Invoice.MarginPct}}<tr><td>Margin ({{.Invoice.MarginPct}}%)</td><td>{{.Formatted.Margin}}</td></tr>{{end}}
    <tr><td>Tax ({{.Invoice.TaxPct}}%)</td><td>{{.Formatted.Tax}}</td></tr>
    <tr class="total-row"><td>TOTAL</td><td>{{.Formatted.Total}}</td></tr>
  </table>
</div>

{{if .Invoice.Notes}}<p style="margin-top:30px"><strong>Notes:</strong> {{.Invoice.Notes}}</p>{{end}}
</body>
</html>`
