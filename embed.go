package main

import (
	"fmt"
	"html"
	"strings"
)

func badgeColor(nc int) string {
	switch {
	case nc >= 20:
		return "#2ea44f"
	case nc >= 10:
		return "#a3c644"
	case nc >= 5:
		return "#dfb317"
	default:
		return "#e05d44"
	}
}

// renderBadgeSVG returns a shields.io-style SVG: "{chainName} | {nc}".
func renderBadgeSVG(chainName string, nc int) string {
	label := chainName
	value := fmt.Sprintf("%d", nc)

	const charPx = 6
	labelW := len(label)*charPx + 12
	valueW := len(value)*charPx + 12
	if valueW < 24 {
		valueW = 24
	}
	totalW := labelW + valueW

	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
  <title>%s Nakamoto Coefficient: %s</title>
  <linearGradient id="s" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>
  <g clip-path="url(#r)">
    <rect width="%d" height="20" fill="#555"/>
    <rect x="%d" width="%d" height="20" fill="%s"/>
    <rect width="%d" height="20" fill="url(#s)"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
    <text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>`,
		totalW, html.EscapeString(label), value,
		html.EscapeString(label), value,
		totalW,
		labelW,
		labelW, valueW, badgeColor(nc),
		totalW,
		labelW/2, html.EscapeString(label),
		labelW/2, html.EscapeString(label),
		labelW+valueW/2, value,
		labelW+valueW/2, value,
	)
}

type widgetData struct {
	ChainName  string
	ChainToken string
	NC         int
	Prev       int
	Change     int
	BaseURL    string
}

// renderWidgetHTML returns a self-contained embeddable HTML page for an iframe.
func renderWidgetHTML(d widgetData) string {
	arrow := ""
	changeColor := "#6b7280"
	if d.Change > 0 {
		arrow = "▲ "
		changeColor = "#2ea44f"
	} else if d.Change < 0 {
		arrow = "▼ "
		changeColor = "#e05d44"
	}

	changeText := "no change since last update"
	if d.Change != 0 {
		changeText = fmt.Sprintf("%s%d from %d", arrow, abs(d.Change), d.Prev)
	}

	r := strings.NewReplacer(
		"{{NAME}}", html.EscapeString(d.ChainName),
		"{{TOKEN}}", html.EscapeString(d.ChainToken),
		"{{NC}}", fmt.Sprintf("%d", d.NC),
		"{{CHANGE}}", html.EscapeString(changeText),
		"{{BASE}}", html.EscapeString(d.BaseURL),
		"{{VALCOLOR}}", badgeColor(d.NC),
		"{{CHGCOLOR}}", changeColor,
	)

	tpl := `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>{{NAME}} Nakamoto Coefficient</title>
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>
  html,body{margin:0;padding:0;font-family:-apple-system,BlinkMacSystemFont,Segoe UI,Roboto,sans-serif;background:transparent}
  .w{box-sizing:border-box;width:100%;max-width:360px;padding:14px 16px;border-radius:10px;border:1px solid #e5e7eb;background:#fff;color:#111}
  .row{display:flex;align-items:center;justify-content:space-between;gap:12px}
  .label{font-size:12px;color:#6b7280;text-transform:uppercase;letter-spacing:.05em}
  .name{font-size:14px;font-weight:600;margin-top:2px}
  .val{font-size:34px;font-weight:700;line-height:1;color:{{VALCOLOR}}}
  .chg{font-size:12px;margin-top:6px;color:{{CHGCOLOR}}}
  .foot{margin-top:10px;font-size:11px;color:#9ca3af}
  .foot a{color:#4b5563;text-decoration:none}
  .foot a:hover{text-decoration:underline}
  @media (prefers-color-scheme: dark){
    .w{background:#0b0f19;border-color:#1f2937;color:#e5e7eb}
    .label{color:#9ca3af}
    .foot{color:#6b7280}
    .foot a{color:#9ca3af}
  }
</style>
</head>
<body>
  <div class="w">
    <div class="row">
      <div>
        <div class="label">Nakamoto Coefficient</div>
        <div class="name">{{NAME}} ({{TOKEN}})</div>
      </div>
      <div class="val">{{NC}}</div>
    </div>
    <div class="chg">{{CHANGE}}</div>
    <div class="foot">Powered by <a href="{{BASE}}/?chain={{TOKEN}}" target="_blank" rel="noopener">Nakaflow</a></div>
  </div>
</body>
</html>`
	return r.Replace(tpl)
}
