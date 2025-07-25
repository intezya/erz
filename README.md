# erz - Enhanced Error Handling for Go

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.19-blue.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/intezya/erz?status.svg)](https://godoc.org/github.com/intezya/erz)
[![Go Report Card](https://goreportcard.com/badge/github.com/intezya/erz)](https://goreportcard.com/report/github.com/intezya/erz)

`erz` is a comprehensive Go package that transforms error handling with rich context, structured validation errors, automatic stack traces, and seamless HTTP/gRPC integration. Build robust APIs with consistent, user-friendly error responses.

## 🚀 Features

- **Typed Error Codes** - Consistent error classification with predefined codes
- **Rich Validation Errors** - Collect and manage multiple field validation errors
- **Automatic Stack Traces** - Debug errors with precise location information
- **Dual Error Messages** - Technical logs and user-friendly public messages
- **HTTP Integration** - Direct mapping to HTTP status codes (400, 404, 500, etc.)
- **gRPC Support** - Full gRPC status integration with detailed error information
- **Immutable Design** - Safe error modification without side effects
- **Type-Safe Helpers** - Convenient functions for error type checking
- **Validation Collector** - Streamlined validation error aggregation

## 📦 Installation

```bash
go get github.com/intezya/erz
```

## 🎯 Quick Start

### Basic Error Creation

```go
package main

import (
    "fmt"
    "github.com/intezya/erz"
)

func main() {
    // Create common HTTP-mapped errors
    notFoundErr := erz.NotFound("user")           // 404 Not Found
    badRequestErr := erz.InvalidInput("email")    // 400 Bad Request
    serverErr := erz.Internal("database error")   // 500 Internal Server Error
    
    fmt.Println(notFoundErr.Error())      // Technical message for logs
    fmt.Println(notFoundErr.PublicError()) // User-friendly message
    fmt.Println(notFoundErr.HTTPStatus())  // 404
}
```

### Validation Error Collection

```go
func validateUser(user User) error {
    vc := erz.CollectValidationErrors()
    
    if !isValidEmail(user.Email) {
        vc.Add("email", "must be a valid email address", user.Email)
    }
    
    if user.Age < 18 {
        vc.Add("age", "must be at least 18 years old", user.Age)
    }
    
    if len(user.Name) < 2 {
        vc.Add("name", "must be at least 2 characters long", user.Name)
    }
    
    // Return all validation errors as a single error
    if vc.HasErrors() {
        return vc.Error()
    }
    
    return nil
}
```

### Single Validation Errors

```go
// Create individual validation errors
emailErr := erz.ValidationSingle("email", "invalid format", "invalid@example")
ageErr := erz.ValidationSingle("age", "must be positive", -5)
```

## 🔧 Core Interface

The main `Error` interface provides comprehensive error handling capabilities:

```go
type Error interface {
    // Standard error interface
    Error() string
    
    // Core properties
    Code() ErrorCode
    HTTPStatus() int
    GRPCStatus() *status.Status
    PublicError() string
    
    // Error details
    GetValidationErrors() []ValidationError
    GetStackTrace() []StackFrame
    
    // Immutable modifications
    WithDetail(detail string) Error
    WithPublicMessage(msg string) Error
    WithWrapped(err error) Error
    WithValidationError(field, msg string, value any) Error
    WithStackTrace() Error
}
```

## 🌐 HTTP Integration

### Converting Errors to HTTP Status Codes

```go
func handleUserRequest(w http.ResponseWriter, r *http.Request) {
    user, err := getUserByID(userID)
    if err != nil {
        // Automatically get appropriate HTTP status
        status := err.HTTPStatus() // 404, 400, 500, etc.
        http.Error(w, err.PublicError(), status)
        return
    }
    
    // Handle success case...
}
```

### Creating Errors from HTTP Status

```go
err := erz.FromHTTPStatus(404, "user not found")
err := erz.FromHTTPStatus(429, "rate limit exceeded")
```

## 🔌 gRPC Integration

### Converting to gRPC Status

```go
func GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, err := userService.GetUser(req.Id)
    if err != nil {
        // Convert to gRPC status with rich details
        return nil, err.GRPCStatus().Err()
    }
    
    return user, nil
}
```

### Creating Errors from gRPC Status

```go
// Convert gRPC status back to erz.Error
if st, ok := status.FromError(err); ok {
    erzErr := erz.FromGRPCStatusWithDetails(st)
    // Now you have full access to validation errors, stack trace, etc.
}
```

## 🔍 Helper Functions

### Type Checking

```go
if erz.IsNotFound(err) {
    // Handle not found case
}

if erz.IsValidation(err) {
    // Handle validation errors
    validationErrors := erz.GetValidationErrors(err)
    for _, vErr := range validationErrors {
        fmt.Printf("Field: %s, Message: %s, Value: %v\n", 
            vErr.Field, vErr.Message, vErr.Value)
    }
}

if erz.IsInternal(err) {
    // Log internal errors but don't expose details to users
    log.Error("Internal error", "error", err.Error())
}
```

### Stack Trace Extraction

```go
stackFrames := erz.GetStackTrace(err)
for _, frame := range stackFrames {
    fmt.Printf("%s:%d in %s\n", frame.File, frame.Line, frame.Function)
}
```

## 💡 Real-World Examples

### REST API Handler

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        apiErr := erz.InvalidInput("request body").
            WithDetail("invalid JSON format").
            WithPublicMessage("Please provide valid user data")
        
        http.Error(w, apiErr.PublicError(), apiErr.HTTPStatus())
        return
    }
    
    if err := validateUser(user); err != nil {
        http.Error(w, err.PublicError(), err.HTTPStatus())
        return
    }
    
    // Create user...
}
```

### Service Layer with Error Enhancement

```go
func (s *UserService) CreateUser(user User) error {
    if err := s.validateUser(user); err != nil {
        return err // Validation errors bubble up
    }
    
    if err := s.repo.Create(user); err != nil {
        return erz.Internal("failed to create user").
            WithDetail(fmt.Sprintf("database error: %v", err)).
            WithPublicMessage("Unable to create user at this time").
            WithStackTrace()
    }
    
    return nil
}
```

### Complex Validation with Multiple Checks

```go
func validateOrderRequest(req OrderRequest) error {
    vc := erz.CollectValidationErrors()
    
    // Validate items
    if len(req.Items) == 0 {
        vc.Add("items", "at least one item is required", req.Items)
    }
    
    for i, item := range req.Items {
        if item.Quantity <= 0 {
            vc.Add(fmt.Sprintf("items[%d].quantity", i), 
                "must be greater than 0", item.Quantity)
        }
        if item.Price < 0 {
            vc.Add(fmt.Sprintf("items[%d].price", i), 
                "cannot be negative", item.Price)
        }
    }
    
    // Validate shipping
    if req.ShippingAddress == "" {
        vc.Add("shipping_address", "is required", req.ShippingAddress)
    }
    
    return vc.ErrorOrNil() // Returns nil if no errors
}
```

## 📊 Error Code Reference

| Error Code | HTTP Status | Description |
|------------|-------------|-------------|
| `NotFound` | 404 | Resource not found |
| `InvalidInput` | 400 | Bad request/validation error |
| `Unauthorized` | 401 | Authentication required |
| `Forbidden` | 403 | Access denied |
| `Conflict` | 409 | Resource conflict |
| `Internal` | 500 | Internal server error |
| `Validation` | 400 | Field validation errors |

## 🏗️ Architecture Benefits

### Why Choose `erz`?

1. **Consistency** - Standardized error handling across your entire application
2. **Debugging** - Rich context with stack traces and detailed messages
3. **User Experience** - Clean, user-friendly error messages for APIs
4. **Protocol Agnostic** - Works seamlessly with HTTP REST and gRPC services
5. **Type Safety** - Leverage Go's type system for robust error handling
6. **Maintainability** - Immutable design prevents accidental error modification

### Best Practices

- Use validation collectors for complex input validation
- Always include stack traces for internal errors
- Provide meaningful public messages for user-facing errors
- Leverage helper functions for error type checking
- Chain error modifications for rich context

---

**Star ⭐ this repo if you find it helpful!**
