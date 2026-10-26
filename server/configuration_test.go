package main

import "testing"

func TestMaxLifetimeDays(t *testing.T) {
	cases := []struct {
		in      any
		want    int
		wantErr bool
	}{
		{nil, defaultMaxLifetimeDays, false},
		{float64(3), 3, false},
		{"14", 14, false},
		{"0", 0, true},
		{float64(maxLifetimeDaysCap + 1), 0, true},
		{"x", 0, true},
		{true, 0, true},
	}
	for _, c := range cases {
		got, err := configuration{MaxLifetimeDays: c.in}.maxLifetimeDays()
		if (err != nil) != c.wantErr || got != c.want {
			t.Errorf("%v: got %d, %v", c.in, got, err)
		}
	}
}
