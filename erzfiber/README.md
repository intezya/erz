# erzfiber - Fiber Integration for erz

`erzfiber` is a Fiber framework integration package for `erz`, providing seamless error handling, automatic JSON responses, and middleware support for Fiber applications.

## 🚀 Features

- **Automatic Error Conversion** - Convert any error to structured erz errors
- **JSON Response Handling** - Consistent JSON error and success responses
- **HTTP Options Context** - Per-request configuration for error responses
- **Panic Recovery Middleware** - Graceful panic handling with structured errors
- **Error Middleware** - Centralized error handling for all routes
- **Success Response Helper** - Standardized success response format

## 📦 Installation

```bash
go get github.com/intezya/erz/erzfiber
```

## 🎯 Quick Start

### Basic Setup

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/intezya/erz"
    "github.com/intezya/erz/erzfiber"
)

func main() {
    app := fiber.New(fiber.Config{
        ErrorHandler: erzfiber.DefaultErrorHandler,
    })
    
    app.Use(erzfiber.RecoverMiddleware())
    app.Use(erzfiber.ErrorMiddleware())
    
    app.Get("/users/:id", getUserHandler)
    
    app.Listen(":3000")
}
```

### Error Handling in Routes

```go
func getUserHandler(c *fiber.Ctx) error {
    userID := c.Params("id")
    
    user, err := userService.GetUser(userID)
    if err != nil {
        return err // Automatically handled by ErrorMiddleware
    }
    
    return erzfiber.WriteFiberSuccessResponse(c, user)
}

func createUserHandler(c *fiber.Ctx) error {
    var req CreateUserRequest
    if err := c.BodyParser(&req); err != nil {
        return erz.InvalidInput("invalid request body").WithWrapped(err)
    }
    
    user, err := userService.CreateUser(req)
    if err != nil {
        return err
    }
    
    return erzfiber.WriteFiberSuccessResponse(c, user)
}
```

## 🔧 API Reference

### Error Handling Functions

#### HandleError
```go
func HandleError(c *fiber.Ctx, err error) error
```
Converts any error to erz.Error and writes JSON response.

#### DefaultErrorHandler
```go
func DefaultErrorHandler(err error, c *fiber.Ctx) error
```
Default Fiber error handler that can be used in `fiber.Config.ErrorHandler`.

### HTTP Options Management

#### GetHTTPOptions
```go
func GetHTTPOptions(c *fiber.Ctx) *erz.HTTPOptions
```
Retrieves HTTP options from Fiber context locals.

#### SetHTTPOptions
```go
func SetHTTPOptions(c *fiber.Ctx, opts *erz.HTTPOptions)
```
Sets HTTP options in Fiber context locals for the current request.

### Response Helpers

#### WriteFiberSuccessResponse
```go
func WriteFiberSuccessResponse(c *fiber.Ctx, data interface{}) error
```
Writes standardized success response with HTTP 200 status.

### Middleware

#### RecoverMiddleware
```go
func RecoverMiddleware() fiber.Handler
```
Recovers from panics and converts them to structured erz errors.

#### ErrorMiddleware
```go
func ErrorMiddleware() fiber.Handler
```
Catches errors from route handlers and converts them to JSON responses.

## 💡 Usage Examples

### Custom HTTP Options per Route

```go
func sensitiveDataHandler(c *fiber.Ctx) error {
    erzfiber.SetHTTPOptions(c, &erz.HTTPOptions{
        IncludeStackTrace: false,
        IncludeDetails:    false,
    })
    
    data, err := getSecretData()
    if err != nil {
        return erz.Internal("operation failed")
    }
    
    return erzfiber.WriteFiberSuccessResponse(c, data)
}
```

### Validation Error Handling

```go
func validateUserInput(c *fiber.Ctx) error {
    var req UserRequest
    if err := c.BodyParser(&req); err != nil {
        return erz.InvalidInput("invalid JSON").WithWrapped(err)
    }
    
    var validationErrors []erz.ValidationError
    
    if req.Email == "" {
        validationErrors = append(validationErrors, erz.ValidationError{
            Field:   "email",
            Message: "is required",
            Value:   req.Email,
        })
    }
    
    if req.Age < 18 {
        validationErrors = append(validationErrors, erz.ValidationError{
            Field:   "age",
            Message: "must be at least 18",
            Value:   req.Age,
        })
    }
    
    if len(validationErrors) > 0 {
        return erz.ValidationMultiple("validation failed", validationErrors...)
    }
    
    return erzfiber.WriteFiberSuccessResponse(c, "valid")
}
```

### Advanced Error Handling Setup

```go
func setupApp() *fiber.App {
    app := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            erzfiber.SetHTTPOptions(c, &erz.HTTPOptions{
                IncludeStackTrace: isDevelopment(),
                IncludeDetails:    true,
            })
            return erzfiber.DefaultErrorHandler(err, c)
        },
    })
    
    app.Use(erzfiber.RecoverMiddleware())
    
    app.Use(func(c *fiber.Ctx) error {
        erzfiber.SetHTTPOptions(c, &erz.HTTPOptions{
            IncludeStackTrace: false,
            IncludeDetails:    true,
        })
        return c.Next()
    })
    
    app.Use(erzfiber.ErrorMiddleware())
    
    return app
}
```

### Service Layer Integration

```go
type UserService struct {
    repo UserRepository
}

func (s *UserService) GetUser(id string) (*User, error) {
    if id == "" {
        return nil, erz.InvalidInput("user ID is required")
    }
    
    user, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, erz.NotFound("user").WithDetail(fmt.Sprintf("user with ID %s not found", id))
        }
        return nil, erz.Internal("database error").WithWrapped(err)
    }
    
    return user, nil
}

func userController(c *fiber.Ctx) error {
    userService := getUserService(c)
    
    user, err := userService.GetUser(c.Params("id"))
    if err != nil {
        return err
    }
    
    return erzfiber.WriteFiberSuccessResponse(c, user)
}
```

### Custom Error Handler with Logging

```go
func customErrorHandler(err error, c *fiber.Ctx) error {
    erzErr := toErzError(err)
    
    if erz.IsInternal(erzErr) {
        log.Error("Internal server error",
            "path", c.Path(),
            "method", c.Method(),
            "error", erzErr.Error(),
            "detail", erzErr.GetDetail(),
            "stack", erzErr.GetStackTrace(),
        )
    }
    
    return erzfiber.DefaultErrorHandler(erzErr, c)
}
```

## 🔄 Response Format

### Error Response
```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "validation failed",
        "detail": "input validation errors",
        "validation_errors": [
            {
                "field": "email",
                "message": "is required",
                "value": ""
            }
        ]
    }
}
```

### Success Response
```json
{
    "success": true,
    "data": {
        "id": "123",
        "name": "John Doe",
        "email": "john@example.com"
    }
}
```

## 🏗️ Best Practices

1. **Use ErrorMiddleware** - Always add `ErrorMiddleware()` to catch and handle errors consistently
2. **Configure HTTP Options** - Set appropriate options for different environments (dev/prod)
3. **Panic Recovery** - Include `RecoverMiddleware()` to handle unexpected panics gracefully
4. **Service Layer Errors** - Return erz errors from service layer for consistent handling
5. **Logging Internal Errors** - Log internal errors while keeping user responses clean
6. **Validation Patterns** - Use structured validation errors for input validation

## 🔗 Integration with erz

This package works seamlessly with the main `erz` package. All erz error types and helper functions can be used directly in your Fiber handlers, and they will be automatically converted to appropriate HTTP responses.
