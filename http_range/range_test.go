package http_range

import (
	"reflect"
	"testing"
)

func TestParseRange(t *testing.T) {
	type args struct {
		s    string
		size int64
	}
	tests := []struct {
		name    string
		args    args
		want    []Range
		wantErr bool
	}{
		{
			name: "blank",
		},
		{
			name: "invalid",
			args: args{
				s:    "keks=100500",
				size: 100,
			},
			wantErr: true,
		},
		{
			name: "invalid single value",
			args: args{
				s:    "bytes=200",
				size: 500,
			},
			wantErr: true,
		},
		{
			name: "invalid non-digit end",
			args: args{
				s:    "bytes=-f",
				size: 500,
			},
			wantErr: true,
		},
		{
			name: "invalid no start or end",
			args: args{
				s:    "bytes=-",
				size: 500,
			},
			wantErr: true,
		},
		{
			name: "invalid non-digit start",
			args: args{
				s:    "bytes=f-",
				size: 500,
			},
			wantErr: true,
		},
		{
			name: "single",
			args: args{
				s: "bytes=100-200", size: 200,
			},
			want: []Range{
				{
					Start:  100,
					Length: 100,
				},
			},
		},
		{
			name: "no overlap",
			args: args{
				s: "bytes=100-50", size: 200,
			},
			wantErr: true,
		},
		{
			name: "after end",
			args: args{
				s: "bytes=200-250", size: 200,
			},
			wantErr: true,
		},
		{
			name: "from offset till end",
			args: args{
				s: "bytes=50-", size: 200,
			},
			want: []Range{
				{
					Start:  50,
					Length: 150,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRange(tt.args.s, tt.args.size)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRange() got = %v, want %v", got, tt.want)
			}
		})
	}
}
