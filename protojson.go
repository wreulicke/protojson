// Package protojson provides encoding of protocol buffer messages
// to JSON format, compatible with google.golang.org/protobuf/encoding/protojson
// and supporting io.Writer interface.
package protojson

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

// MarshalOptions configures the marshaling behavior.
// It is compatible with google.golang.org/protobuf/encoding/protojson.MarshalOptions.
type MarshalOptions struct {
	// Indent specifies the set of indentation characters to use in a multiline
	// formatted output such that every entry is preceded by Indent and
	// terminated by a newline. If non-empty, then Multiline is treated as true.
	// Indent can only be composed of space or tab characters.
	Indent string

	// Resolver is used for looking up types when expanding google.protobuf.Any
	// messages. If nil, this defaults to using protoregistry.GlobalTypes.
	Resolver interface {
		FindMessageByName(message protoreflect.FullName) (protoreflect.MessageType, error)
		FindMessageByURL(url string) (protoreflect.MessageType, error)
	}

	// Multiline specifies whether the marshaler should format the output in
	// multiple lines. If false, the entire output will be on a single line.
	Multiline bool

	// AllowPartial allows messages that have missing required fields to marshal
	// without returning an error. If AllowPartial is false (the default),
	// Marshal will return an error if there are any missing required fields.
	AllowPartial bool

	// UseProtoNames uses proto field names instead of lowerCamelCase names
	// in JSON field names.
	UseProtoNames bool

	// UseEnumNumbers emits enum values as numbers instead of strings.
	UseEnumNumbers bool

	// EmitUnpopulated specifies whether to emit unpopulated fields. It does not
	// emit unpopulated oneof fields or unpopulated extension fields.
	// The JSON value emitted for unpopulated fields are as follows:
	//  ╔═══════╤════════════════════════════╗
	//  ║ JSON  │ Protobuf field             ║
	//  ╠═══════╪════════════════════════════╣
	//  ║ false │ proto3 boolean fields      ║
	//  ║ 0     │ proto3 numeric fields      ║
	//  ║ ""    │ proto3 string/bytes fields ║
	//  ║ null  │ proto2 scalar fields       ║
	//  ║ null  │ message fields             ║
	//  ║ []    │ list fields                ║
	//  ║ {}    │ map fields                 ║
	//  ╚═══════╧════════════════════════════╝
	EmitUnpopulated bool

	// EmitDefaultValues specifies whether to emit default-valued fields.
	// It is an alias for EmitUnpopulated for backward compatibility.
	// Deprecated: Use EmitUnpopulated instead.
	EmitDefaultValues bool

	// FieldMaskFunc is called for each field during marshaling to determine
	// if the field value should be masked. If it returns true, the field value
	// will be replaced with "***" in the JSON output.
	//
	// The function receives the FieldDescriptor which can be used to check:
	// - Field name: fd.Name() or fd.JSONName()
	// - Field type: fd.Kind()
	// - Custom options: fd.Options() with proto.GetExtension()
	// - Parent message: fd.ContainingMessage()
	//
	// This allows users to implement custom masking logic based on:
	// - Custom field options (e.g., (mypackage.sensitive) = true)
	// - Field naming patterns (e.g., fields containing "password", "token")
	// - Any other criteria based on the field descriptor
	//
	// If FieldMaskFunc is nil, no masking is performed.
	FieldMaskFunc func(fd protoreflect.FieldDescriptor) bool
}

// Marshal writes the given proto.Message in JSON format using default options.
// Do not depend on the output being stable. It may change over time across
// different versions of the program.
func Marshal(m proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encoder is the internal JSON encoder
type encoder struct {
	w     *bufio.Writer
	opts  MarshalOptions
	depth int
	buf   [64]byte // Scratch buffer for number formatting
}

// marshalMessage marshals a protobuf message to JSON
func (e *encoder) marshalMessage(m protoreflect.Message) error {
	msgDesc := m.Descriptor()

	// Handle well-known types
	if msgDesc.FullName() == "google.protobuf.Timestamp" {
		return e.marshalTimestamp(m)
	}
	if msgDesc.FullName() == "google.protobuf.Duration" {
		return e.marshalDuration(m)
	}
	if msgDesc.FullName() == "google.protobuf.Struct" {
		return e.marshalStruct(m)
	}
	if msgDesc.FullName() == "google.protobuf.Value" {
		return e.marshalValue(m)
	}
	if msgDesc.FullName() == "google.protobuf.ListValue" {
		return e.marshalListValue(m)
	}
	if msgDesc.FullName() == "google.protobuf.Any" {
		return e.marshalAny(m)
	}
	if msgDesc.FullName() == "google.protobuf.Empty" {
		e.w.WriteString("{}")
		return nil
	}

	// Handle wrapper types
	if e.isWrapperType(msgDesc.FullName()) {
		return e.marshalWrapper(m)
	}

	e.w.WriteByte('{')
	e.depth++

	fields := m.Descriptor().Fields()
	first := true

	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)

		// Skip unpopulated fields
		// For optional/oneof fields: skip if not present
		// For regular proto3 fields: skip unless EmitUnpopulated is set
		if !m.Has(fd) {
			if fd.HasPresence() || !e.opts.EmitUnpopulated {
				continue
			}
		}

		if !first {
			e.writeComma()
		}
		first = false

		e.writeIndent()

		// Write field name
		name := e.fieldName(fd)
		e.w.WriteByte('"')
		e.w.WriteString(name)
		e.w.WriteString(`":`)

		// Add space after colon in Multiline or Indent mode
		if e.opts.Multiline || e.opts.Indent != "" {
			e.w.WriteByte(' ')
		}

		// Write field value
		if err := e.marshalField(fd, m.Get(fd)); err != nil {
			return err
		}
	}

	e.depth--
	if !first {
		e.writeIndent()
	}
	e.w.WriteByte('}')

	return nil
}

// fieldName returns the JSON field name for a field descriptor
func (e *encoder) fieldName(fd protoreflect.FieldDescriptor) string {
	if e.opts.UseProtoNames {
		return string(fd.Name())
	}
	return fd.JSONName()
}

// writeIndent writes indentation based on current depth
func (e *encoder) writeComma() {
	e.w.WriteByte(',')
	// Standard library does not add space after comma
}

func (e *encoder) writeColon() {
	e.w.WriteByte(':')
	// Always add one space after colon
	e.w.WriteByte(' ')
}

func (e *encoder) writeIndent() {
	if e.opts.Indent == "" && !e.opts.Multiline {
		return
	}

	e.w.WriteByte('\n')
	indent := e.opts.Indent
	if indent == "" {
		indent = "  "
	}
	for i := 0; i < e.depth; i++ {
		e.w.WriteString(indent)
	}
}

// marshalField marshals a field value
func (e *encoder) marshalField(fd protoreflect.FieldDescriptor, v protoreflect.Value) error {
	if fd.IsList() {
		return e.marshalList(fd, v.List())
	}
	if fd.IsMap() {
		return e.marshalMap(fd, v.Map())
	}
	return e.marshalSingular(fd, v)
}

// marshalSingular marshals a singular field value
func (e *encoder) marshalSingular(fd protoreflect.FieldDescriptor, v protoreflect.Value) error {
	// Check if this field should be masked
	if e.opts.FieldMaskFunc != nil && e.opts.FieldMaskFunc(fd) {
		// Mask string and bytes fields with "***"
		kind := fd.Kind()
		if kind == protoreflect.StringKind || kind == protoreflect.BytesKind {
			e.w.WriteString(`"***"`)
			return nil
		}
		// For other types, fall through to normal processing
		// (user may have set mask condition for non-string/bytes fields)
	}

	switch fd.Kind() {
	case protoreflect.BoolKind:
		if v.Bool() {
			e.w.WriteString("true")
		} else {
			e.w.WriteString("false")
		}
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		b := strconv.AppendInt(e.buf[:0], v.Int(), 10)
		e.w.Write(b)
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		e.w.WriteByte('"')
		b := strconv.AppendInt(e.buf[:0], v.Int(), 10)
		e.w.Write(b)
		e.w.WriteByte('"')
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		b := strconv.AppendUint(e.buf[:0], v.Uint(), 10)
		e.w.Write(b)
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		e.w.WriteByte('"')
		b := strconv.AppendUint(e.buf[:0], v.Uint(), 10)
		e.w.Write(b)
		e.w.WriteByte('"')
	case protoreflect.FloatKind:
		e.marshalFloat32(float32(v.Float()))
	case protoreflect.DoubleKind:
		e.marshalFloat64(v.Float())
	case protoreflect.StringKind:
		e.marshalString(v.String())
	case protoreflect.BytesKind:
		e.w.WriteByte('"')
		encoder := base64.NewEncoder(base64.StdEncoding, e.w)
		encoder.Write(v.Bytes())
		encoder.Close()
		e.w.WriteByte('"')
	case protoreflect.EnumKind:
		if e.opts.UseEnumNumbers {
			b := strconv.AppendInt(e.buf[:0], int64(v.Enum()), 10)
			e.w.Write(b)
		} else {
			enumVal := fd.Enum().Values().ByNumber(v.Enum())
			if enumVal == nil {
				b := strconv.AppendInt(e.buf[:0], int64(v.Enum()), 10)
				e.w.Write(b)
			} else {
				e.w.WriteByte('"')
				e.w.WriteString(string(enumVal.Name()))
				e.w.WriteByte('"')
			}
		}
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return e.marshalMessage(v.Message())
	default:
		return fmt.Errorf("unknown field kind: %v", fd.Kind())
	}
	return nil
}

// marshalFloat32 marshals a float32 value
func (e *encoder) marshalFloat32(f float32) {
	switch {
	case math.IsNaN(float64(f)):
		e.w.WriteString(`"NaN"`)
	case math.IsInf(float64(f), 1):
		e.w.WriteString(`"Infinity"`)
	case math.IsInf(float64(f), -1):
		e.w.WriteString(`"-Infinity"`)
	default:
		b := strconv.AppendFloat(e.buf[:0], float64(f), 'g', -1, 32)
		e.w.Write(b)
	}
}

// marshalFloat64 marshals a float64 value
func (e *encoder) marshalFloat64(f float64) {
	switch {
	case math.IsNaN(f):
		e.w.WriteString(`"NaN"`)
	case math.IsInf(f, 1):
		e.w.WriteString(`"Infinity"`)
	case math.IsInf(f, -1):
		e.w.WriteString(`"-Infinity"`)
	default:
		b := strconv.AppendFloat(e.buf[:0], f, 'g', -1, 64)
		e.w.Write(b)
	}
}

// marshalString marshals a string value with proper escaping
func (e *encoder) marshalString(s string) {
	e.w.WriteByte('"')

	// Fast path: check if escaping is needed
	needsEscape := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < 0x20 || c == '"' || c == '\\' {
			needsEscape = true
			break
		}
	}

	if !needsEscape {
		e.w.WriteString(s)
		e.w.WriteByte('"')
		return
	}

	// Slow path: write with escaping, chunking between special characters
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		var escape string

		switch c {
		case '"':
			escape = `\"`
		case '\\':
			escape = `\\`
		case '\n':
			escape = `\n`
		case '\r':
			escape = `\r`
		case '\t':
			escape = `\t`
		case '\b':
			escape = `\b`
		case '\f':
			escape = `\f`
		default:
			if c < 0x20 {
				escape = fmt.Sprintf(`\u%04x`, c)
			} else {
				continue
			}
		}

		// Write chunk before escape
		if i > start {
			e.w.WriteString(s[start:i])
		}
		e.w.WriteString(escape)
		start = i + 1
	}

	// Write remaining chunk
	if start < len(s) {
		e.w.WriteString(s[start:])
	}

	e.w.WriteByte('"')
}

// marshalList marshals a repeated field
func (e *encoder) marshalList(fd protoreflect.FieldDescriptor, list protoreflect.List) error {
	e.w.WriteByte('[')
	for i := 0; i < list.Len(); i++ {
		if i > 0 {
			e.writeComma()
		}
		if err := e.marshalSingular(fd, list.Get(i)); err != nil {
			return err
		}
	}
	e.w.WriteByte(']')
	return nil
}

// marshalMap marshals a map field
func (e *encoder) marshalMap(fd protoreflect.FieldDescriptor, m protoreflect.Map) error {
	e.w.WriteByte('{')

	// Get key and value field descriptors once
	keyFd := fd.MapKey()
	valFd := fd.MapValue()

	// Sort keys for deterministic output
	// Pre-allocate with capacity to avoid reallocation
	keys := make([]protoreflect.MapKey, 0, m.Len())
	m.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		keys = append(keys, k)
		return true
	})

	slices.SortFunc(keys, func(a, b protoreflect.MapKey) int {
		return strings.Compare(a.String(), b.String())
	})

	// Check key type once
	isStringKey := keyFd.Kind() == protoreflect.StringKind

	for i, k := range keys {
		if i > 0 {
			e.writeComma()
		}

		// Marshal key
		if isStringKey {
			e.marshalString(k.String())
		} else {
			e.w.WriteByte('"')
			e.w.WriteString(k.String())
			e.w.WriteByte('"')
		}

		e.w.WriteByte(':')

		// Marshal value
		if err := e.marshalSingular(valFd, m.Get(k)); err != nil {
			return err
		}
	}

	e.w.WriteByte('}')
	return nil
}

// isWrapperType checks if the given type is a wrapper type
func (e *encoder) isWrapperType(name protoreflect.FullName) bool {
	switch name {
	case "google.protobuf.StringValue",
		"google.protobuf.Int32Value",
		"google.protobuf.Int64Value",
		"google.protobuf.UInt32Value",
		"google.protobuf.UInt64Value",
		"google.protobuf.BoolValue",
		"google.protobuf.FloatValue",
		"google.protobuf.DoubleValue",
		"google.protobuf.BytesValue":
		return true
	}
	return false
}

// marshalWrapper marshals a wrapper type
func (e *encoder) marshalWrapper(m protoreflect.Message) error {
	fd := m.Descriptor().Fields().ByName("value")
	if fd == nil {
		return fmt.Errorf("wrapper type missing value field")
	}
	return e.marshalSingular(fd, m.Get(fd))
}

// marshalTimestamp marshals google.protobuf.Timestamp
func (e *encoder) marshalTimestamp(m protoreflect.Message) error {
	seconds := m.Get(m.Descriptor().Fields().ByName("seconds")).Int()
	nanos := m.Get(m.Descriptor().Fields().ByName("nanos")).Int()

	// Convert to time.Time
	t := time.Unix(seconds, nanos).UTC()

	// Format in RFC 3339 nano format
	e.w.WriteByte('"')
	formatted := t.Format("2006-01-02T15:04:05")

	e.w.WriteString(formatted)

	// Add fractional seconds if nanos > 0
	if nanos > 0 {
		fracStr := fmt.Sprintf(".%09d", nanos)
		// Trim trailing zeros
		fracStr = strings.TrimRight(fracStr, "0")
		e.w.WriteString(fracStr)
	}

	e.w.WriteByte('Z')
	e.w.WriteByte('"')
	return nil
}

// marshalDuration marshals google.protobuf.Duration
func (e *encoder) marshalDuration(m protoreflect.Message) error {
	seconds := m.Get(m.Descriptor().Fields().ByName("seconds")).Int()
	nanos := m.Get(m.Descriptor().Fields().ByName("nanos")).Int()

	e.w.WriteByte('"')
	e.w.WriteString(strconv.FormatInt(seconds, 10))

	if nanos != 0 {
		fracStr := fmt.Sprintf(".%09d", nanos)
		// Trim trailing zeros
		fracStr = strings.TrimRight(fracStr, "0")
		e.w.WriteString(fracStr)
	}

	e.w.WriteByte('s')
	e.w.WriteByte('"')
	return nil
}

// marshalStruct marshals google.protobuf.Struct
func (e *encoder) marshalStruct(m protoreflect.Message) error {
	fields := m.Get(m.Descriptor().Fields().ByName("fields")).Map()

	e.w.WriteByte('{')
	first := true
	fields.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		if !first {
			e.writeComma()
		}
		first = false

		e.marshalString(k.String())
		e.writeColon()
		e.marshalValue(v.Message())
		return true
	})
	e.w.WriteByte('}')
	return nil
}

// marshalValue marshals google.protobuf.Value
func (e *encoder) marshalValue(m protoreflect.Message) error {
	od := m.WhichOneof(m.Descriptor().Oneofs().ByName("kind"))
	if od == nil {
		e.w.WriteString("null")
		return nil
	}

	switch od.Name() {
	case "null_value":
		e.w.WriteString("null")
	case "number_value":
		e.marshalFloat64(m.Get(od).Float())
	case "string_value":
		e.marshalString(m.Get(od).String())
	case "bool_value":
		if m.Get(od).Bool() {
			e.w.WriteString("true")
		} else {
			e.w.WriteString("false")
		}
	case "struct_value":
		return e.marshalStruct(m.Get(od).Message())
	case "list_value":
		return e.marshalListValue(m.Get(od).Message())
	}
	return nil
}

// marshalListValue marshals google.protobuf.ListValue
func (e *encoder) marshalListValue(m protoreflect.Message) error {
	values := m.Get(m.Descriptor().Fields().ByName("values")).List()

	e.w.WriteByte('[')
	for i := 0; i < values.Len(); i++ {
		if i > 0 {
			e.writeComma()
		}
		e.marshalValue(values.Get(i).Message())
	}
	e.w.WriteByte(']')
	return nil
}

// marshalAny marshals google.protobuf.Any
func (e *encoder) marshalAny(m protoreflect.Message) error {
	typeURL := m.Get(m.Descriptor().Fields().ByName("type_url")).String()
	value := m.Get(m.Descriptor().Fields().ByName("value")).Bytes()

	e.w.WriteByte('{')
	e.marshalString("@type")
	e.w.WriteString(": ")
	e.marshalString(typeURL)

	if len(value) > 0 {
		// Try to unmarshal and re-marshal the embedded message
		// For now, we'll just include the type_url
		// A full implementation would need to resolve the type and unmarshal
		resolver := e.opts.Resolver
		if resolver == nil {
			resolver = protoregistry.GlobalTypes
		}

		// Extract message name from type URL
		messageName := protoreflect.FullName(typeURL)
		if i := strings.LastIndexByte(typeURL, '/'); i >= 0 {
			messageName = protoreflect.FullName(typeURL[i+1:])
		}

		if mt, err := resolver.FindMessageByName(messageName); err == nil {
			msg := mt.New()
			if err := proto.Unmarshal(value, msg.Interface()); err == nil {
				// Marshal the embedded message fields
				fields := msg.Descriptor().Fields()
				for i := 0; i < fields.Len(); i++ {
					fd := fields.Get(i)
					if !msg.Has(fd) {
						if fd.HasPresence() || !e.opts.EmitUnpopulated {
							continue
						}
					}

					e.w.WriteString(", ")
					name := e.fieldName(fd)
					e.marshalString(name)
					e.w.WriteString(`: `)
					e.marshalField(fd, msg.Get(fd))
				}
			}
		}
	}

	e.w.WriteByte('}')
	return nil
}

// Encoder writes protocol buffer messages to an output stream in JSON format.
type Encoder struct {
	bw   *bufio.Writer
	opts MarshalOptions
}

// NewEncoder returns a new encoder that writes to w using default options.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		bw:   bufio.NewWriter(w),
		opts: MarshalOptions{},
	}
}

// NewEncoderWithOptions returns a new encoder that writes to w using the
// provided MarshalOptions.
func NewEncoderWithOptions(w io.Writer, opts MarshalOptions) *Encoder {
	return &Encoder{
		bw:   bufio.NewWriter(w),
		opts: opts,
	}
}

// Encode writes the JSON encoding of m to the stream.
// It does not write a newline after the JSON encoding.
func (e *Encoder) Encode(m proto.Message) error {
	opts := e.opts
	if opts.EmitDefaultValues {
		opts.EmitUnpopulated = true
	}

	enc := &encoder{
		w:    e.bw,
		opts: opts,
	}

	if err := enc.marshalMessage(m.ProtoReflect()); err != nil {
		return err
	}

	return e.bw.Flush()
}

// SetOptions updates the MarshalOptions used by the encoder.
func (e *Encoder) SetOptions(opts MarshalOptions) {
	e.opts = opts
}
