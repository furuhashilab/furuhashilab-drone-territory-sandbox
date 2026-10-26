package djisummary

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/furuhashilab/furuhashilab-drone-territory-sandbox/worker/internal/flightrecord"
)

const dummyFileSHA = "0000000000000000000000000000000000000000000000000000000000000000"

func readDummy(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/djirecord_dummy.json")
	if err != nil {
		t.Fatalf("テスト用 JSON を読めません: %v", err)
	}
	return data
}

func TestConvertDummy(t *testing.T) {
	rec, err := Convert(readDummy(t), dummyFileSHA, "")
	if err != nil {
		t.Fatalf("Convert が失敗しました: %v", err)
	}

	if rec.Source != "dji" {
		t.Errorf("source = %q, want %q", rec.Source, "dji")
	}
	if rec.ParseLevel != "summary" {
		t.Errorf("parse_level = %q, want %q", rec.ParseLevel, "summary")
	}
	if rec.DurationSec != 300.0 {
		t.Errorf("duration_sec = %v, want 300", rec.DurationSec)
	}
	if rec.DistanceM != 1234.5 {
		t.Errorf("distance_m = %v, want 1234.5", rec.DistanceM)
	}
	if rec.MaxHeightM == nil || *rec.MaxHeightM != 42.5 {
		t.Errorf("max_height_m = %v, want 42.5", rec.MaxHeightM)
	}
	if rec.AircraftName == nil || *rec.AircraftName != "DUMMY Aircraft" {
		t.Errorf("aircraft_name = %v, want %q", rec.AircraftName, "DUMMY Aircraft")
	}
	if rec.FileSHA256 != dummyFileSHA {
		t.Errorf("file_sha256 = %q, want %q", rec.FileSHA256, dummyFileSHA)
	}

	// 概要のみなので経路は空配列。nil だと JSON で null になってしまう。
	if rec.Track == nil || len(rec.Track) != 0 {
		t.Errorf("track = %v, want 空配列", rec.Track)
	}

	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if !rec.StartedAt.Equal(want) {
		t.Errorf("started_at = %v, want %v", rec.StartedAt, want)
	}
	// ended_at = started_at + duration。
	if got := rec.EndedAt.Sub(rec.StartedAt).Seconds(); got != 300.0 {
		t.Errorf("ended_at - started_at = %v, want 300", got)
	}

	// 平均速度 1234.5/300 = 4.1 m/s、高度 42.5 m。どちらも閾値内なので警告は出ない。
	if len(rec.Warnings) != 0 {
		t.Errorf("warnings = %v, want 空", rec.Warnings)
	}
}

func TestConvertAircraftIDIsHMACNotSerial(t *testing.T) {
	const serial = "DUMMY-AIRCRAFT-SN-0000"
	raw := readDummy(t)

	// FLIGHT_ID_SECRET が未設定なら aircraft_id は null。
	withoutSecret, err := Convert(raw, dummyFileSHA, "")
	if err != nil {
		t.Fatalf("Convert が失敗しました: %v", err)
	}
	if withoutSecret.AircraftID != nil {
		t.Errorf("secret なしの aircraft_id = %v, want nil", *withoutSecret.AircraftID)
	}

	withSecret, err := Convert(raw, dummyFileSHA, "test-secret")
	if err != nil {
		t.Fatalf("Convert が失敗しました: %v", err)
	}
	if withSecret.AircraftID == nil {
		t.Fatal("secret ありなのに aircraft_id が nil です")
	}
	if want := flightrecord.AircraftID(serial, "test-secret"); *withSecret.AircraftID != *want {
		t.Errorf("aircraft_id が HMAC-SHA256 と一致しません")
	}

	// 鍵が違えば別の値になる。
	other, _ := Convert(raw, dummyFileSHA, "another-secret")
	if *other.AircraftID == *withSecret.AircraftID {
		t.Error("鍵を変えても aircraft_id が同じです")
	}

	// 生のシリアルが JSON のどこにも出ないこと。
	out, err := json.Marshal(withSecret)
	if err != nil {
		t.Fatalf("Marshal が失敗しました: %v", err)
	}
	if strings.Contains(string(out), serial) {
		t.Error("出力に生のシリアル番号が含まれています")
	}
}

func TestConvertInvalidJSON(t *testing.T) {
	if _, err := Convert([]byte("not json"), dummyFileSHA, ""); err == nil {
		t.Fatal("壊れた JSON でエラーになりませんでした")
	}
	if _, err := Convert([]byte(`{"version":14,"details":{}}`), dummyFileSHA, ""); err == nil {
		t.Fatal("start_time が無いのにエラーになりませんでした")
	}
}

// summaryJSON は samples/dji/neo_2026-06-20_summary.json の内容。
type summaryJSON struct {
	SourceSHA256 string  `json:"source_sha256"`
	AircraftName string  `json:"aircraft_name"`
	StartTimeUTC string  `json:"start_time_utc"`
	DurationSec  float64 `json:"duration_sec"`
	DistanceM    float64 `json:"distance_m"`
	MaxHeightM   float64 `json:"max_height_m"`
}

// TestParseRealLog は samples/private/ に生ログがあるときだけ動く。
// 生ログはリポジトリに入っていないので、CI や他の人の手元では skip される。
//
// 失敗時に出すのは、要約 JSON に載っている値（飛行時間・距離・高度・ハッシュ）だけにしてある。
// シリアル番号や座標は出さない。
func TestParseRealLog(t *testing.T) {
	logs, err := filepath.Glob("../../../samples/private/*.txt")
	if err != nil || len(logs) == 0 {
		t.Skip("samples/private/ に生ログが無いので skip します")
	}
	if os.Getenv(EnvBin) == "" {
		t.Skipf("%s が未設定なので skip します", EnvBin)
	}

	raw, err := os.ReadFile("../../../samples/dji/neo_2026-06-20_summary.json")
	if err != nil {
		t.Fatalf("要約 JSON を読めません: %v", err)
	}
	var want summaryJSON
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("要約 JSON を解釈できません: %v", err)
	}

	// 要約 JSON と同じログを探す（source_sha256 が一致するもの）。
	var path string
	var data []byte
	for _, p := range logs {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if flightrecord.FileSHA256(b) == want.SourceSHA256 {
			path, data = p, b
			break
		}
	}
	if path == "" {
		t.Skipf("source_sha256 が %s と一致する生ログが samples/private/ にありません", want.SourceSHA256[:8])
	}

	rec, err := Parse(path, data, "")
	if err != nil {
		t.Fatalf("Parse が失敗しました: %v", err)
	}

	if rec.FileSHA256 != want.SourceSHA256 {
		t.Errorf("file_sha256 = %q, want %q", rec.FileSHA256, want.SourceSHA256)
	}
	if rec.DurationSec != want.DurationSec {
		t.Errorf("duration_sec = %v, want %v", rec.DurationSec, want.DurationSec)
	}
	if rec.DistanceM != want.DistanceM {
		t.Errorf("distance_m = %v, want %v", rec.DistanceM, want.DistanceM)
	}
	if rec.MaxHeightM == nil || *rec.MaxHeightM != want.MaxHeightM {
		t.Errorf("max_height_m = %v, want %v", rec.MaxHeightM, want.MaxHeightM)
	}
	if rec.AircraftName == nil || *rec.AircraftName != want.AircraftName {
		t.Errorf("aircraft_name = %v, want %q", rec.AircraftName, want.AircraftName)
	}

	wantStart, err := time.Parse(time.RFC3339, want.StartTimeUTC)
	if err != nil {
		t.Fatalf("要約 JSON の start_time_utc を解釈できません: %v", err)
	}
	if !rec.StartedAt.Equal(wantStart) {
		t.Errorf("started_at = %v, want %v", rec.StartedAt, wantStart)
	}

	if len(rec.Track) != 0 {
		t.Errorf("track の点数 = %d, want 0（概要のみ）", len(rec.Track))
	}
}
