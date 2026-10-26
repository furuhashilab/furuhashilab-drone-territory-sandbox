package flightrecord

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	height := func(v float64) *float64 { return &v }

	tests := []struct {
		name         string
		rec          Record
		wantWarnings int
	}{
		{
			name:         "正常な値では警告なし",
			rec:          Record{DurationSec: 241, DistanceM: 103.2, MaxHeightM: height(5)},
			wantWarnings: 0,
		},
		{
			name:         "duration_sec が 0",
			rec:          Record{DurationSec: 0, DistanceM: 0},
			wantWarnings: 1,
		},
		{
			name:         "duration_sec が負",
			rec:          Record{DurationSec: -1, DistanceM: 0},
			wantWarnings: 1,
		},
		{
			name:         "duration_sec が 3600 超",
			rec:          Record{DurationSec: 3601, DistanceM: 0},
			wantWarnings: 1,
		},
		{
			name:         "3600 ちょうどは警告なし",
			rec:          Record{DurationSec: 3600, DistanceM: 0},
			wantWarnings: 0,
		},
		{
			name:         "平均速度が 25 m/s 超",
			rec:          Record{DurationSec: 10, DistanceM: 300},
			wantWarnings: 1,
		},
		{
			name:         "max_height_m が 150 m 超",
			rec:          Record{DurationSec: 60, DistanceM: 10, MaxHeightM: height(151)},
			wantWarnings: 1,
		},
		{
			name:         "max_height_m が nil なら高度の警告は出ない",
			rec:          Record{DurationSec: 60, DistanceM: 10, MaxHeightM: nil},
			wantWarnings: 0,
		},
		{
			name:         "複数の警告が同時に出る",
			rec:          Record{DurationSec: 3601, DistanceM: 1000000, MaxHeightM: height(200)},
			wantWarnings: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rec.Validate(); len(got) != tt.wantWarnings {
				t.Errorf("警告 %d 件, want %d 件: %v", len(got), tt.wantWarnings, got)
			}
		})
	}
}

func TestAircraftID(t *testing.T) {
	const serial = "DUMMY-SN-0001"

	if got := AircraftID(serial, ""); got != nil {
		t.Errorf("secret が空のとき = %v, want nil", *got)
	}
	if got := AircraftID("", "secret"); got != nil {
		t.Errorf("serial が空のとき = %v, want nil", *got)
	}

	got := AircraftID(serial, "secret")
	if got == nil {
		t.Fatal("nil が返りました")
	}
	if len(*got) != 64 {
		t.Errorf("長さ = %d, want 64（SHA-256 の 16 進表記）", len(*got))
	}
	if strings.Contains(*got, serial) {
		t.Error("出力に生のシリアルが含まれています")
	}

	// 同じ入力なら同じ値、違う鍵なら違う値。
	if again := AircraftID(serial, "secret"); *again != *got {
		t.Error("同じ入力で値が変わりました")
	}
	if other := AircraftID(serial, "other"); *other == *got {
		t.Error("鍵を変えても同じ値になりました")
	}
}

// Track と Warnings は空のとき JSON で null ではなく [] になってほしい。
func TestEmptySlicesMarshalAsArrays(t *testing.T) {
	rec := Record{Track: []TrackPoint{}, Warnings: []string{}}
	out, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("Marshal が失敗しました: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `"track":[]`) {
		t.Errorf("track が [] になっていません: %s", s)
	}
	if !strings.Contains(s, `"warnings":[]`) {
		t.Errorf("warnings が [] になっていません: %s", s)
	}
}
