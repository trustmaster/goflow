package goflow

import (
	"reflect"
	"testing"
)

func Test_address_kind(t *testing.T) {
	type fields struct {
		proc  string
		port  string
		key   string
		index int
	}

	tests := []struct {
		name   string
		fields fields
		want   portKind
	}{
		{
			name:   "empty",
			fields: fields{proc: "", port: "", key: "", index: noIndex},
			want:   portKindNone,
		},
		{
			name:   "no port name",
			fields: fields{proc: "echo", port: "", key: "", index: noIndex},
			want:   portKindNone,
		},
		{
			name:   "no proc name",
			fields: fields{proc: "", port: "in", key: "", index: noIndex},
			want:   portKindNone,
		},
		{
			name:   "chan port",
			fields: fields{proc: "echo", port: "in", key: "", index: noIndex},
			want:   portKindChan,
		},
		{
			name:   "map port",
			fields: fields{proc: "echo", port: "in", key: "key", index: noIndex},
			want:   portKindMap,
		},
		{
			name:   "array port",
			fields: fields{proc: "echo", port: "in", key: "", index: 10},
			want:   portKindArray,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			a := address{
				proc:  tt.fields.proc,
				port:  tt.fields.port,
				key:   tt.fields.key,
				index: tt.fields.index,
			}
			if got := a.kind(); got != tt.want {
				t.Errorf("address.kind() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_address_String(t *testing.T) {
	type fields struct {
		proc  string
		port  string
		key   string
		index int
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "empty",
			fields: fields{proc: "", port: "", key: "", index: noIndex},
			want:   "<none>",
		},
		{
			name:   "no port name",
			fields: fields{proc: "echo", port: "", key: "", index: noIndex},
			want:   "<none>",
		},
		{
			name:   "no proc name",
			fields: fields{proc: "", port: "in", key: "", index: noIndex},
			want:   "<none>",
		},
		{
			name:   "chan port",
			fields: fields{proc: "echo", port: "in", key: "", index: noIndex},
			want:   "echo.in",
		},
		{
			name:   "map port",
			fields: fields{proc: "echo", port: "in", key: "key", index: noIndex},
			want:   "echo.in[key]",
		},
		{
			name:   "array port",
			fields: fields{proc: "echo", port: "in", key: "", index: 10},
			want:   "echo.in[10]",
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			a := address{
				proc:  tt.fields.proc,
				port:  tt.fields.port,
				key:   tt.fields.key,
				index: tt.fields.index,
			}
			if got := a.String(); got != tt.want {
				t.Errorf("address.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseAddress(t *testing.T) {
	type args struct {
		proc string
		port string
	}

	tests := []struct {
		name string
		args args
		want address
	}{
		{
			name: "empty",
			args: args{proc: "", port: ""},
			want: address{proc: "", port: "", key: "", index: noIndex},
		},
		{
			name: "empty proc",
			args: args{proc: "", port: "in"},
			want: address{proc: "", port: "In", key: "", index: noIndex}, // TODO: this does not look valid, should return an error?
		},
		{
			name: "empty port",
			args: args{proc: "echo", port: ""},
			want: address{proc: "echo", port: "", key: "", index: noIndex}, // TODO: this does not look valid, should return an error?
		},
		{
			name: "chan port",
			args: args{proc: "echo", port: "in"},
			want: address{proc: "echo", port: "In", key: "", index: noIndex},
		},
		{
			name: "map port",
			args: args{proc: "echo", port: "in[key1]"},
			want: address{proc: "echo", port: "In", key: "key1", index: noIndex},
		},
		{
			name: "array port",
			args: args{proc: "echo", port: "in[10]"},
			want: address{proc: "echo", port: "In", key: "10", index: 10}, // TODO: why should key be "10"?
		},
		{
			name: "negative: malformed port",
			args: args{proc: "echo", port: "in[[key1]"},
			want: address{proc: "echo", port: "In[", key: "key1", index: noIndex}, // TODO: return error?
		},
		// TODO: add more negative tests
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			if got := parseAddress(tt.args.proc, tt.args.port); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
