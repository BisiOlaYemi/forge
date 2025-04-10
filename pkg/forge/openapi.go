package forge

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// OpenAPISpec represents an OpenAPI specification
type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo           `json:"info"`
	Paths      map[string]PathItem   `json:"paths"`
	Components OpenAPIComponents     `json:"components"`
}

// OpenAPIInfo represents the OpenAPI info object
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

// PathItem represents a path in the OpenAPI specification
type PathItem struct {
	Get     *Operation `json:"get,omitempty"`
	Post    *Operation `json:"post,omitempty"`
	Put     *Operation `json:"put,omitempty"`
	Delete  *Operation `json:"delete,omitempty"`
	Patch   *Operation `json:"patch,omitempty"`
}

// Operation represents an operation in the OpenAPI specification
type Operation struct {
	Summary     string                    `json:"summary"`
	Description string                    `json:"description"`
	OperationID string                    `json:"operationId,omitempty"`
	Parameters  []*Parameter              `json:"parameters,omitempty"`
	RequestBody *RequestBody              `json:"requestBody,omitempty"`
	Responses   map[string]*Response      `json:"responses"`
	Tags        []string                  `json:"tags,omitempty"`
	Security    []map[string][]string     `json:"security,omitempty"`
}

// Parameter represents a parameter in the OpenAPI specification
type Parameter struct {
	Name        string      `json:"name"`
	In          string      `json:"in"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Schema      *Schema     `json:"schema"`
}

// RequestBody represents a request body in the OpenAPI specification
type RequestBody struct {
	Description string                      `json:"description"`
	Required    bool                        `json:"required"`
	Content     map[string]MediaTypeObject  `json:"content"`
}

// Response represents a response in the OpenAPI specification
type Response struct {
	Description string                      `json:"description"`
	Content     map[string]MediaTypeObject  `json:"content,omitempty"`
}

// MediaTypeObject represents a media type in the OpenAPI specification
type MediaTypeObject struct {
	Schema *Schema `json:"schema"`
}

// Schema represents a schema in the OpenAPI specification
type Schema struct {
	Type                 string                `json:"type,omitempty"`
	Format              string                `json:"format,omitempty"`
	Description         string                `json:"description,omitempty"`
	Properties          map[string]*Schema    `json:"properties,omitempty"`
	Items               *Schema               `json:"items,omitempty"`
	Required            []string             `json:"required,omitempty"`
	AdditionalProperties *Schema              `json:"additionalProperties,omitempty"`
}

// OpenAPIComponents represents the components section of the OpenAPI specification
type OpenAPIComponents struct {
	Schemas         map[string]*Schema        `json:"schemas"`
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes"`
}

// SecurityScheme represents a security scheme in the OpenAPI specification
type SecurityScheme struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`
	In           string `json:"in,omitempty"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

func getMethodAnnotations(methodName string) []string {
	annotations := make([]string, 0)
	for _, httpMethod := range []string{"Get", "Post", "Put", "Delete", "Patch"} {
		if strings.HasPrefix(methodName, httpMethod) {
			annotations = append(annotations, httpMethod)
		}
	}
	return annotations
}

func getHTTPMethodFromAnnotations(method reflect.Method) string {
	if method.PkgPath != "" {
		return "" // Skip unexported methods
	}

	annotations := getMethodAnnotations(method.Name)
	for _, annotation := range annotations {
		httpMethod := strings.ToUpper(annotation)
		if httpMethod != "" {
			return httpMethod
		}
	}
	return ""
}

// GenerateOpenAPI generates OpenAPI documentation for the application
func (app *Application) GenerateOpenAPI() (*OpenAPISpec, error) {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       app.config.Name,
			Description: app.config.Description,
			Version:     app.config.Version,
		},
		Paths: make(map[string]PathItem),
		Components: OpenAPIComponents{
			Schemas:         make(map[string]*Schema),
			SecuritySchemes: make(map[string]SecurityScheme),
		},
	}

	// Add JWT security scheme
	spec.Components.SecuritySchemes["bearerAuth"] = SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  "JWT token for authentication",
	}

	// Add paths from controllers
	app.mu.RLock()
	defer app.mu.RUnlock()

	for _, controller := range app.controllers {
		controllerType := reflect.TypeOf(controller)
		for i := 0; i < controllerType.NumMethod(); i++ {
			method := controllerType.Method(i)
			httpMethod := getHTTPMethodFromAnnotations(method)
			if httpMethod == "" {
				continue
			}

			// Get path from method name or annotation
			path := fmt.Sprintf("/%s/%s", strings.ToLower(controllerType.Name()), strings.ToLower(method.Name))

			// Create operation
			operation := &Operation{
				Summary:     fmt.Sprintf("%s %s", httpMethod, path),
				OperationID: fmt.Sprintf("%s_%s", strings.ToLower(controllerType.Name()), strings.ToLower(method.Name)),
				Parameters:  make([]*Parameter, 0),
				Responses: map[string]*Response{
					"200": {
						Description: "Successful operation",
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: &Schema{
									Type: "object",
									Properties: map[string]*Schema{
										"success": {Type: "boolean"},
										"data":    {Type: "object"},
									},
								},
							},
						},
					},
				},
			}

			// Add request body if needed
			if httpMethod == "POST" || httpMethod == "PUT" || httpMethod == "PATCH" {
				requestType := getRequestTypeFromMethod(method)
				if requestType != nil {
					operation.RequestBody = &RequestBody{
						Required: true,
						Content: map[string]MediaTypeObject{
							"application/json": {
								Schema: generateSchemaFromType(requestType),
							},
						},
					}
				}
			}

			// Add operation to path
			pathItem, ok := spec.Paths[path]
			if !ok {
				pathItem = PathItem{}
			}

			switch httpMethod {
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

			spec.Paths[path] = pathItem
		}
	}

	return spec, nil
}

// getRequestTypeFromMethod gets the request type from a method
func getRequestTypeFromMethod(method reflect.Method) reflect.Type {
	methodType := method.Type
	if methodType.NumIn() < 3 {
		return nil
	}
	return methodType.In(2)
}

// GenerateSwaggerUI generates a Swagger UI HTML page
func GenerateSwaggerUI(spec *OpenAPISpec) (string, error) {
	specJSON, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Forge API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3/swagger-ui.css">
    <script src="https://unpkg.com/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
</head>
<body>
    <div id="swagger-ui"></div>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                spec: %s,
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
            });
        }
    </script>
</body>
</html>`, string(specJSON)), nil
}

// generateSchemaFromType generates an OpenAPI schema from a Go type
func generateSchemaFromType(t reflect.Type) *Schema {
	if t == nil {
		return nil
	}

	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Struct:
		schema := &Schema{
			Type:       "object",
			Properties: make(map[string]*Schema),
			Required:   make([]string, 0),
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			jsonTag := field.Tag.Get("json")
			if jsonTag == "-" {
				continue
			}

			name := strings.Split(jsonTag, ",")[0]
			if name == "" {
				name = field.Name
			}

			fieldSchema := generateSchemaFromType(field.Type)
			if fieldSchema != nil {
				schema.Properties[name] = fieldSchema
				if !strings.Contains(jsonTag, "omitempty") {
					schema.Required = append(schema.Required, name)
				}
			}
		}
		return schema

	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: generateSchemaFromType(t.Elem()),
		}

	case reflect.Map:
		return &Schema{
			Type: "object",
			AdditionalProperties: generateSchemaFromType(t.Elem()),
		}

	case reflect.Ptr:
		return generateSchemaFromType(t.Elem())
	}

	return nil
} 