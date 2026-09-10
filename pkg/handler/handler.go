package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"text/template"

	"github.com/labstack/echo/v4"
	uuid "github.com/satori/go.uuid"
)

type key int

const (
	// EchoCtx context identifier for echo
	EchoCtx key = iota
	// Header identifier for echo header
	Header key = iota
	// RequestID gives unique identifier for every incoming request for debugging purpose
	RequestID key = iota
)

// GraphQLHandler handle handler wrapper between go-graphql relay with echo
func GraphQLHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		oriReq := c.Request()

		// all header context
		ctxHeader := http.Header{}

		oriCtx := context.WithValue(oriReq.Context(), Header, ctxHeader)

		requestID := uuid.NewV4().String()
		oriCtx = context.WithValue(oriCtx, RequestID, requestID)

		req := oriReq.WithContext(context.WithValue(oriCtx, EchoCtx, c))

		h.ServeHTTP(c.Response(), req)
		return nil
	}
}

func GraphiQLHandler(c echo.Context) (err error) {
	// Define the data to pass into the template
	docContent, err := os.ReadFile("./web/documentation.graphql")
	if err != nil {
		// Handle the error if the file cannot be read
		fmt.Printf("Failed to read file: %s", err)
		return
	}
	varContent, err := os.ReadFile("./web/variables.json")
	if err != nil {
		// Handle the error if the file cannot be read
		fmt.Printf("Failed to read file: %s", err)
		return
	}

	data := map[string]interface{}{
		"query":     string(docContent),
		"variables": string(varContent),
	}

	// Parse your template file
	tmpl, err := template.ParseFiles("./web/graphiql.html")
	if err != nil {
		c.NoContent(400)
		return
	}

	// Execute the template with data and write the response
	err = tmpl.Execute(c.Response().Writer, data)
	if err != nil {
		c.NoContent(400)
		return
	}
	return
}
