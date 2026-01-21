package cli

import (
	"reflect"
	"testing"
)

func Test_parseStringSliceFlag(t *testing.T) {
	type args struct {
		flagName      string
		input         []string
		allowedValues []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "nil == no panic",
			args: args{
				flagName:      "publisher",
				input:         nil,
				allowedValues: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "default test case",
			args: args{
				flagName:      "publisher",
				input:         []string{"dc"},
				allowedValues: []string{"dc"},
			},
			want:    []string{"dc"},
			wantErr: false,
		},
		{
			name: "no duplicates allowed",
			args: args{
				flagName:      "publisher",
				input:         []string{"dc", "dc"},
				allowedValues: []string{"dc"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid value",
			args: args{
				flagName:      "publisher",
				input:         []string{"dc"},
				allowedValues: []string{"Marvel"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "comma separated allowed",
			args: args{
				flagName:      "publisher",
				input:         []string{"dc,marvel"},
				allowedValues: []string{"marvel", "dc"},
			},
			want:    []string{"dc", "marvel"},
			wantErr: false,
		},
		{
			name: "invalid characters",
			args: args{
				flagName:      "publisher",
				input:         []string{"dc%marvel"},
				allowedValues: []string{"Marvel", "dc"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseStringSliceFlag(tt.args.flagName, tt.args.input, tt.args.allowedValues)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseStringSliceFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseStringSliceFlag() got = %v, want %v", got, tt.want)
			}
		})
	}
}
