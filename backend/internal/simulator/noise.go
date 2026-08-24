package simulator

import "gocamping/internal/model"

// InjectOutliers appends teleport jumps that the denoise engine must drop.
func InjectOutliers(pts []model.RawPoint, every int, jumpM float64) []model.RawPoint {
	if every <= 0 {
		every = 7
	}
	out := make([]model.RawPoint, 0, len(pts)+len(pts)/every)
	for i, p := range pts {
		out = append(out, p)
		if i > 0 && i%every == 0 {
			out = append(out, Spike(p, jumpM))
		}
	}
	return out
}
