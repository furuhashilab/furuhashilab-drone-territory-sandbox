// Package gpx は GPX ファイルを読んで FlightRecord に変換する。
//
// GPX は経路をそのまま持っているので parse_level は常に "full" になる。
// ただし中身はただの XML で、時刻や座標を書き換えればポイントを水増しできる。
// 改ざんの検出はこのパッケージの範囲外。
package gpx

import (
	"encoding/xml"
	"fmt"
	"math"
	"time"

	"github.com/furuhashilab/furuhashilab-drone-territory-sandbox/worker/internal/flightrecord"
)

// earthRadiusM は haversine で使う地球半径。
const earthRadiusM = 6371000.0

type document struct {
	Tracks []track `xml:"trk"`
}

type track struct {
	Segments []segment `xml:"trkseg"`
}

type segment struct {
	Points []point `xml:"trkpt"`
}

type point struct {
	Lat  float64  `xml:"lat,attr"`
	Lon  float64  `xml:"lon,attr"`
	Ele  *float64 `xml:"ele"`
	Time *string  `xml:"time"`
}

// Parse は GPX のバイト列を FlightRecord に変換する。
//
// 飛行時間は最後の点の時刻から最初の点の時刻を引いたもの、距離は各点の間を haversine で
// 計算した合計。最高高度は最初の点の ele を 0 とした相対高度の最大値で、DJI の
// 「離陸地点からの高さ」と意味をそろえてある。
//
// <time> を持たない点が1つでもあればエラーを返す。飛行時間が出せないため。
// <ele> が無い場合はエラーにせず、max_height_m を nil にして警告を足す。
func Parse(data []byte) (*flightrecord.Record, error) {
	var doc document
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("GPX の解析に失敗しました: %w", err)
	}

	var points []point
	for _, trk := range doc.Tracks {
		for _, seg := range trk.Segments {
			points = append(points, seg.Points...)
		}
	}
	if len(points) < 2 {
		return nil, fmt.Errorf("GPX のトラックポイントが %d 個しかありません（2 個以上必要）", len(points))
	}

	times := make([]time.Time, len(points))
	hasEle := true
	for i, p := range points {
		if p.Time == nil {
			return nil, fmt.Errorf("%d 番目のトラックポイントに <time> がありません", i+1)
		}
		t, err := time.Parse(time.RFC3339, *p.Time)
		if err != nil {
			return nil, fmt.Errorf("%d 番目のトラックポイントの <time> を解釈できません (%q): %w", i+1, *p.Time, err)
		}
		times[i] = t.UTC()
		if p.Ele == nil {
			hasEle = false
		}
	}

	startedAt, endedAt := times[0], times[len(points)-1]

	var distance float64
	for i := 1; i < len(points); i++ {
		distance += haversine(points[i-1].Lat, points[i-1].Lon, points[i].Lat, points[i].Lon)
	}

	warnings := []string{}

	// 離陸地点（最初の点）の ele を 0 とした相対高度。
	var maxHeight *float64
	baseEle := 0.0
	if hasEle {
		baseEle = *points[0].Ele
		m := 0.0
		for _, p := range points {
			if rel := *p.Ele - baseEle; rel > m {
				m = rel
			}
		}
		maxHeight = &m
	} else {
		warnings = append(warnings, "<ele> の無いトラックポイントがあるため max_height_m を null にしました")
	}

	trackPoints := make([]flightrecord.TrackPoint, len(points))
	for i, p := range points {
		altRel := 0.0
		if p.Ele != nil {
			altRel = *p.Ele - baseEle
		}
		trackPoints[i] = flightrecord.TrackPoint{
			T:      times[i].Sub(startedAt).Seconds(),
			Lat:    p.Lat,
			Lon:    p.Lon,
			AltRel: altRel,
		}
	}

	rec := &flightrecord.Record{
		Source:       flightrecord.SourceGPX,
		ParseLevel:   flightrecord.ParseLevelFull,
		StartedAt:    startedAt,
		EndedAt:      endedAt,
		DurationSec:  endedAt.Sub(startedAt).Seconds(),
		DistanceM:    distance,
		MaxHeightM:   maxHeight,
		AircraftName: nil, // GPX は機体名を持たない
		AircraftID:   nil, // GPX は機体シリアルを持たない
		FileSHA256:   flightrecord.FileSHA256(data),
		Track:        trackPoints,
		Warnings:     warnings,
	}
	rec.Warnings = append(rec.Warnings, rec.Validate()...)

	return rec, nil
}

// haversine は2点間の大円距離をメートルで返す。
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	s1 := math.Sin(Δφ / 2)
	s2 := math.Sin(Δλ / 2)
	a := s1*s1 + math.Cos(φ1)*math.Cos(φ2)*s2*s2
	return 2 * earthRadiusM * math.Asin(math.Sqrt(a))
}
