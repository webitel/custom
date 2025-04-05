package data

import (
	"testing"
	"time"
)

func TestCastDateTimeNumber(t *testing.T) {
	type args struct {
		v    float64
		pres time.Duration
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		// TODO: Add test cases.
		{
			name: "nano",
			args: args{
				v:    1731951463.527302496,
				pres: time.Nanosecond, // 1
			},
			want: time.Date(2024, 11, 18, 17, 37, 43, 527302496, time.UTC),
		},
		{
			name: "micro",
			args: args{
				v:    1731951463.527302,
				pres: time.Microsecond, // 1e3
			},
			want: time.Date(2024, 11, 18, 17, 37, 43, 527302000, time.UTC),
		},
		{
			name: "milli",
			args: args{
				v:    1731951463.527,
				pres: time.Millisecond, // 1e6
			},
			want: time.Date(2024, 11, 18, 17, 37, 43, 527000000, time.UTC),
		},
		{
			name: "sec",
			args: args{
				v:    1731951463,
				pres: time.Second, // 1e9
			},
			want: time.Date(2024, 11, 18, 17, 37, 43, 000000000, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date := CastNumberAsDateTime(tt.args.v, tt.args.pres).UTC()
			if !date.Equal(tt.want) { // reflect.DeepEqual(dt, tt.want) {
				t.Errorf("CastNumberAsDateTime() = %v, want %v", date, tt.want)
			}
			num := CastDateTimeAsNumber(date)
			if num != tt.args.v { //  !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CastDateTimeAsNumber() = %f, want %f", num, tt.args.v)
			}
		})
	}
}

func TestDateTimeAsFloat64(t *testing.T) {
	date := time.Now()
	tsec := date.Unix()
	nsec := date.Nanosecond()
	frac := float64(nsec)
	frac = frac / 1e9
	t.Logf("[%s]: %f", date, (float64(tsec) + frac))
}
