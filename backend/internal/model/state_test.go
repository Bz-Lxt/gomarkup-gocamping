package model

import "testing"

func TestTripTransit(t *testing.T) {
	cases := []struct {
		from, to string
		ok       bool
	}{
		{TripDraft, TripActive, true},
		{TripActive, TripPaused, true},
		{TripActive, TripFinished, true},
		{TripPaused, TripActive, true},
		{TripFinished, TripActive, false},
		{TripDraft, TripFinished, false},
	}
	for _, c := range cases {
		if CanTransit(c.from, c.to) != c.ok {
			t.Fatalf("%s -> %s want %v", c.from, c.to, c.ok)
		}
	}
}
