package track

import "gocamping/internal/model"

// Pipeline is the public Delta Track Merge entry used by services and tests.
type Pipeline struct {
	MemberID int64
	TripID   int64
}

type PipelineOut struct {
	Clean     []ParsedPoint
	Smoothed  []ParsedPoint
	Merged    []ParsedPoint
	Display   []ParsedPoint
	Stats     FilterStats
	Metrics   Metrics
	Segments  []model.TrackSegment
	Dups      int
	BatchHash string
}

func (p Pipeline) Run(raw []model.RawPoint, existing []ParsedPoint, known map[string]struct{}) (*PipelineOut, error) {
	parsed, err := ValidateBatch(raw)
	if err != nil {
		return nil, err
	}
	parsed = AttachFingerprints(p.MemberID, parsed)
	fresh, dups := DedupeByFingerprint(parsed, known)
	clean, st := Denoise(fresh)
	smooth := Smooth(clean)
	merged, segs := MergeIncremental(existing, smooth)
	out := &PipelineOut{
		Clean:     clean,
		Smoothed:  smooth,
		Merged:    merged,
		Display:   DouglasPeucker(merged, 8),
		Stats:     st,
		Metrics:   ComputeMetrics(merged),
		Segments:  segs,
		Dups:      dups,
		BatchHash: BatchHash(p.MemberID, p.TripID, raw),
	}
	return out, nil
}

func EmptyKnown() map[string]struct{} { return map[string]struct{}{} }
