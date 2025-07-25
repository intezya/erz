package erzfiber

import (
	"errors"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/intezya/erz"
	"net/http"
)

const httpOptionsContextKey = "erz_http_options"

func GetHTTPOptions(c *fiber.Ctx) *erz.HTTPOptions {
	raw := c.Locals(httpOptionsContextKey)
	opts, ok := raw.(*erz.HTTPOptions)
	if !ok || opts == nil {
		opts = erz.DefaultHTTPOptions()
	}

	return opts
}

func SetHTTPOptions(c *fiber.Ctx, opts *erz.HTTPOptions) {
	c.Locals(httpOptionsContextKey, opts)
}

func toErzError(err error) erz.Error {
	var erzErr erz.Error
	if !errors.As(err, &erzErr) {
		erzErr = erz.InternalWithCause("Unknown error", err)
	}
	return erzErr
}

func HandleError(c *fiber.Ctx, err error) error {
	erzErr := toErzError(err)
	opts := GetHTTPOptions(c)
	resp := erzErr.ToHTTPResponse(opts)

	return c.Status(erzErr.HTTPStatus()).JSON(resp)
}

func WriteFiberSuccessResponse(c *fiber.Ctx, data interface{}) error {
	opts := GetHTTPOptions(c)
	response := erz.CreateSuccessResponse(data, opts)

	return c.Status(http.StatusOK).JSON(response)
}

type ErrorHandler func(error, *fiber.Ctx) error

func DefaultErrorHandler(err error, c *fiber.Ctx) error {
	erzErr := toErzError(err)
	opts := GetHTTPOptions(c)
	resp := erzErr.ToHTTPResponse(opts)

	return c.Status(erzErr.HTTPStatus()).JSON(resp)
}

func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if recovered := recover(); recovered != nil {
				var err error

				switch v := recovered.(type) {
				case error:
					err = v
				case string:
					err = errors.New(v)
				default:
					err = fmt.Errorf("panic recovered: %v", v)
				}

				erzErr := erz.InternalWithCause("panic recovered", err)
				DefaultErrorHandler(erzErr, c)
			}
		}()

		return c.Next()
	}
}

func ErrorMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			erzErr := toErzError(err)
			return DefaultErrorHandler(erzErr, c)
		}
		return nil
	}
}
