package postgres

import (
	"reflect"
	"testing"
)

func TestBindNamedOffset(t *testing.T) {
	type args struct {
		query  string
		params map[string]any
		offset int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   []any
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			args: args{
				query: "SELECT VALUES(:a,:b,:c,:d,:e,:a,:f)",
				params: map[string]any{
					"a": "a",
					"b": "b",
					"c": "c",
					"d": "d",
					"e": "e",
					"f": "f",
				},
				offset: 5,
			},
			want: "SELECT VALUES($6,$7,$8,$9,$10,$6,$11)",
			want1: []any{
				"a", "b", "c", "d", "e", "f",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := BindNamedOffset(tt.args.query, tt.args.params, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("BindNamedOffset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BindNamedOffset() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("BindNamedOffset() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
