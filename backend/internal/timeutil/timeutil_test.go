package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBeijingOffset(t *testing.T) {
	n := Now()
	_, off := n.Zone()
	require.Equal(t, 8*3600, off)
	s := FormatDisplay(n)
	require.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, s)
}

func TestParseISO(t *testing.T) {
	raw := "2026-08-24 08:00:00"
	tt, err := ParseISO(raw)
	require.NoError(t, err)
	require.Equal(t, 8, tt.Hour())
	y, m, d := CivilDate(tt)
	require.Equal(t, 2026, y)
	require.Equal(t, time.August, m)
	require.Equal(t, 24, d)
}

func TestSunsetRange(t *testing.T) {
	noon := time.Date(2026, 6, 21, 12, 0, 0, 0, Beijing)
	h := HoursUntilSunset(noon, 118.8, 30.1)
	require.Greater(t, h, 2.0)
	require.Less(t, h, 12.0)
}
