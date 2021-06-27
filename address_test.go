package goflow

import "testing"

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
		name    string
		args    args
		want    address
		wantErr bool
	}{
		{name: "empty", args: args{proc: "", port: ""}, want: address{}, wantErr: true},
		{name: "empty proc", args: args{proc: "", port: "in"}, want: address{}, wantErr: true},
		{name: "empty port", args: args{proc: "echo", port: ""}, want: address{}, wantErr: true},
		{name: "malformed port", args: args{proc: "echo", port: "in[[key1]"}, want: address{}, wantErr: true},
		{name: "unmatched opening bracket", args: args{proc: "echo", port: "in[3"}, want: address{}, wantErr: true},
		{name: "unmatched closing bracket", args: args{proc: "echo", port: "in]3"}, want: address{}, wantErr: true},
		{name: "chars after closing bracket", args: args{proc: "echo", port: "in[3]abc"}, want: address{}, wantErr: true},
		{name: "non-UTF-8 in proc", args: args{proc: "echo\xbd", port: "in"}, want: address{}, wantErr: true},
		{name: "non-UTF-8 in port", args: args{proc: "echo", port: "in\xb2"}, want: address{}, wantErr: true},
		{
			name:    "chan port",
			args:    args{proc: "echo", port: "in1"},
			want:    address{proc: "echo", port: "In1", key: "", index: noIndex},
			wantErr: false,
		},
		{
			name:    "non-Latin chan port",
			args:    args{proc: "эхо", port: "ввод"},
			want:    address{proc: "эхо", port: "Ввод", key: "", index: noIndex},
			wantErr: false,
		},
		{
			name:    "map port",
			args:    args{proc: "echo", port: "in[key1]"},
			want:    address{proc: "echo", port: "In", key: "key1", index: noIndex},
			wantErr: false,
		},
		{
			name:    "array port",
			args:    args{proc: "echo", port: "in[10]"},
			want:    address{proc: "echo", port: "In", key: "", index: 10},
			wantErr: false,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAddress(tt.args.proc, tt.args.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("parseAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
