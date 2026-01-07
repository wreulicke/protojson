# protojson

> All implementations in this project were crafted by AI.

Protocol Buffers to JSON encoder for Go that is compatible with the standard `google.golang.org/protobuf/encoding/protojson` package while supporting the `io.Writer` interface for streaming.

## Features

- **Streaming Support**: Encodes directly to `io.Writer` without intermediate buffers
- **Standard Compatible**: Drop-in replacement for `google.golang.org/protobuf/encoding/protojson`
- **Field Masking**: Flexible field-level masking support for sensitive data (e.g., passwords, tokens)
- **Well-Known Types**: Full support for Protocol Buffer well-known types:
  - `google.protobuf.Timestamp`
  - `google.protobuf.Duration`
  - `google.protobuf.Struct`
  - `google.protobuf.Value`
  - `google.protobuf.ListValue`
  - `google.protobuf.Any`
  - `google.protobuf.Empty`
  - Wrapper types (StringValue, Int32Value, etc.)
- **Configurable**: Supports all standard marshaling options
- **Type Safe**: Handles all Protocol Buffer field types correctly

## Installation

```bash
go get github.com/wreulicke/protojson
```

## Usage

```go
import "github.com/wreulicke/protojson"

// Marshal to bytes
data, err := protojson.Marshal(msg)
if err != nil {
    log.Fatal(err)
}

// Stream directly to writer
encoder := protojson.NewEncoder(os.Stdout)
if err := encoder.Encode(msg); err != nil {
    log.Fatal(err)
}
```

### Field Masking

Mask sensitive fields during JSON encoding by providing a custom function that inspects field descriptors:

```go
import (
    "strings"
    "github.com/wreulicke/protojson"
    "google.golang.org/protobuf/reflect/protoreflect"
)

// Mask fields by name pattern
opts := protojson.MarshalOptions{
    FieldMaskFunc: func(fd protoreflect.FieldDescriptor) bool {
        name := string(fd.Name())
        return strings.Contains(name, "password") ||
               strings.Contains(name, "token") ||
               strings.Contains(name, "secret")
    },
}

encoder := protojson.NewEncoderWithOptions(writer, opts)
encoder.Encode(msg) // sensitive fields will be replaced with "***"
```

**Note**: Only `string` and `bytes` fields are masked with `"***"`. Other field types are processed normally even if the mask function returns true.

## License

MIT License. See `LICENSE` file for details.
