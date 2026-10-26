package gpx

import (
	"math"
	"os"
	"testing"
)

// 期待値は samples/gpx/synthetic_normal.gpx のコメントに書いてあるもの。
// distance_m は Go の実装とは別に Python で haversine を計算して出した値。
const (
	wantDurationSec = 120.0
	wantDistanceM   = 121.368287
	wantMaxHeightM  = 30.0
	wantPoints      = 13
)

func readSample(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../../samples/gpx/" + name)
	if err != nil {
		t.Fatalf("サンプルを読めません: %v", err)
	}
	return data
}

func TestParseSyntheticNormal(t *testing.T) {
	rec, err := Parse(readSample(t, "synthetic_normal.gpx"))
	if err != nil {
		t.Fatalf("Parse が失敗しました: %v", err)
	}

	if rec.Source != "gpx" {
		t.Errorf("source = %q, want %q", rec.Source, "gpx")
	}
	if rec.ParseLevel != "full" {
		t.Errorf("parse_level = %q, want %q", rec.ParseLevel, "full")
	}
	if rec.DurationSec != wantDurationSec {
		t.Errorf("duration_sec = %v, want %v", rec.DurationSec, wantDurationSec)
	}
	if math.Abs(rec.DistanceM-wantDistanceM) > 0.01 {
		t.Errorf("distance_m = %v, want %v ± 0.01", rec.DistanceM, wantDistanceM)
	}
	if rec.MaxHeightM == nil {
		t.Fatal("max_height_m が nil です")
	}
	if *rec.MaxHeightM != wantMaxHeightM {
		t.Errorf("max_height_m = %v, want %v", *rec.MaxHeightM, wantMaxHeightM)
	}
	if len(rec.Track) != wantPoints {
		t.Errorf("track の点数 = %d, want %d", len(rec.Track), wantPoints)
	}
	if len(rec.Warnings) != 0 {
		t.Errorf("warnings = %v, want 空", rec.Warnings)
	}

	// GPX は機体の情報を持たない。
	if rec.AircraftName != nil {
		t.Errorf("aircraft_name = %v, want nil", *rec.AircraftName)
	}
	if rec.AircraftID != nil {
		t.Errorf("aircraft_id = %v, want nil", *rec.AircraftID)
	}

	// started_at / ended_at の差が duration_sec と一致すること。
	if got := rec.EndedAt.Sub(rec.StartedAt).Seconds(); got != rec.DurationSec {
		t.Errorf("ended_at - started_at = %v, duration_sec = %v", got, rec.DurationSec)
	}

	// 最初の点は t=0、alt_rel=0 で始まる。
	if rec.Track[0].T != 0 || rec.Track[0].AltRel != 0 {
		t.Errorf("track[0] = %+v, want t=0 / alt_rel=0", rec.Track[0])
	}
	if last := rec.Track[len(rec.Track)-1]; last.T != wantDurationSec {
		t.Errorf("track の最後の t = %v, want %v", last.T, wantDurationSec)
	}
}

func TestParseNoTimeIsError(t *testing.T) {
	if _, err := Parse(readSample(t, "synthetic_no_time.gpx")); err == nil {
		t.Fatal("<time> の無い GPX でエラーになりませんでした")
	}
}

func TestParseNoEleWarnsAndNullsMaxHeight(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
  <trkpt lat="35.0" lon="135.0"><time>2026-06-20T01:00:00Z</time></trkpt>
  <trkpt lat="35.0" lon="135.0001"><time>2026-06-20T01:00:10Z</time></trkpt>
</trkseg></trk></gpx>`)

	rec, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse が失敗しました: %v", err)
	}
	if rec.MaxHeightM != nil {
		t.Errorf("max_height_m = %v, want nil", *rec.MaxHeightM)
	}
	if len(rec.Warnings) == 0 {
		t.Error("<ele> が無いのに警告が出ていません")
	}
}

func TestParseTooFewPointsIsError(t *testing.T) {
	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<gpx version="1.1" xmlns="http://www.topografix.com/GPX/1/1"><trk><trkseg>
  <trkpt lat="35.0" lon="135.0"><ele>0</ele><time>2026-06-20T01:00:00Z</time></trkpt>
</trkseg></trk></gpx>`)

	if _, err := Parse(data); err == nil {
		t.Fatal("点が1つしかない GPX でエラーになりませんでした")
	}
}
