# erz - Enhanced Error Handling for Go

`erz` is a comprehensive Go package that transforms error handling with rich context, structured validation errors, automatic stack traces, and seamless HTTP/gRPC integration. Build robust APIs with consistent, user-friendly error responses.

## 🚀 Features

- **Typed Error Codes** - Consistent error classification with predefined codes
- **Rich Validation Errors** - Collect and manage multiple field validation errors
- **Automatic Stack Traces** - Debug errors with precise location information
- **Structured Error Messages** - Separate message and detail fields for flexible error handling
- **HTTP Integration** - Direct mapping to HTTP status codes with JSON response support
- **gRPC Support** - Full gRPC status integration with detailed error information
- **Immutable Design** - Safe error modification without side effects
- **Multiple Error Wrapping** - Chain multiple errors for comprehensive context
- **JSON Serialization** - Built-in JSON output for API responses

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
    err := erz.New(erz.CodeNotFound, "user not found")
    
    fmt.Println(err.Error())         // "user not found"
    fmt.Println(err.GetMessage())    // "user not found"
    fmt.Println(err.HTTPStatus())    // 404
    
    errWithDetail := err.WithDetail("user with ID 123 does not exist")
    fmt.Println(errWithDetail.GetDetail()) // "user with ID 123 does not exist"
}
```

### Validation Errors

```go
func validateUser(user User) error {
    var validationErrors []erz.ValidationError
    
    if !isValidEmail(user.Email) {
        validationErrors = append(validationErrors, erz.ValidationError{
            Field:   "email",
            Message: "must be a valid email address",
            Value:   user.Email,
        })
    }
    
    if user.Age < 18 {
        validationErrors = append(validationErrors, erz.ValidationError{
            Field:   "age", 
            Message: "must be at least 18 years old",
            Value:   user.Age,
        })
    }
    
    if len(validationErrors) > 0 {
        return erz.New(erz.CodeValidation, "validation failed").
            WithValidationErrors(validationErrors...)
    }
    
    return nil
}
```

### Error Wrapping

```go
func processUser(id string) error {
    user, err := database.GetUser(id)
    if err != nil {
        return erz.Wrap(err, erz.CodeInternal, "failed to get user").
            WithDetail(fmt.Sprintf("database query failed for user ID: %s", id))
    }
    
    return nil
}
```

## 🔧 Core Interface

The main `Error` interface provides comprehensive error handling capabilities:

```go
type Error interface {
    erz()
    Error() string
    Code() ErrorCode
    HTTPStatus() int
    GRPCStatus() *status.Status
    GetMessage() string
    GetDetail() string
    GetStackTrace() []StackFrame
    GetValidationErrors() []ValidationError
    WithDetail(detail string) Error
    WithWrapped(err error) Error
    WithValidationErrors(errs ...ValidationError) Error
    WithStackTrace() Error
    Unwrap() error
    ToHTTPResponse(options *HTTPOptions) *HTTPResponse
    AsJSON(options *HTTPOptions) []byte
}
```

## 🌐 HTTP Integration

### Converting Errors to HTTP Responses

```go
func handleUserRequest(w http.ResponseWriter, r *http.Request) {
    user, err := getUserByID(userID)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(err.HTTPStatus())
        w.Write(err.AsJSON(nil))
        return
    }
    
    // Handle success case...
}
```

### Using HTTP Response Helper

```go
func handleError(w http.ResponseWriter, err erz.Error) {
    httpResponse := err.ToHTTPResponse(&erz.HTTPOptions{
        IncludeStackTrace: false,
        IncludeDetails:    true,
    })
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(httpResponse.Status)
    w.Write(httpResponse.Body)
}
```

## 🔌 gRPC Integration

### Converting to gRPC Status

```go
func GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    user, err := userService.GetUser(req.Id)
    if err != nil {
        return nil, err.GRPCStatus().Err()
    }
    
    return user, nil
}
```

## 🔍 Helper Functions

### Type Checking

```go
if erz.IsNotFound(err) {
    // Handle not found case
}

if erz.IsValidation(err) {
    validationErrors := err.GetValidationErrors()
    for _, vErr := range validationErrors {
        fmt.Printf("Field: %s, Message: %s, Value: %v\n", 
            vErr.Field, vErr.Message, vErr.Value)
    }
}

if erz.IsInternal(err) {
    log.Error("Internal error", "error", err.Error(), "detail", err.GetDetail())
}
```

### Stack Trace Access

```go
stackFrames := err.GetStackTrace()
for _, frame := range stackFrames {
    fmt.Printf("%s:%d in %s\n", frame.File, frame.Line, frame.Function)
}
```

### Error Unwrapping

```go
if wrappedErr := errors.Unwrap(err); wrappedErr != nil {
    fmt.Printf("Original error: %v\n", wrappedErr)
}
```

## 💡 Real-World Examples

### REST API Handler

```go
func createUser(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        apiErr := erz.New(erz.CodeInvalidInput, "invalid request body").
            WithDetail("failed to parse JSON").
            WithWrapped(err)
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(apiErr.HTTPStatus())
        w.Write(apiErr.AsJSON(nil))
        return
    }
    
    if err := validateUser(user); err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(err.HTTPStatus())
        w.Write(err.AsJSON(nil))
        return
    }
    
    // Create user...
}
```

### Service Layer with Error Enhancement

```go
func (s *UserService) CreateUser(user User) error {
    if err := s.validateUser(user); err != nil {
        return err
    }
    
    if err := s.repo.Create(user); err != nil {
        return erz.Wrap(err, erz.CodeInternal, "failed to create user").
            WithDetail("database operation failed").
            WithStackTrace()
    }
    
    return nil
}
```

### Complex Validation

```go
func validateOrderRequest(req OrderRequest) error {
    var validationErrors []erz.ValidationError
    
    if len(req.Items) == 0 {
        validationErrors = append(validationErrors, erz.ValidationError{
            Field:   "items",
            Message: "at least one item is required",
            Value:   req.Items,
        })
    }
    
    for i, item := range req.Items {
        if item.Quantity <= 0 {
            validationErrors = append(validationErrors, erz.ValidationError{
                Field:   fmt.Sprintf("items[%d].quantity", i),
                Message: "must be greater than 0",
                Value:   item.Quantity,
            })
        }
    }
    
    if req.ShippingAddress == "" {
        validationErrors = append(validationErrors, erz.ValidationError{
            Field:   "shipping_address",
            Message: "is required",
            Value:   req.ShippingAddress,
        })
    }
    
    if len(validationErrors) > 0 {
        return erz.New(erz.CodeValidation, "validation failed").
            WithValidationErrors(validationErrors...)
    }
    
    return nil
}
```

## 🏗️ Architecture Benefits

### Why Choose `erz`?

1. **Consistency** - Standardized error handling across your entire application
2. **Debugging** - Rich context with stack traces and detailed messages
3. **API-Ready** - Built-in JSON serialization and HTTP response helpers
4. **Protocol Agnostic** - Works seamlessly with HTTP REST and gRPC services
5. **Type Safety** - Leverage Go's type system for robust error handling
6. **Maintainability** - Immutable design prevents accidental error modification
7. **Flexible** - Separate message and detail fields for different use cases

### Best Practices

- Use validation errors for input validation scenarios
- Use `WithDetail()` to add context without changing the main error message
- Leverage `WithWrapped()` to preserve original error information
- Use `AsJSON()` for consistent API error responses
- Chain error modifications for rich context
- Write your custom middlewares for cleaner code

---

**Star ⭐ this repo if you find it helpful!**
