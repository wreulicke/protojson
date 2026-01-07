package protojson_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wreulicke/protojson"
	pb_basic "github.com/wreulicke/protojson/gen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// TestFieldMask tests the Field MaskFunc functionality
func TestFieldMask(t *testing.T) {
	tests := []struct {
		name     string
		msg      *pb_basic.BasicTypes
		maskFunc func(fd protoreflect.FieldDescriptor) bool
		want     string
	}{
		{
			name: "MaskStringField",
			msg: &pb_basic.BasicTypes{
				StringField: "sensitive-data",
				Int32Field:  42,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return string(fd.Name()) == "string_field"
			},
			want: `{"stringField":"***","int32Field":42}`,
		},
		{
			name: "MaskBytesField",
			msg: &pb_basic.BasicTypes{
				StringField: "normal-data",
				BytesField:  []byte("secret-bytes"),
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return string(fd.Name()) == "bytes_field"
			},
			want: `{"stringField":"normal-data","bytesField":"***"}`,
		},
		{
			name: "MaskByFieldNamePattern",
			msg: &pb_basic.BasicTypes{
				StringField: "password123",
				Int32Field:  42,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				name := string(fd.Name())
				return strings.Contains(name, "string")
			},
			want: `{"stringField":"***","int32Field":42}`,
		},
		{
			name: "NoMaskWhenFuncIsNil",
			msg: &pb_basic.BasicTypes{
				StringField: "normal-data",
				Int32Field:  42,
			},
			maskFunc: nil,
			want:     `{"stringField":"normal-data","int32Field":42}`,
		},
		{
			name: "NoMaskWhenFuncReturnsFalse",
			msg: &pb_basic.BasicTypes{
				StringField: "normal-data",
				Int32Field:  42,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return false
			},
			want: `{"stringField":"normal-data","int32Field":42}`,
		},
		{
			name: "MaskMultipleFields",
			msg: &pb_basic.BasicTypes{
				StringField: "secret1",
				BytesField:  []byte("secret2"),
				Int32Field:  42,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				name := string(fd.Name())
				return name == "string_field" || name == "bytes_field"
			},
			want: `{"stringField":"***","int32Field":42,"bytesField":"***"}`,
		},
		{
			name: "MaskDoesNotAffectNonStringBytesFields",
			msg: &pb_basic.BasicTypes{
				StringField: "normal",
				Int32Field:  42,
				BoolField:   true,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				// Try to mask int32 field (should not affect output)
				return string(fd.Name()) == "int32_field"
			},
			want: `{"stringField":"normal","int32Field":42,"boolField":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := protojson.MarshalOptions{
				FieldMaskFunc: tt.maskFunc,
			}
			var buf bytes.Buffer
			enc := protojson.NewEncoderWithOptions(&buf, opts)
			if err := enc.Encode(tt.msg); err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFieldMaskWithOptionalFields tests masking with optional fields
func TestFieldMaskWithOptionalFields(t *testing.T) {
	optionalString := "secret"
	optionalInt := int32(42)

	tests := []struct {
		name     string
		msg      *pb_basic.OptionalFields
		maskFunc func(fd protoreflect.FieldDescriptor) bool
		want     string
	}{
		{
			name: "MaskOptionalString",
			msg: &pb_basic.OptionalFields{
				OptionalString: &optionalString,
				OptionalInt32:  &optionalInt,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return string(fd.Name()) == "optional_string"
			},
			want: `{"optionalString":"***","optionalInt32":42}`,
		},
		{
			name: "MaskUnsetOptionalField",
			msg: &pb_basic.OptionalFields{
				OptionalInt32: &optionalInt,
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return string(fd.Name()) == "optional_string"
			},
			want: `{"optionalInt32":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := protojson.MarshalOptions{
				FieldMaskFunc: tt.maskFunc,
			}
			var buf bytes.Buffer
			enc := protojson.NewEncoderWithOptions(&buf, opts)
			if err := enc.Encode(tt.msg); err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFieldMaskWithEmitUnpopulated tests masking with EmitUnpopulated option
func TestFieldMaskWithEmitUnpopulated(t *testing.T) {
	tests := []struct {
		name     string
		msg      *pb_basic.BasicTypes
		maskFunc func(fd protoreflect.FieldDescriptor) bool
		want     string
	}{
		{
			name: "MaskEmptyStringWithEmitUnpopulated",
			msg:  &pb_basic.BasicTypes{},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return string(fd.Name()) == "string_field"
			},
			want: `{"stringField":"***","int32Field":0,"int64Field":"0","uint32Field":0,"uint64Field":"0","sint32Field":0,"sint64Field":"0","fixed32Field":0,"fixed64Field":"0","sfixed32Field":0,"sfixed64Field":"0","boolField":false,"floatField":0,"doubleField":0,"bytesField":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := protojson.MarshalOptions{
				FieldMaskFunc:   tt.maskFunc,
				EmitUnpopulated: true,
			}
			var buf bytes.Buffer
			enc := protojson.NewEncoderWithOptions(&buf, opts)
			if err := enc.Encode(tt.msg); err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestFieldMaskByKind tests masking based on field kind
func TestFieldMaskByKind(t *testing.T) {
	tests := []struct {
		name     string
		msg      *pb_basic.BasicTypes
		maskFunc func(fd protoreflect.FieldDescriptor) bool
		want     string
	}{
		{
			name: "MaskAllStringFields",
			msg: &pb_basic.BasicTypes{
				StringField: "secret",
				Int32Field:  42,
				BytesField:  []byte("data"),
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return fd.Kind() == protoreflect.StringKind
			},
			want: `{"stringField":"***","int32Field":42,"bytesField":"ZGF0YQ=="}`,
		},
		{
			name: "MaskAllBytesFields",
			msg: &pb_basic.BasicTypes{
				StringField: "normal",
				Int32Field:  42,
				BytesField:  []byte("secret"),
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				return fd.Kind() == protoreflect.BytesKind
			},
			want: `{"stringField":"normal","int32Field":42,"bytesField":"***"}`,
		},
		{
			name: "MaskAllStringAndBytesFields",
			msg: &pb_basic.BasicTypes{
				StringField: "secret1",
				Int32Field:  42,
				BytesField:  []byte("secret2"),
			},
			maskFunc: func(fd protoreflect.FieldDescriptor) bool {
				kind := fd.Kind()
				return kind == protoreflect.StringKind || kind == protoreflect.BytesKind
			},
			want: `{"stringField":"***","int32Field":42,"bytesField":"***"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := protojson.MarshalOptions{
				FieldMaskFunc: tt.maskFunc,
			}
			var buf bytes.Buffer
			enc := protojson.NewEncoderWithOptions(&buf, opts)
			if err := enc.Encode(tt.msg); err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			got := buf.String()
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Encode() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
