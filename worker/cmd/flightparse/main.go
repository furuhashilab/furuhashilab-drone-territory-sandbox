// flightparse は GPX / DJI ログを FlightRecord の JSON に変換して標準出力に出す。
//
//	flightparse <ファイル>
//
// 環境変数:
//
//	DJIRECORD_BIN     djirecord の実行ファイルのパス（DJI ログを読むのに必要）
//	FLIGHT_ID_SECRET  機体シリアルを HMAC-SHA256 にかける鍵。未設定なら aircraft_id は null
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"unicode"

	"github.com/furuhashilab/furuhashilab-drone-territory-sandbox/worker/internal/djisummary"
	"github.com/furuhashilab/furuhashilab-drone-territory-sandbox/worker/internal/flightrecord"
	"github.com/furuhashilab/furuhashilab-drone-territory-sandbox/worker/internal/gpx"
)

// envSecret は機体シリアルを HMAC にかける鍵を渡す環境変数名。
const envSecret = "FLIGHT_ID_SECRET"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "flightparse:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("使い方: flightparse <ファイル>")
	}
	path := args[0]

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ファイルを読めません: %w", err)
	}

	// 形式はファイル名ではなく中身で判定する。拡張子は当てにならないため
	// （DJI のログは .txt だが中身はバイナリ）。
	var rec *flightrecord.Record
	if looksLikeXML(data) {
		rec, err = gpx.Parse(data)
	} else {
		rec, err = djisummary.Parse(path, data, os.Getenv(envSecret))
	}
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON に変換できません: %w", err)
	}
	fmt.Println(string(out))

	return nil
}

// looksLikeXML は中身が XML（＝ GPX）かどうかを、先頭の空白と BOM を飛ばして判定する。
func looksLikeXML(data []byte) bool {
	data = trimBOM(data)
	for i, b := range data {
		if unicode.IsSpace(rune(b)) {
			continue
		}
		return data[i] == '<'
	}
	return false
}

func trimBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}
