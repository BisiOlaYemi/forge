package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

// OpenAPI represents the OpenAPI specification
type OpenAPI struct {
	OpenAPI    string                 `json:"openapi"`
	Info       Info                   `json:"info"`
	Servers    []Server              `json:"servers"`
	Paths      map[string]PathItem   `json:"paths"`
	Components Components            `json:"components"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Server struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
	Options *Operation `json:"options,omitempty"`
	Head    *Operation `json:"head,omitempty"`
}

type Operation struct {
	Tags        []string              `json:"tags"`
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	OperationID string                `json:"operationId"`
	Parameters  []Parameter           `json:"parameters"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Schema      Schema `json:"schema"`
}

type RequestBody struct {
	Description string                `json:"description"`
	Required    bool                  `json:"required"`
	Content     map[string]MediaType `json:"content"`
}

type Response struct {
	Description string                `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

type MediaType struct {
	Schema Schema `json:"schema"`
}

type Schema struct {
	Type       string            `json:"type,omitempty"`
	Properties map[string]Schema `json:"properties,omitempty"`
	Items      *Schema          `json:"items,omitempty"`
	Required   []string         `json:"required,omitempty"`
	Format     string           `json:"format,omitempty"`
	Example    interface{}      `json:"example,omitempty"`
}

type Components struct {
	Schemas    map[string]Schema    `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
}

type SecurityScheme struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	In          string `json:"in,omitempty"`
	Name        string `json:"name,omitempty"`
}

// GenerateOpenAPIDocs generates OpenAPI documentation from the application
func (app *Application) GenerateOpenAPIDocs() error {
	spec := &OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       app.config.Name,
			Description: "API documentation for " + app.config.Name,
			Version:     "1.0.0",
		},
		Servers: []Server{
			{
				URL:         fmt.Sprintf("http://localhost:%d", app.config.Port),
				Description: "Local development server",
			},
		},
		Paths:      make(map[string]PathItem),
		Components: Components{
			Schemas:         make(map[string]Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// Add security schemes
	spec.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
		Type:        "http",
		Description: "JWT Authentication",
		In:          "header",
		Name:        "Authorization",
	}

	// Process controllers
	for _, controller := range app.controllers {
		controllerType := reflect.TypeOf(controller)
		controllerValue := reflect.ValueOf(controller)

		for i := 0; i < controllerType.NumMethod(); i++ {
			method := controllerType.Method(i)
			if !strings.HasSuffix(method.Name, "Action") {
				continue
			}

			// Get route information from annotations
			route := getRouteFromAnnotations(method)
			if route == "" {
				continue
			}

			// Create operation
			operation := &Operation{
				Tags:        []string{strings.TrimSuffix(method.Name, "Action")},
				Summary:     getSummaryFromAnnotations(method),
				Description: getDescriptionFromAnnotations(method),
				OperationID: method.Name,
				Responses: map[string]Response{
					"200": {
						Description: "Successful operation",
						Content: map[string]MediaType{
							"application/json": {
								Schema: Schema{
									Type: "object",
									Properties: map[string]Schema{
										"success": {Type: "boolean"},
										"data":    {Type: "object"},
									},
								},
							},
						},
					},
				},
			}

			// Add security if required
			if isSecureFromAnnotations(method) {
				operation.Security = []map[string][]string{
					{"bearerAuth": {}},
				}
			}

			// Add request body if method is POST/PUT/PATCH
			if isRequestBodyMethod(method.Name) {
				requestType := getRequestTypeFromMethod(controllerValue, method)
				if requestType != nil {
					operation.RequestBody = &RequestBody{
						Required: true,
						Content: map[string]MediaType{
							"application/json": {
								Schema: generateSchemaFromType(requestType),
							},
						},
					}
				}
			}

			// Add path parameters
			params := getPathParameters(route)
			for _, param := range params {
				operation.Parameters = append(operation.Parameters, Parameter{
					Name:        param,
					In:          "path",
					Description: fmt.Sprintf("Parameter %s", param),
					Required:    true,
					Schema:      Schema{Type: "string"},
				})
			}

			// Add to paths
			pathItem := spec.Paths[route]
			switch getHTTPMethodFromAnnotations(method) {
			case "GET":
				pathItem.Get = operation
			case "POST":
				pathItem.Post = operation
			case "PUT":
				pathItem.Put = operation
			case "DELETE":
				pathItem.Delete = operation
			case "PATCH":
				pathItem.Patch = operation
			}
			spec.Paths[route] = pathItem
		}
	}

	// Write to file
	output, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return err
	}

	docsDir := filepath.Join(app.config.Root, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(docsDir, "openapi.json"), output, 0644)
}

// Helper functions for annotation parsing
func getRouteFromAnnotations(method reflect.Method) string {
	// Implementation would parse method annotations for route information
	// This is a placeholder - actual implementation would use reflection
	// to read struct tags or comments
	return ""
}

func getSummaryFromAnnotations(method reflect.Method) string {
	// Implementation would parse method annotations for summary
	return method.Name
}

func getDescriptionFromAnnotations(method reflect.Method) string {
	// Implementation would parse method annotations for description
	return ""
}

func isSecureFromAnnotations(method reflect.Method) bool {
	// Implementation would parse method annotations for security requirements
	return false
}

func isRequestBodyMethod(methodName string) bool {
	method := strings.ToUpper(getHTTPMethodFromAnnotations(reflect.ValueOf(methodName).Method(0)))
	return method == "POST" || method == "PUT" || method == "PATCH"
}

func getHTTPMethodFromAnnotations(method reflect.Method) string {
	// Implementation would parse method annotations for HTTP method
	return "GET"
}

func getRequestTypeFromMethod(controllerValue reflect.Value, method reflect.Method) reflect.Type {
	// Implementation would determine request type from method signature
	return nil
}

func getPathParameters(route string) []string {
	var params []string
	parts := strings.Split(route, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			params = append(params, strings.TrimPrefix(part, ":"))
		}
	}
	return params
}

func generateSchemaFromType(t reflect.Type) Schema {
	schema := Schema{
		Type:       "object",
		Properties: make(map[string]Schema),
		Required:   []string{},
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			continue
		}

		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "-" {
			continue
		}

		fieldSchema := Schema{
			Type: getJSONType(field.Type),
		}

		if field.Tag.Get("validate") != "" {
			schema.Required = append(schema.Required, fieldName)
		}

		schema.Properties[fieldName] = fieldSchema
	}

	return schema
}

func getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "string"
	}
}

func GenerateSwaggerUI(spec *OpenAPISpec) (string, error) {
	tmpl, err := template.New("swagger").Parse(swaggerUITemplate)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, spec)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <meta name="description" content="SwaggerUI" />
    <title>SwaggerUI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui.css" />
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.5.0/swagger-ui-bundle.js" crossorigin></script>
    <script>
        window.onload = () => {
            window.ui = SwaggerUIBundle({
                spec: {{.}},
                dom_id: '#swagger-ui',
            });
        };
    </script>
</body>
</html>` 