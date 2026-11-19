package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"go/token"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

const output = "../internal/controller/routes_v2.gen.go"

type specification struct {
	Servers  []struct{ URL string }
	Security []map[string][]string
	Paths    map[string]map[string]operation
}

type operation struct {
	ID         string `yaml:"operationId"`
	Security   *[]map[string][]string
	Responses  map[string]any
	Parameters []struct {
		Name string
		In   string
	}
	Body struct {
		Required bool
		Content  map[string]struct {
			Schema struct {
				Ref string `yaml:"$ref"`
			}
		}
	} `yaml:"requestBody"`
}

var pathParam = regexp.MustCompile(`\{([^}]+)\}`)

func generate(data []byte) ([]byte, error) {
	var spec specification
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	if len(spec.Servers) != 1 || !strings.HasPrefix(spec.Servers[0].URL, "/") {
		return nil, fmt.Errorf("expected one relative server URL")
	}

	var fields, routes bytes.Buffer
	ids := make(map[string]bool)
	paths := make([]string, 0, len(spec.Paths))
	for path := range spec.Paths {
		paths = append(paths, path)
	}
	slices.Sort(paths)

	for _, path := range paths {
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options", "trace"} {
			op, ok := spec.Paths[path][method]
			if !ok {
				continue
			}
			if !token.IsIdentifier(op.ID) || op.ID == "_" || ids[op.ID] {
				return nil, fmt.Errorf("invalid or duplicate operationId %q for %s %s", op.ID, method, path)
			}
			ids[op.ID] = true
			fmt.Fprintf(&fields, "\t%s echo.HandlerFunc\n", op.ID)
			fmt.Fprintf(&routes, "\tgroup.Add(%q, %q, handlers.%s", strings.ToUpper(method), pathParam.ReplaceAllString(path, ":$1"), op.ID)

			security := spec.Security
			if op.Security != nil {
				security = *op.Security
			}
			if len(security) != 0 {
				scopes, bearer := security[0]["bearerAuth"]
				if len(security) != 1 || len(security[0]) != 1 || !bearer || len(scopes) != 0 {
					return nil, fmt.Errorf("unsupported security on %s", op.ID)
				}
				fmt.Fprint(&routes, ", auth")
			}

			if len(op.Body.Content) > 1 {
				return nil, fmt.Errorf("expected one request media type on %s", op.ID)
			}
			for mediaType, content := range op.Body.Content {
				if content.Schema.Ref == "" {
					return nil, fmt.Errorf("expected a request schema reference on %s", op.ID)
				}
				invalidStatus := 400
				if _, ok := op.Responses["422"]; ok {
					invalidStatus = 422
				}
				mediaStatus := 400
				if _, ok := op.Responses["415"]; ok {
					mediaStatus = 415
				}
				var headerNames []string
				for _, parameter := range op.Parameters {
					if parameter.In == "header" {
						headerNames = append(headerNames, parameter.Name)
					}
				}
				fmt.Fprintf(&routes, ", requests.validate(%q, %q, %t, %d, %d",
					mediaType, content.Schema.Ref, op.Body.Required, invalidStatus, mediaStatus)
				for _, header := range headerNames {
					fmt.Fprintf(&routes, ", %q", header)
				}
				fmt.Fprint(&routes, ")")
			}
			fmt.Fprintln(&routes, ")")
		}
	}

	return format.Source([]byte(fmt.Sprintf(`// Code generated from api/openapi.yaml by go generate. DO NOT EDIT.

package controller

import "github.com/labstack/echo/v5"

const v2Prefix = %q

type v2Operations struct {
%s}

func registerV2(e *echo.Echo, handlers v2Operations, auth echo.MiddlewareFunc, requests *v2Requests) {
	group := e.Group(v2Prefix)
%s}
`, spec.Servers[0].URL, fields.String(), routes.String())))
}

func main() {
	check := flag.Bool("check", false, "check whether the generated routes are current")
	flag.Parse()

	data, err := os.ReadFile("openapi.yaml")
	if err != nil {
		log.Fatal(err)
	}
	generated, err := generate(data)
	if err != nil {
		log.Fatal(err)
	}
	if *check {
		current, err := os.ReadFile(output)
		if err != nil {
			log.Fatal(err)
		}
		if !bytes.Equal(current, generated) {
			log.Fatal("generated routes are outdated: run make api_generate")
		}
		return
	}
	if err := os.WriteFile(output, generated, 0644); err != nil {
		log.Fatal(err)
	}
}
