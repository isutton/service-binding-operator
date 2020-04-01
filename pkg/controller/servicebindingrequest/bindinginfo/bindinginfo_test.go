package bindinginfo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewBindingInfo exercises annotation binding information parsing.
func TestNewBindingInfo(t *testing.T) {
	type args struct {
		name  string
		value string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    *BindingInfo
	}{
		{
			args: args{name: "status.configMapRef-password", value: "binding"},
			want: &BindingInfo{
				FieldPath:  "status.configMapRef",
				Descriptor: "binding:password",
				Path:       "password",
				Value:      "binding",
			},
			name:    "{fieldPath}-{path} annotation",
			wantErr: false,
		},
		{
			args: args{name: "status.connectionString", value: "binding"},
			want: &BindingInfo{
				Descriptor: "binding:status.connectionString",
				FieldPath:  "status.connectionString",
				Path:       "status.connectionString",
				Value:      "binding",
			},
			name:    "{path} annotation",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBindingInfo(tt.args.name, tt.args.value)
			if err != nil && !tt.wantErr {
				t.Errorf("NewBindingInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err == nil {
				require.Equal(t, tt.want, b)
			}
		})
	}
}
