package track

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"gocamping/internal/geo"
	"gocamping/internal/httpx"
	"gocamping/internal/model"
	"gocamping/internal/timeutil"
)

const (
	MaxBatchPoints   = 5000
	MaxAccuracyM     = 50
	MaxHikeKmh       = 25
	MaxAccelMS2      = 3.0
	StillRadiusM     = 15
	StillWindowSec   = 60
	QuantizePlaces   = 6
	GapThresholdSec  = 180
)

type ParsedPoint struct {
	Lat        float64
	Lon        float64
	Elevation  *float64
	Accuracy   *float64
	Speed      *float64
	RecordedAt time.Time
	Fingerprint string
}

func ValidateBatch(raw []model.RawPoint) ([]ParsedPoint, error) {
	if len(raw) == 0 {
		return nil, httpx.Validation("轨迹点不能为空")
	}
	if len(raw) > MaxBatchPoints {
		return nil, httpx.TooLarge(fmt.Sprintf("单批最多 %d 点", MaxBatchPoints))
	}
	out := make([]ParsedPoint, 0, len(raw))
	for i, p := range raw {
		if !geo.ValidateLatLon(p.Lat, p.Lon) {
			return nil, httpx.Validation(fmt.Sprintf("第 %d 点经纬度越界", i+1))
		}
		if p.RecordedAt == "" {
			return nil, httpx.Validation(fmt.Sprintf("第 %d 点缺少 recorded_at", i+1))
		}
		ts, err := timeutil.ParseISO(p.RecordedAt)
		if err != nil {
			return nil, httpx.Validation(fmt.Sprintf("第 %d 点时间无法解析", i+1))
		}
		now := timeutil.Now()
		if ts.After(now.Add(10 * time.Minute)) {
			return nil, httpx.Validation(fmt.Sprintf("第 %d 点时间在未来", i+1))
		}
		if ts.Before(now.Add(-90 * 24 * time.Hour)) {
			return nil, httpx.Validation(fmt.Sprintf("第 %d 点时间过旧", i+1))
		}
		if p.Accuracy != nil && *p.Accuracy < 0 {
			return nil, httpx.Validation(fmt.Sprintf("第 %d 点 accuracy 非法", i+1))
		}
		out = append(out, ParsedPoint{
			Lat:        p.Lat,
			Lon:        p.Lon,
			Elevation:  p.Elevation,
			Accuracy:   p.Accuracy,
			Speed:      p.Speed,
			RecordedAt: ts,
		})
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].RecordedAt.Before(out[j].RecordedAt)
	})
	return out, nil
}

func Fingerprint(memberID int64, lat, lon float64, t time.Time) string {
	qlat := geo.Quantize(lat, QuantizePlaces)
	qlon := geo.Quantize(lon, QuantizePlaces)
	s := fmt.Sprintf("%d|%.6f|%.6f|%d", memberID, qlat, qlon, t.Unix())
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func BatchHash(memberID, tripID int64, raw []model.RawPoint) string {
	h := sha256.New()
	fmt.Fprintf(h, "%d|%d|%d", tripID, memberID, len(raw))
	for _, p := range raw {
		fmt.Fprintf(h, "|%.6f,%.6f,%s", p.Lat, p.Lon, p.RecordedAt)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func AttachFingerprints(memberID int64, pts []ParsedPoint) []ParsedPoint {
	for i := range pts {
		pts[i].Fingerprint = Fingerprint(memberID, pts[i].Lat, pts[i].Lon, pts[i].RecordedAt)
	}
	return pts
}
