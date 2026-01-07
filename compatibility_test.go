package protojson_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"time"

	"github.com/wreulicke/protojson"
	pb_basic "github.com/wreulicke/protojson/gen"
	stdprotojson "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// TestMarshalCompatibility tests that our Marshal implementation produces
// the same output as google.golang.org/protobuf/encoding/protojson
func TestMarshalCompatibility(t *testing.T) {
	tests := []struct {
		name string
		msg  proto.Message
		opts protojson.MarshalOptions
	}{
		{
			name: "BasicTypes_AllFields",
			msg: &pb_basic.BasicTypes{
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
			},
		},
		{
			name: "BasicTypes_Empty",
			msg:  &pb_basic.BasicTypes{},
		},
		{
			name: "BasicTypes_WithEmitUnpopulated",
			msg:  &pb_basic.BasicTypes{},
			opts: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
		},
		{
			name: "OptionalFields_AllSet",
			msg: &pb_basic.OptionalFields{
				OptionalString: proto.String("optional value"),
				OptionalInt32:  proto.Int32(100),
				OptionalBool:   proto.Bool(true),
			},
		},
		{
			name: "OptionalFields_NoneSet",
			msg:  &pb_basic.OptionalFields{},
		},
		{
			name: "OptionalFields_NoneSet_WithEmitUnpopulated",
			msg:  &pb_basic.OptionalFields{},
			opts: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
		},
		{
			name: "EmptyMessage",
			msg:  &pb_basic.EmptyMessage{},
		},
		{
			name: "DefaultValues",
			msg: &pb_basic.DefaultValues{
				EmptyString: "",
				ZeroInt:     0,
				FalseBool:   false,
				EmptyArray:  []string{},
			},
		},
		{
			name: "RepeatedFields_WithValues",
			msg: &pb_basic.RepeatedFields{
				Strings: []string{"a", "b", "c"},
				Numbers: []int32{1, 2, 3, 4, 5},
				Bools:   []bool{true, false, true},
				Doubles: []float64{1.1, 2.2, 3.3},
				BytesList: [][]byte{
					[]byte("data1"),
					[]byte("data2"),
				},
			},
		},
		{
			name: "RepeatedFields_Empty",
			msg:  &pb_basic.RepeatedFields{},
		},
		{
			name: "RepeatedMessages",
			msg: &pb_basic.RepeatedMessages{
				Items: []*pb_basic.Item{
					{Name: "item1", Value: 100},
					{Name: "item2", Value: 200},
					{Name: "item3", Value: 300},
				},
			},
		},
		{
			name: "Nested_Deep",
			msg: &pb_basic.Nested{
				Id: "root",
				Inner: &pb_basic.Inner{
					Name:  "inner",
					Value: 42,
					Deep: &pb_basic.DeepInner{
						Detail: "deep detail",
						Tags:   []string{"tag1", "tag2"},
					},
				},
			},
		},
		{
			name: "Nested_Partial",
			msg: &pb_basic.Nested{
				Id: "root",
				Inner: &pb_basic.Inner{
					Name:  "inner",
					Value: 42,
				},
			},
		},
		{
			name: "MultipleNested",
			msg: &pb_basic.MultipleNested{
				Person: &pb_basic.SimplePerson{
					Name: "John Doe",
					Age:  30,
				},
				Address: &pb_basic.SimpleAddress{
					Street:  "123 Main St",
					City:    "Tokyo",
					Country: "Japan",
				},
				Contact: &pb_basic.SimpleContactInfo{
					Email: "john@example.com",
					Phone: "+81-90-1234-5678",
				},
			},
		},
		{
			name: "MapFields_StringMap",
			msg: &pb_basic.MapFields{
				StringMap: map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
		},
		{
			name: "MapFields_IntMap",
			msg: &pb_basic.MapFields{
				IntMap: map[string]int32{
					"one":   1,
					"two":   2,
					"three": 3,
				},
			},
		},
		{
			name: "MapFields_MessageMap",
			msg: &pb_basic.MapFields{
				MessageMap: map[string]*pb_basic.Value{
					"first":  {Data: "first data", Count: 10},
					"second": {Data: "second data", Count: 20},
				},
			},
		},
		{
			name: "MapFields_Mixed",
			msg: &pb_basic.MapFields{
				StringMap: map[string]string{"a": "A"},
				IntMap:    map[string]int32{"b": 2},
				BoolMap:   map[string]bool{"c": true},
				IntKeyMap: map[int32]string{1: "one", 2: "two"},
				MessageMap: map[string]*pb_basic.Value{
					"msg": {Data: "data", Count: 5},
				},
			},
		},
		{
			name: "EnumFields",
			msg: &pb_basic.EnumFields{
				Status:   pb_basic.Status_STATUS_ACTIVE,
				Priority: pb_basic.Priority_PRIORITY_HIGH,
			},
		},
		{
			name: "EnumFields_WithUseEnumNumbers",
			msg: &pb_basic.EnumFields{
				Status:   pb_basic.Status_STATUS_ACTIVE,
				Priority: pb_basic.Priority_PRIORITY_HIGH,
			},
			opts: protojson.MarshalOptions{
				UseEnumNumbers: true,
			},
		},
		{
			name: "RepeatedEnums",
			msg: &pb_basic.RepeatedEnums{
				Statuses: []pb_basic.Status{
					pb_basic.Status_STATUS_ACTIVE,
					pb_basic.Status_STATUS_INACTIVE,
					pb_basic.Status_STATUS_PENDING,
				},
				Priorities: []pb_basic.Priority{
					pb_basic.Priority_PRIORITY_LOW,
					pb_basic.Priority_PRIORITY_HIGH,
				},
			},
		},
		{
			name: "OneOfFields_StringValue",
			msg: &pb_basic.OneOfFields{
				Id:    "test",
				Value: &pb_basic.OneOfFields_StringValue{StringValue: "hello"},
			},
		},
		{
			name: "OneOfFields_IntValue",
			msg: &pb_basic.OneOfFields{
				Id:    "test",
				Value: &pb_basic.OneOfFields_IntValue{IntValue: 42},
			},
		},
		{
			name: "OneOfFields_MessageValue",
			msg: &pb_basic.OneOfFields{
				Id:    "test",
				Value: &pb_basic.OneOfFields_MessageValue{MessageValue: &pb_basic.Message{Content: "content"}},
			},
		},
		{
			name: "OneOfFields_NoValueSet",
			msg: &pb_basic.OneOfFields{
				Id: "test",
			},
		},
		{
			name: "WellKnownTypes_Timestamp",
			msg: &pb_basic.WellKnownTypes{
				Timestamp: timestamppb.New(time.Unix(1609459200, 0)), // 2021-01-01 00:00:00 UTC
			},
		},
		{
			name: "WellKnownTypes_Duration",
			msg: &pb_basic.WellKnownTypes{
				Duration: durationpb.New(3600 * time.Second), // 1 hour
			},
		},
		{
			name: "WrapperTypes_AllSet",
			msg: &pb_basic.WrapperTypes{
				StringValue: wrapperspb.String("wrapped string"),
				Int32Value:  wrapperspb.Int32(42),
				Int64Value:  wrapperspb.Int64(9223372036854775807),
				Uint32Value: wrapperspb.UInt32(123),
				Uint64Value: wrapperspb.UInt64(456),
				BoolValue:   wrapperspb.Bool(true),
				FloatValue:  wrapperspb.Float(3.14),
				DoubleValue: wrapperspb.Double(2.718281828),
				BytesValue:  wrapperspb.Bytes([]byte("wrapped bytes")),
			},
		},
		{
			name: "WrapperTypes_NullValues",
			msg:  &pb_basic.WrapperTypes{},
		},
		{
			name: "EdgeCases_Unicode",
			msg: &pb_basic.EdgeCases{
				UnicodeString: "日本語テスト",
				EmojiString:   "😀🎉🚀",
				SpecialChars:  "Special: \n\t\r\"\\",
			},
		},
		{
			name: "EdgeCases_LargeNumbers",
			msg: &pb_basic.EdgeCases{
				LargeInt64:  9223372036854775807,
				LargeUint64: 18446744073709551615,
			},
		},
		{
			name: "DeepNesting",
			msg: &pb_basic.DeepNesting{
				Level1: &pb_basic.Level1{
					Data: "level1",
					Level2: &pb_basic.Level2{
						Data: "level2",
						Level3: &pb_basic.Level3{
							Data: "level3",
							Level4: &pb_basic.Level4{
								Data: "level4",
								Level5: &pb_basic.Level5{
									Data:  "level5",
									Items: []string{"a", "b", "c"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ComplexMessage",
			msg: &pb_basic.ComplexMessage{
				Id: "complex-1",
				Users: []*pb_basic.User{
					{
						Id:          "user-1",
						Name:        "Alice",
						Email:       "alice@example.com",
						Role:        pb_basic.Role_ROLE_ADMIN,
						Permissions: []string{"read", "write", "delete"},
						Profile: &pb_basic.Profile{
							AvatarUrl: "https://example.com/avatar.jpg",
							Bio:       "Software Engineer",
							Address: &pb_basic.Address{
								Street:     "123 Main St",
								City:       "Tokyo",
								Country:    "Japan",
								PostalCode: "100-0001",
								Location: &pb_basic.Location{
									Latitude:  35.6762,
									Longitude: 139.6503,
								},
							},
							SocialLinks: []*pb_basic.SocialLink{
								{Platform: "twitter", Url: "https://twitter.com/alice"},
								{Platform: "github", Url: "https://github.com/alice"},
							},
						},
						Metadata: map[string]string{
							"department": "engineering",
							"team":       "backend",
						},
					},
				},
				Projects: map[string]*pb_basic.Project{
					"proj-1": {
						Id:          "proj-1",
						Name:        "Project Alpha",
						Description: "First project",
						Status:      pb_basic.ProjectStatus_PROJECT_STATUS_ACTIVE,
						Tasks: []*pb_basic.Task{
							{
								Id:          "task-1",
								Title:       "Implement feature",
								Description: "Add new feature",
								Priority:    pb_basic.TaskPriority_TASK_PRIORITY_HIGH,
								AssignedTo:  "user-1",
								Labels:      []string{"feature", "backend"},
								Result:      &pb_basic.Task_SuccessMessage{SuccessMessage: "Completed successfully"},
							},
						},
						Tags:      []string{"important", "active"},
						CreatedAt: timestamppb.New(time.Unix(1609459200, 0)),
					},
				},
				Settings: &pb_basic.Settings{
					NotificationsEnabled: true,
					Theme:                "dark",
					Language:             "ja",
					Features: map[string]*pb_basic.FeatureFlag{
						"new_ui": {
							Enabled:     true,
							Description: "Enable new UI",
							Config: map[string]string{
								"version": "2.0",
							},
						},
					},
					Preferences: &pb_basic.Preferences{
						ItemsPerPage:     50,
						DateFormat:       "YYYY-MM-DD",
						TimeZone:         "Asia/Tokyo",
						FavoriteProjects: []string{"proj-1", "proj-2"},
					},
				},
				CreatedAt: timestamppb.New(time.Unix(1609459200, 0)),
				UpdatedAt: timestamppb.New(time.Unix(1609459200, 0)),
			},
		},
		{
			name: "UseProtoNames",
			msg: &pb_basic.JsonNaming{
				SnakeCaseField:       "snake",
				CamelCaseField:       "camel",
				PascalCaseField:      "pascal",
				FieldWith_123Numbers: "numbers",
				SCREAMING_SNAKE_CASE: "screaming",
			},
			opts: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		},
		{
			name: "Multiline",
			msg: &pb_basic.BasicTypes{
				StringField: "hello",
				Int32Field:  42,
				BoolField:   true,
			},
			opts: protojson.MarshalOptions{
				Multiline: true,
			},
		},
		{
			name: "Indent",
			msg: &pb_basic.BasicTypes{
				StringField: "hello",
				Int32Field:  42,
				BoolField:   true,
			},
			opts: protojson.MarshalOptions{
				Indent: "  ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal with google.golang.org/protobuf/encoding/protojson
			stdOpts := stdprotojson.MarshalOptions{
				Multiline:       tt.opts.Multiline,
				Indent:          tt.opts.Indent,
				AllowPartial:    tt.opts.AllowPartial,
				UseProtoNames:   tt.opts.UseProtoNames,
				UseEnumNumbers:  tt.opts.UseEnumNumbers,
				EmitUnpopulated: tt.opts.EmitUnpopulated,
			}
			expectedJSON, err := stdOpts.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("standard protojson.Marshal failed: %v", err)
			}

			// Marshal with our implementation
			var gotBuf bytes.Buffer
			encoder := protojson.NewEncoderWithOptions(&gotBuf, tt.opts)
			if err := encoder.Encode(tt.msg); err != nil {
				t.Fatalf("our protojson.Encode failed: %v", err)
			}
			gotJSON := gotBuf.Bytes()

			// Compare JSON outputs
			if diff := cmp.Diff(string(expectedJSON), string(gotJSON)); diff != "" {
				t.Errorf("Marshal() output mismatch (-want +got):\n%s", diff)
				t.Logf("Expected JSON: %s", expectedJSON)
				t.Logf("Got JSON: %s", gotJSON)
			}
		})
	}
}

// TestEncoderCompatibility tests that our Encoder implementation produces
// the same output as repeated calls to google.golang.org/protobuf/encoding/protojson.Marshal
func TestEncoderCompatibility(t *testing.T) {
	messages := []proto.Message{
		&pb_basic.BasicTypes{StringField: "first", Int32Field: 1},
		&pb_basic.BasicTypes{StringField: "second", Int32Field: 2},
		&pb_basic.BasicTypes{StringField: "third", Int32Field: 3},
	}

	// Expected output: marshal each message separately
	var expectedBuf bytes.Buffer
	for _, msg := range messages {
		data, err := stdprotojson.Marshal(msg)
		if err != nil {
			t.Fatalf("standard protojson.Marshal failed: %v", err)
		}
		expectedBuf.Write(data)
	}

	// Our encoder
	var gotBuf bytes.Buffer
	encoder := protojson.NewEncoder(&gotBuf)
	for _, msg := range messages {
		if err := encoder.Encode(msg); err != nil {
			t.Fatalf("Encoder.Encode failed: %v", err)
		}
	}

	if diff := cmp.Diff(expectedBuf.String(), gotBuf.String()); diff != "" {
		t.Errorf("Encoder output mismatch (-want +got):\n%s", diff)
	}
}

// TestEncoderWithOptions tests Encoder with various MarshalOptions
func TestEncoderWithOptions(t *testing.T) {
	tests := []struct {
		name string
		msg  proto.Message
		opts protojson.MarshalOptions
	}{
		{
			name: "DefaultOptions",
			msg: &pb_basic.BasicTypes{
				StringField: "test",
				Int32Field:  42,
			},
		},
		{
			name: "WithIndent",
			msg: &pb_basic.BasicTypes{
				StringField: "test",
				Int32Field:  42,
			},
			opts: protojson.MarshalOptions{
				Indent: "  ",
			},
		},
		{
			name: "WithEmitUnpopulated",
			msg:  &pb_basic.BasicTypes{},
			opts: protojson.MarshalOptions{
				EmitUnpopulated: true,
			},
		},
		{
			name: "WithUseProtoNames",
			msg: &pb_basic.BasicTypes{
				StringField: "test",
				Int32Field:  42,
			},
			opts: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Standard encoder
			var expectedBuf bytes.Buffer
			stdOpts := stdprotojson.MarshalOptions{
				Multiline:       tt.opts.Multiline,
				Indent:          tt.opts.Indent,
				AllowPartial:    tt.opts.AllowPartial,
				UseProtoNames:   tt.opts.UseProtoNames,
				UseEnumNumbers:  tt.opts.UseEnumNumbers,
				EmitUnpopulated: tt.opts.EmitUnpopulated,
			}
			expectedData, err := stdOpts.Marshal(tt.msg)
			if err != nil {
				t.Fatalf("standard protojson.Marshal failed: %v", err)
			}
			expectedBuf.Write(expectedData)

			// Our encoder with options
			var gotBuf bytes.Buffer
			encoder := protojson.NewEncoderWithOptions(&gotBuf, tt.opts)
			if err := encoder.Encode(tt.msg); err != nil {
				t.Fatalf("Encoder.Encode failed: %v", err)
			}

			if diff := cmp.Diff(expectedBuf.String(), gotBuf.String()); diff != "" {
				t.Errorf("Encoder output mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

// TestEncoderSetOptions tests that SetOptions updates the encoder's options
func TestEncoderSetOptions(t *testing.T) {
	msg := &pb_basic.BasicTypes{
		StringField: "test",
		Int32Field:  42,
	}

	var buf bytes.Buffer
	encoder := protojson.NewEncoder(&buf)

	// First encode with default options
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("First Encode failed: %v", err)
	}

	// Change options to use indentation
	encoder.SetOptions(protojson.MarshalOptions{
		Indent: "  ",
	})

	// Clear buffer and encode again
	buf.Reset()
	if err := encoder.Encode(msg); err != nil {
		t.Fatalf("Second Encode failed: %v", err)
	}

	// Verify output has indentation
	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("  ")) {
		t.Errorf("Expected indented output after SetOptions, got: %s", output)
	}

	// Compare with expected output
	stdOpts := stdprotojson.MarshalOptions{
		Indent: "  ",
	}
	expected, err := stdOpts.Marshal(msg)
	if err != nil {
		t.Fatalf("standard protojson.Marshal failed: %v", err)
	}

	if diff := cmp.Diff(string(expected), output); diff != "" {
		t.Errorf("Output after SetOptions mismatch (-want +got):\n%s", diff)
	}
}

// TestEncoderMultipleMessages tests encoding multiple different message types
func TestEncoderMultipleMessages(t *testing.T) {
	messages := []proto.Message{
		&pb_basic.BasicTypes{StringField: "basic", Int32Field: 1},
		&pb_basic.EnumFields{Status: pb_basic.Status_STATUS_ACTIVE},
		&pb_basic.RepeatedFields{Strings: []string{"a", "b", "c"}},
	}

	// Expected output
	var expectedBuf bytes.Buffer
	for _, msg := range messages {
		data, err := stdprotojson.Marshal(msg)
		if err != nil {
			t.Fatalf("standard protojson.Marshal failed: %v", err)
		}
		expectedBuf.Write(data)
	}

	// Our encoder
	var gotBuf bytes.Buffer
	encoder := protojson.NewEncoder(&gotBuf)
	for _, msg := range messages {
		if err := encoder.Encode(msg); err != nil {
			t.Fatalf("Encoder.Encode failed: %v", err)
		}
	}

	if diff := cmp.Diff(expectedBuf.String(), gotBuf.String()); diff != "" {
		t.Errorf("Encoder output mismatch (-want +got):\n%s", diff)
	}
}
