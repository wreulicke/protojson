package protojson_test

import (
	"bytes"
	"testing"

	pb "github.com/wreulicke/protojson/gen"
	"github.com/wreulicke/protojson"
	stdprotojson "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Benchmark basic types
func BenchmarkBasicTypes_Custom(b *testing.B) {
	msg := &pb.BasicTypes{
		StringField:   "hello",
		Int32Field:    42,
		Int64Field:    9223372036854775807,
		Uint32Field:   123,
		Uint64Field:   456,
		Sint32Field:   -789,
		Sint64Field:   -1011,
		Fixed32Field:  111,
		Fixed64Field:  222,
		Sfixed32Field: -333,
		Sfixed64Field: -444,
		BoolField:     true,
		FloatField:    3.14,
		DoubleField:   2.718281828,
		BytesField:    []byte("binary data"),
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBasicTypes_Standard(b *testing.B) {
	msg := &pb.BasicTypes{
		StringField:   "hello",
		Int32Field:    42,
		Int64Field:    9223372036854775807,
		Uint32Field:   123,
		Uint64Field:   456,
		Sint32Field:   -789,
		Sint64Field:   -1011,
		Fixed32Field:  111,
		Fixed64Field:  222,
		Sfixed32Field: -333,
		Sfixed64Field: -444,
		BoolField:     true,
		FloatField:    3.14,
		DoubleField:   2.718281828,
		BytesField:    []byte("binary data"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark nested messages
func BenchmarkNestedMessage_Custom(b *testing.B) {
	msg := &pb.Nested{
		Id: "user-123",
		Inner: &pb.Inner{
			Name:  "John Doe",
			Value: 42,
			Deep: &pb.DeepInner{
				Detail: "deep detail",
				Tags:   []string{"tag1", "tag2", "tag3"},
			},
		},
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNestedMessage_Standard(b *testing.B) {
	msg := &pb.Nested{
		Id: "user-123",
		Inner: &pb.Inner{
			Name:  "John Doe",
			Value: 42,
			Deep: &pb.DeepInner{
				Detail: "deep detail",
				Tags:   []string{"tag1", "tag2", "tag3"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark repeated fields
func BenchmarkRepeatedFields_Custom(b *testing.B) {
	msg := &pb.RepeatedFields{
		Strings: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Numbers: []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Bools:   []bool{true, false, true, false, true},
		Doubles: []float64{1.1, 2.2, 3.3, 4.4, 5.5},
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRepeatedFields_Standard(b *testing.B) {
	msg := &pb.RepeatedFields{
		Strings: []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Numbers: []int32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		Bools:   []bool{true, false, true, false, true},
		Doubles: []float64{1.1, 2.2, 3.3, 4.4, 5.5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark map fields
func BenchmarkMapFields_Custom(b *testing.B) {
	msg := &pb.MapFields{
		StringMap: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
			"key4": "value4",
			"key5": "value5",
		},
		IntMap: map[string]int32{
			"key1": 100,
			"key2": 200,
			"key3": 300,
			"key4": 400,
			"key5": 500,
		},
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapFields_Standard(b *testing.B) {
	msg := &pb.MapFields{
		StringMap: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
			"key4": "value4",
			"key5": "value5",
		},
		IntMap: map[string]int32{
			"key1": 100,
			"key2": 200,
			"key3": 300,
			"key4": 400,
			"key5": 500,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with well-known types
func BenchmarkWellKnownTypes_Custom(b *testing.B) {
	msg := &pb.WellKnownTypes{
		Timestamp: timestamppb.New(timestamppb.Now().AsTime()),
		Duration:  durationpb.New(3600000000000),
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWellKnownTypes_Standard(b *testing.B) {
	msg := &pb.WellKnownTypes{
		Timestamp: timestamppb.New(timestamppb.Now().AsTime()),
		Duration:  durationpb.New(3600000000000),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark complex message
func BenchmarkComplexMessage_Custom(b *testing.B) {
	msg := &pb.ComplexMessage{
		Id: "complex-123",
		Users: []*pb.User{
			{
				Id:    "user1",
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  pb.Role_ROLE_ADMIN,
				Permissions: []string{"read", "write", "admin"},
				Metadata: map[string]string{
					"department": "engineering",
					"team":       "backend",
				},
			},
			{
				Id:    "user2",
				Name:  "Jane Smith",
				Email: "jane@example.com",
				Role:  pb.Role_ROLE_USER,
				Permissions: []string{"read", "write"},
				Metadata: map[string]string{
					"department": "sales",
					"team":       "frontend",
				},
			},
		},
		Projects: map[string]*pb.Project{
			"proj1": {
				Id:          "proj1",
				Name:        "Project Alpha",
				Description: "First project",
				Status:      pb.ProjectStatus_PROJECT_STATUS_ACTIVE,
				Tags:        []string{"backend", "api"},
			},
		},
		Settings: &pb.Settings{
			Theme:                "dark",
			NotificationsEnabled: true,
			Language:             "en",
		},
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkComplexMessage_Standard(b *testing.B) {
	msg := &pb.ComplexMessage{
		Id: "complex-123",
		Users: []*pb.User{
			{
				Id:    "user1",
				Name:  "John Doe",
				Email: "john@example.com",
				Role:  pb.Role_ROLE_ADMIN,
				Permissions: []string{"read", "write", "admin"},
				Metadata: map[string]string{
					"department": "engineering",
					"team":       "backend",
				},
			},
			{
				Id:    "user2",
				Name:  "Jane Smith",
				Email: "jane@example.com",
				Role:  pb.Role_ROLE_USER,
				Permissions: []string{"read", "write"},
				Metadata: map[string]string{
					"department": "sales",
					"team":       "frontend",
				},
			},
		},
		Projects: map[string]*pb.Project{
			"proj1": {
				Id:          "proj1",
				Name:        "Project Alpha",
				Description: "First project",
				Status:      pb.ProjectStatus_PROJECT_STATUS_ACTIVE,
				Tags:        []string{"backend", "api"},
			},
		},
		Settings: &pb.Settings{
			Theme:                "dark",
			NotificationsEnabled: true,
			Language:             "en",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with indent option
func BenchmarkWithIndent_Custom(b *testing.B) {
	msg := &pb.BasicTypes{
		StringField: "hello",
		Int32Field:  42,
		BoolField:   true,
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoderWithOptions(&buf, protojson.MarshalOptions{
		Indent: "  ",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWithIndent_Standard(b *testing.B) {
	msg := &pb.BasicTypes{
		StringField: "hello",
		Int32Field:  42,
		BoolField:   true,
	}

	opts := stdprotojson.MarshalOptions{
		Indent: "  ",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := opts.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark with EmitUnpopulated option
func BenchmarkWithEmitUnpopulated_Custom(b *testing.B) {
	msg := &pb.BasicTypes{}

	var buf bytes.Buffer
	encoder := protojson.NewEncoderWithOptions(&buf, protojson.MarshalOptions{
		EmitUnpopulated: true,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWithEmitUnpopulated_Standard(b *testing.B) {
	msg := &pb.BasicTypes{}

	opts := stdprotojson.MarshalOptions{
		EmitUnpopulated: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := opts.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark multiple messages encoding
func BenchmarkMultipleMessages_Custom(b *testing.B) {
	messages := []*pb.BasicTypes{
		{StringField: "msg1", Int32Field: 1},
		{StringField: "msg2", Int32Field: 2},
		{StringField: "msg3", Int32Field: 3},
		{StringField: "msg4", Int32Field: 4},
		{StringField: "msg5", Int32Field: 5},
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for _, msg := range messages {
			if err := encoder.Encode(msg); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkMultipleMessages_Standard(b *testing.B) {
	messages := []*pb.BasicTypes{
		{StringField: "msg1", Int32Field: 1},
		{StringField: "msg2", Int32Field: 2},
		{StringField: "msg3", Int32Field: 3},
		{StringField: "msg4", Int32Field: 4},
		{StringField: "msg5", Int32Field: 5},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, msg := range messages {
			_, err := stdprotojson.Marshal(msg)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// Benchmark large repeated fields
func BenchmarkLargeRepeatedFields_Custom(b *testing.B) {
	strings := make([]string, 1000)
	numbers := make([]int32, 1000)
	for i := 0; i < 1000; i++ {
		strings[i] = "item"
		numbers[i] = int32(i)
	}

	msg := &pb.RepeatedFields{
		Strings: strings,
		Numbers: numbers,
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := encoder.Encode(msg); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLargeRepeatedFields_Standard(b *testing.B) {
	strings := make([]string, 1000)
	numbers := make([]int32, 1000)
	for i := 0; i < 1000; i++ {
		strings[i] = "item"
		numbers[i] = int32(i)
	}

	msg := &pb.RepeatedFields{
		Strings: strings,
		Numbers: numbers,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := stdprotojson.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}
