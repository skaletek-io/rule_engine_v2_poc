// Package api exposes the OpenAPI spec and a Swagger UI handler.
// The spec is embedded at compile time so the binary is self-contained.
package api

import (
	_ "embed"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed openapi.yaml
var openAPISpec []byte

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Skaletek Rule Engine — API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: "{{.SpecURL}}",
      dom_id: "#swagger-ui",
      presets: [SwaggerUIBundle.presets.apis, SwaggerUIBundle.SwaggerUIStandalonePreset],
      layout: "BaseLayout",
      deepLinking: true,
    });
  </script>
</body>
</html>`

var uiTmpl = template.Must(template.New("swagger-ui").Parse(swaggerUIHTML))

// RegisterDocs mounts the following routes on the provided Echo group:
//
//	GET /openapi.yaml  – raw OpenAPI spec (YAML)
//	GET /docs          – Swagger UI
func RegisterDocs(g *echo.Group, specURL string) {
	g.GET("/openapi.yaml", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "application/yaml", openAPISpec)
	})

	g.GET("/docs", func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
		c.Response().WriteHeader(http.StatusOK)
		return uiTmpl.Execute(c.Response().Writer, map[string]string{
			"SpecURL": specURL,
		})
	})
}
