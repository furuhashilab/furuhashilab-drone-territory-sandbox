// Package djisummary は DJI の飛行ログ（FlightRecord_*.txt）の概要部分を FlightRecord に変換する。
//
// 読み取りには外部コマンド djirecord（pydjirecord）を使う。パスは環境変数 DJIRECORD_BIN で指定する。
//
// ログ本体（経路の時系列データ）は暗号化されていて、復号には DJI の API キーが要る。
// キー無しで読めるのは概要だけなので parse_level は "summary" になり、track は空配列になる。
// MVP では陣地を経路ではなくポイントのクリックで取るため、これで足りる。
package djisummary

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/furuhashilab/furuhashilab-drone-territory-sandbox/worker/internal/flightrecord"
)

// EnvBin は djirecord の実行ファイルのパスを渡す環境変数名。
const EnvBin = "DJIRECORD_BIN"

// defaultBin は DJIRECORD_BIN が未設定のときに使う名前（PATH から探す）。
const defaultBin = "djirecord"

// output は djirecord --json の出力のうち、変換に使う項目だけを写したもの。
//
// 実際の出力にはこのほかに離陸地点の座標（latitude / longitude / take_off_altitude）や
// 各種シリアル（camera_sn / rc_sn / battery_sn）、住所（city / street など）も入るが、
// FlightRecord では使わないので読まない。
type output struct {
	Version int `json:"version"`
	Details struct {
		StartTime     string  `json:"start_time"`
		TotalTime     float64 `json:"total_time"`
		TotalDistance float64 `json:"total_distance"`
		MaxHeight     float64 `json:"max_height"`
		AircraftName  string  `json:"aircraft_name"`
		AircraftSN    string  `json:"aircraft_sn"`
		// VideoTime は読むが使わない。実ログで飛行時間より長い値が出ており、
		// 何を指しているのか分かっていないため（samples/README.md 参照）。
		VideoTime int `json:"video_time"`
	} `json:"details"`
}

// startTimeLayouts は djirecord が返す start_time の書式の候補。
var startTimeLayouts = []string{
	time.RFC3339Nano,
	"2006-01-02T15:04:05.999999",
	"2006-01-02 15:04:05.999999",
}

// Convert は djirecord --json の出力を FlightRecord に変換する。
//
// fileSHA256 は元の生ログ全体のハッシュ、secret は機体シリアルを HMAC にかける鍵
// （環境変数 FLIGHT_ID_SECRET）。secret が空なら aircraft_id は nil になる。
// 生のシリアルは戻り値のどこにも入らない。
func Convert(raw []byte, fileSHA256, secret string) (*flightrecord.Record, error) {
	var out output
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("djirecord の JSON を解釈できません: %w", err)
	}
	d := out.Details

	if d.StartTime == "" {
		return nil, fmt.Errorf("djirecord の出力に start_time がありません")
	}
	startedAt, err := parseStartTime(d.StartTime)
	if err != nil {
		return nil, err
	}

	maxHeight := d.MaxHeight
	var aircraftName *string
	if d.AircraftName != "" {
		name := d.AircraftName
		aircraftName = &name
	}

	rec := &flightrecord.Record{
		Source:       flightrecord.SourceDJI,
		ParseLevel:   flightrecord.ParseLevelSummary,
		StartedAt:    startedAt,
		EndedAt:      startedAt.Add(time.Duration(d.TotalTime * float64(time.Second))),
		DurationSec:  d.TotalTime,
		DistanceM:    d.TotalDistance,
		MaxHeightM:   &maxHeight,
		AircraftName: aircraftName,
		AircraftID:   flightrecord.AircraftID(d.AircraftSN, secret),
		FileSHA256:   fileSHA256,
		Track:        []flightrecord.TrackPoint{}, // 概要のみなので経路は無い
		Warnings:     []string{},
	}
	rec.Warnings = append(rec.Warnings, rec.Validate()...)

	return rec, nil
}

// Parse は生ログのパスを受け取り、djirecord を呼び出して FlightRecord に変換する。
func Parse(path string, data []byte, secret string) (*flightrecord.Record, error) {
	raw, err := run(path)
	if err != nil {
		return nil, err
	}
	return Convert(raw, flightrecord.FileSHA256(data), secret)
}

// run は djirecord --json <path> を実行して標準出力を返す。
func run(path string) ([]byte, error) {
	bin := os.Getenv(EnvBin)
	if bin == "" {
		bin = defaultBin
	}

	cmd := exec.Command(bin, "--json", path)
	stdout, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("%s の実行に失敗しました: %w: %s", bin, err, exitErr.Stderr)
		}
		return nil, fmt.Errorf("%s を実行できません（環境変数 %s でパスを指定してください）: %w", bin, EnvBin, err)
	}
	return stdout, nil
}

func parseStartTime(s string) (time.Time, error) {
	for _, layout := range startTimeLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("start_time を解釈できません (%q)", s)
}
