package main

import (
	"encoding/json"
	"strings"
)

func openAPISpec(baseURL string) ([]byte, error) {
	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "Nakaflow API",
			"description": "Public API for Nakamoto coefficient data across proof-of-stake blockchains.",
			"version":     "1.0.0",
			"contact":     map[string]any{"url": baseURL},
		},
		"servers": []map[string]any{{"url": baseURL}},
		"paths": map[string]any{
			"/naka-coeffs": map[string]any{
				"get": map[string]any{
					"summary":     "Current Nakamoto coefficients for all chains",
					"description": "Returns the latest computed Nakamoto coefficient for every supported chain.",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"last_updated": map[string]any{"type": "string", "format": "date-time"},
											"coefficients": map[string]any{
												"type":  "array",
												"items": map[string]any{"$ref": "#/components/schemas/Coefficient"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/nc-history": map[string]any{
				"get": map[string]any{
					"summary": "Historical Nakamoto coefficient snapshots",
					"parameters": []map[string]any{
						{"name": "chain", "in": "query", "schema": map[string]any{"type": "string"}, "description": "Chain token (e.g. SOL). Omit for all chains."},
						{"name": "days", "in": "query", "schema": map[string]any{"type": "integer"}, "description": "Limit to the last N days. Defaults to 30 when no chain is specified."},
					},
					"responses": map[string]any{
						"200": map[string]any{"description": "OK"},
						"503": map[string]any{"description": "History database unavailable"},
					},
				},
			},
			"/solana-details": map[string]any{
				"get": map[string]any{
					"summary":   "Solana Nakamoto coefficient breakdown",
					"responses": map[string]any{"200": map[string]any{"description": "OK"}},
				},
			},
			"/feed.xml": map[string]any{
				"get": map[string]any{
					"summary":     "RSS feed of current Nakamoto coefficients",
					"description": "RSS 2.0 feed; each item represents the current value for one chain, with changed chains prioritized first.",
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
							"content": map[string]any{
								"application/rss+xml": map[string]any{"schema": map[string]any{"type": "string"}},
							},
						},
					},
				},
			},
			"/embed/badge/{chain}": map[string]any{
				"get": map[string]any{
					"summary": "SVG badge for a chain",
					"parameters": []map[string]any{
						{"name": "chain", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
							"content":     map[string]any{"image/svg+xml": map[string]any{"schema": map[string]any{"type": "string"}}},
						},
					},
				},
			},
			"/embed/widget/{chain}": map[string]any{
				"get": map[string]any{
					"summary": "Embeddable HTML widget for a chain (iframe-ready)",
					"parameters": []map[string]any{
						{"name": "chain", "in": "path", "required": true, "schema": map[string]any{"type": "string"}},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "OK",
							"content":     map[string]any{"text/html": map[string]any{"schema": map[string]any{"type": "string"}}},
						},
					},
				},
			},
		},
		"components": map[string]any{
			"schemas": map[string]any{
				"Coefficient": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"chain_name":           map[string]any{"type": "string"},
						"chain_token":          map[string]any{"type": "string"},
						"naka_co_prev_val":     map[string]any{"type": "integer"},
						"naka_co_curr_val":     map[string]any{"type": "integer"},
						"naka_co_change_val":   map[string]any{"type": "integer"},
					},
				},
			},
		},
	}
	return json.MarshalIndent(spec, "", "  ")
}

func swaggerHTML(baseURL string) string {
	r := strings.NewReplacer("{{BASE}}", baseURL)
	return r.Replace(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Nakaflow API Docs</title>
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
  <style>body{margin:0}</style>
</head>
<body>
  <div id="ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: '{{BASE}}/openapi.json',
      dom_id: '#ui',
      deepLinking: true,
      layout: 'BaseLayout'
    });
  </script>
</body>
</html>`)
}
