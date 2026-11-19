package controller

import (
	"JoinUp/api"
	"JoinUp/internal/controller/dto"
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

const specificationURL = "https://joinup.invalid/openapi.yaml"

type v2Requests struct {
	compiler *jsonschema.Compiler
}

func newV2Requests() *v2Requests {
	var spec map[string]any
	if err := yaml.Unmarshal(api.Specification, &spec); err != nil {
		panic(err)
	}
	data, err := json.Marshal(spec)
	if err != nil {
		panic(err)
	}
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	if err := compiler.AddResource(specificationURL, document); err != nil {
		panic(err)
	}
	return &v2Requests{compiler: compiler}
}

func (r *v2Requests) validate(mediaType, ref string, required bool, invalidStatus, mediaStatus int, headerCredentials ...string) echo.MiddlewareFunc {
	schema := r.compiler.MustCompile(specificationURL + ref)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			request := c.Request()
			if !required {
				for _, header := range headerCredentials {
					if request.Header.Get(header) != "" {
						return next(c)
					}
				}
				if request.Body == http.NoBody {
					return next(c)
				}
			}

			actualMediaType, _, err := mime.ParseMediaType(request.Header.Get(echo.HeaderContentType))
			if err != nil || actualMediaType != mediaType {
				return c.JSON(mediaStatus, dto.Msg{Msg: "expected " + mediaType})
			}

			var value any
			if mediaType == "multipart/form-data" {
				form, err := c.MultipartForm()
				if err != nil {
					return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "invalid multipart form"})
				}
				fields := make(map[string]any, len(form.File)+len(form.Value))
				for name := range form.Value {
					fields[name] = nil
				}
				for name, files := range form.File {
					if _, repeated := fields[name]; repeated || len(files) != 1 {
						fields[name] = nil
					} else {
						fields[name] = files[0].Filename
					}
				}
				value = fields
			} else {
				var body bytes.Buffer
				value, err = jsonschema.UnmarshalJSON(io.TeeReader(request.Body, &body))
				if err != nil {
					return c.JSON(http.StatusBadRequest, dto.Msg{Msg: "expected one JSON document"})
				}
				_ = request.Body.Close()
				request.Body = io.NopCloser(bytes.NewReader(body.Bytes()))
			}
			if err := schema.Validate(value); err != nil {
				return c.JSON(invalidStatus, dto.Msg{Msg: err.Error()})
			}
			return next(c)
		}
	}
}
