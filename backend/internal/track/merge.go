package track

import (
	"time"

	"gocamping/internal/model"
)

// MergeIncremental folds a new (already denoised+smoothed) batch into the
// existing time-ordered trajectory. Overlap cases:
//   - complete overlap -> drop incoming
//   - partial overlap  -> clip incoming to the non-overlapping tails
//   - temporal hole    -> emit a gap segment
func MergeIncremental(existing, incoming []ParsedPoint) (merged []ParsedPoint, segs []model.TrackSegment) {
	if len(existing) == 0 {
		return incoming, buildSegs(incoming)
	}
	if len(incoming) == 0 {
		return existing, buildSegs(existing)
	}
	seen := map[string]struct{}{}
	for _, p := range existing {
		if p.Fingerprint != "" {
			seen[p.Fingerprint] = struct{}{}
		}
	}
	clipped := make([]ParsedPoint, 0, len(incoming))
	for _, p := range incoming {
		if p.Fingerprint != "" {
			if _, ok := seen[p.Fingerprint]; ok {
				continue
			}
		}
		clipped = append(clipped, p)
	}
	merged = make([]ParsedPoint, 0, len(existing)+len(clipped))
	i, j := 0, 0
	for i < len(existing) && j < len(clipped) {
		if !clipped[j].RecordedAt.After(existing[i].RecordedAt) {
			if !sameTick(existing[i], clipped[j]) {
				merged = append(merged, clipped[j])
			}
			j++
			continue
		}
		merged = append(merged, existing[i])
		i++
	}
	merged = append(merged, existing[i:]...)
	merged = append(merged, clipped[j:]...)
	return merged, buildSegs(merged)
}

func sameTick(a, b ParsedPoint) bool {
	return a.RecordedAt.Unix() == b.RecordedAt.Unix() &&
		abs(a.Lat-b.Lat) < 1e-6 && abs(a.Lon-b.Lon) < 1e-6
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func buildSegs(pts []ParsedPoint) []model.TrackSegment {
	if len(pts) == 0 {
		return []model.TrackSegment{}
	}
	var segs []model.TrackSegment
	start := 0
	for i := 1; i < len(pts); i++ {
		dt := pts[i].RecordedAt.Sub(pts[i-1].RecordedAt)
		if dt > time.Duration(GapThresholdSec)*time.Second {
			segs = append(segs, sliceSeg(pts[start:i], false))
			segs = append(segs, model.TrackSegment{
				StartAt:   pts[i-1].RecordedAt,
				EndAt:     pts[i].RecordedAt,
				DistanceM: 0,
				IsGap:     true,
			})
			start = i
		}
	}
	segs = append(segs, sliceSeg(pts[start:], false))
	return segs
}

func sliceSeg(pts []ParsedPoint, gap bool) model.TrackSegment {
	m := ComputeMetrics(pts)
	st, en := pts[0].RecordedAt, pts[len(pts)-1].RecordedAt
	return model.TrackSegment{StartAt: st, EndAt: en, DistanceM: m.DistanceM, IsGap: gap}
}

func DedupeByFingerprint(pts []ParsedPoint, known map[string]struct{}) (fresh []ParsedPoint, dups int) {
	fresh = make([]ParsedPoint, 0, len(pts))
	for _, p := range pts {
		if p.Fingerprint != "" {
			if _, ok := known[p.Fingerprint]; ok {
				dups++
				continue
			}
		}
		fresh = append(fresh, p)
	}
	return fresh, dups
}
