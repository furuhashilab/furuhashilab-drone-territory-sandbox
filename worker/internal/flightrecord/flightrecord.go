// Package flightrecord は、GPX / DJI ログを変換した先の共通形式 FlightRecord を定義する。
//
// 項目は Issue #6 で検討中のため、確定ではない。
package flightrecord

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// source の値。
const (
	SourceGPX = "gpx"
	SourceDJI = "dji"
)

// parse_level の値。
//
// ParseLevelSummary は DJI ログの概要だけを読んだ状態で、経路を持たない。
// 経路（本体の時系列データ）は暗号化されており、復号に DJI の API キーが要る。
// キー発行後に ParseLevelFull を使う。
const (
	ParseLevelSummary = "summary"
	ParseLevelFull    = "full"
)

// 値チェックの閾値。超えてもエラーにはせず Warnings に入れる。
const (
	MaxDurationSec = 3600.0
	MaxAvgSpeedMPS = 25.0
	MaxHeightM     = 150.0
)

// TrackPoint は経路上の1点。t は飛行開始からの経過秒、alt_rel は離陸地点からの相対高度。
type TrackPoint struct {
	T      float64 `json:"t"`
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	AltRel float64 `json:"alt_rel"`
}

// Record は1回の飛行を表す。
//
// user_id は持たない（サーバー側で付ける）。
// 機体シリアルは AircraftID に HMAC をかけた形でしか入らず、生の値はどこにも出力しない。
// DJI の離陸地点の座標は MVP では使わないので含めない。
type Record struct {
	Source       string       `json:"source"`
	ParseLevel   string       `json:"parse_level"`
	StartedAt    time.Time    `json:"started_at"`
	EndedAt      time.Time    `json:"ended_at"`
	DurationSec  float64      `json:"duration_sec"`
	DistanceM    float64      `json:"distance_m"`
	MaxHeightM   *float64     `json:"max_height_m"`
	AircraftName *string      `json:"aircraft_name"`
	AircraftID   *string      `json:"aircraft_id"`
	FileSHA256   string       `json:"file_sha256"`
	Track        []TrackPoint `json:"track"`
	Warnings     []string     `json:"warnings"`
}

// Validate は値の妥当性を調べ、警告の一覧を返す。おかしな値でもエラーにはしない。
//
// 判定できるのは「値が常識的な範囲に収まっているか」だけで、改ざんの検出はできない。
// GPX は単なる XML なので書き換えられるし、DJI も概要だけでは本体データと照合できない。
func (r *Record) Validate() []string {
	warnings := []string{}

	switch {
	case r.DurationSec <= 0:
		warnings = append(warnings,
			fmt.Sprintf("duration_sec が 0 以下です（%.1f 秒）", r.DurationSec))
	case r.DurationSec > MaxDurationSec:
		warnings = append(warnings,
			fmt.Sprintf("duration_sec が %.0f 秒を超えています（%.1f 秒）", MaxDurationSec, r.DurationSec))
	}

	if r.DurationSec > 0 {
		if avg := r.DistanceM / r.DurationSec; avg > MaxAvgSpeedMPS {
			warnings = append(warnings,
				fmt.Sprintf("平均速度が %.0f m/s を超えています（%.2f m/s）", MaxAvgSpeedMPS, avg))
		}
	}

	if r.MaxHeightM != nil && *r.MaxHeightM > MaxHeightM {
		warnings = append(warnings,
			fmt.Sprintf("max_height_m が %.0f m を超えています（%.1f m）", MaxHeightM, *r.MaxHeightM))
	}

	return warnings
}

// AircraftID は機体シリアルを secret で HMAC-SHA256 した文字列を返す。
//
// 同じ機体を同じ ID として扱えるようにするためのもので、生のシリアルは復元できない。
// secret（環境変数 FLIGHT_ID_SECRET）が空のときは nil を返し、機体の区別をあきらめる。
func AircraftID(serial, secret string) *string {
	if serial == "" || secret == "" {
		return nil
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(serial))
	id := hex.EncodeToString(mac.Sum(nil))
	return &id
}

// FileSHA256 は入力ファイル全体のハッシュを返す。同じログの二重登録をはじくのに使う。
func FileSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
