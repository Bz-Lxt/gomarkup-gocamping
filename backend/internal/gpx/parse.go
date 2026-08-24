package gpx

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"gocamping/internal/geo"
	"gocamping/internal/httpx"
	"gocamping/internal/model"
)

type doc struct {
	XMLName xml.Name `xml:"gpx"`
	Trk     []trk    `xml:"trk"`
	Wpt     []wpt    `xml:"wpt"`
}

type trk struct {
	Name string   `xml:"name"`
	Seg  []trkseg `xml:"trkseg"`
}

type trkseg struct {
	Pts []wpt `xml:"trkpt"`
}

type wpt struct {
	Lat string `xml:"lat,attr"`
	Lon string `xml:"lon,attr"`
	Ele string `xml:"ele"`
	Name string `xml:"name"`
}

func Parse(raw []byte) ([]model.Waypoint, error) {
	if len(raw) == 0 {
		return nil, httpx.Validation("GPX 为空")
	}
	if !strings.Contains(strings.ToLower(string(raw[:min(200, len(raw))])), "gpx") {
		return nil, httpx.Validation("不是 GPX 文档")
	}
	var d doc
	if err := xml.Unmarshal(raw, &d); err != nil {
		return nil, httpx.Validation("GPX 无法解析")
	}
	var out []model.Waypoint
	seq := 0
	push := func(w wpt, typ string) error {
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(w.Lat), 64)
		lon, err2 := strconv.ParseFloat(strings.TrimSpace(w.Lon), 64)
		if err1 != nil || err2 != nil || !geo.ValidateLatLon(lat, lon) {
			return httpx.Validation(fmt.Sprintf("第 %d 个 GPX 点经纬度非法", seq+1))
		}
		var ele *float64
		if strings.TrimSpace(w.Ele) != "" {
			if v, err := strconv.ParseFloat(w.Ele, 64); err == nil {
				ele = &v
			}
		}
		out = append(out, model.Waypoint{
			Seq: seq, Type: typ, Lat: lat, Lon: lon, Elevation: ele,
			Note: w.Name, RiskWeight: 1, Polygon: [][2]float64{},
		})
		seq++
		return nil
	}
	for _, w := range d.Wpt {
		if err := push(w, "waypoint"); err != nil {
			return nil, err
		}
	}
	for _, t := range d.Trk {
		for _, s := range t.Seg {
			for _, p := range s.Pts {
				if err := push(p, "waypoint"); err != nil {
					return nil, err
				}
			}
		}
	}
	if len(out) == 0 {
		return nil, httpx.Validation("GPX 中没有有效点")
	}
	return out, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
