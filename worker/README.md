# worker — 飛行ログ → FlightRecord 変換ツール

GPX と DJI の飛行ログを、共通形式 `FlightRecord` の JSON に変換する CLI です。

DJI の API キーを使わない最小構成（MVP）です。DJI ログについては概要だけを読み、経路は扱いません。
詳しくは [MVP の割り切り](#mvp-の割り切り) を参照してください。

関連 Issue: #16（このツール）、#6（FlightRecord の項目）、#3（サンプルログ）

## ビルド

```bash
cd worker
go build -o flightparse ./cmd/flightparse
```

## 使い方

```bash
flightparse <ファイル>
```

`FlightRecord` の JSON を標準出力に出します。

```bash
# GPX
flightparse ../samples/gpx/synthetic_normal.gpx

# DJI ログ（DJIRECORD_BIN が必要）
export DJIRECORD_BIN=~/.venvs/djirecord/bin/djirecord
export FLIGHT_ID_SECRET=...
flightparse ../samples/private/FlightRecord_2026-06-20_xx-xx-xx.txt
```

**形式はファイル名ではなく中身で判定します。** 先頭の空白と BOM を飛ばして `<` で始まっていれば
GPX（XML）、そうでなければ DJI ログとして扱います。DJI のログは拡張子が `.txt` でも中身はバイナリ
なので、拡張子は当てになりません。

## 環境変数

| 変数 | 必須 | 内容 |
|---|---|---|
| `DJIRECORD_BIN` | DJI ログを読むときのみ | `djirecord` の実行ファイルのパス。未設定なら `PATH` から `djirecord` を探します |
| `FLIGHT_ID_SECRET` | 任意 | 機体シリアルを HMAC-SHA256 にかける鍵。未設定なら `aircraft_id` は `null` |

`djirecord`（pydjirecord）の導入手順は [`../samples/README.md`](../samples/README.md) にあります。

> [!CAUTION]
> `FLIGHT_ID_SECRET` は機体の同一性を判定する鍵です。コミットせず、`.env` などに置いてください。
> 鍵が変わると同じ機体でも別の `aircraft_id` になります。

## FlightRecord の項目

```json
{
  "source": "gpx",
  "parse_level": "full",
  "started_at": "2026-06-20T01:00:00Z",
  "ended_at": "2026-06-20T01:02:00Z",
  "duration_sec": 120,
  "distance_m": 121.36828687107575,
  "max_height_m": 30,
  "aircraft_name": null,
  "aircraft_id": null,
  "file_sha256": "eeb98ebb...",
  "track": [{ "t": 0, "lat": 35, "lon": 135, "alt_rel": 0 }],
  "warnings": []
}
```

| 項目 | 型 | 内容 |
|---|---|---|
| `source` | string | `"gpx"` / `"dji"` |
| `parse_level` | string | `"summary"`（DJI の概要のみ・経路なし）/ `"full"`（経路あり）。GPX は常に `"full"` |
| `started_at` | string | 飛行開始時刻。UTC・RFC3339 |
| `ended_at` | string | 飛行終了時刻。UTC・RFC3339 |
| `duration_sec` | number | 飛行時間（秒） |
| `distance_m` | number | 飛行距離（メートル） |
| `max_height_m` | number \| null | 離陸地点からの最高高度（メートル） |
| `aircraft_name` | string \| null | 機体名。GPX は `null`（GPX は機体情報を持たない） |
| `aircraft_id` | string \| null | 機体シリアルの HMAC-SHA256（16進64文字）。`FLIGHT_ID_SECRET` 未設定なら `null` |
| `file_sha256` | string | 入力ファイル全体の SHA-256。同じログの二重登録をはじくのに使う |
| `track` | array | 経路。`t` は飛行開始からの経過秒、`alt_rel` は離陸地点からの相対高度。DJI の `summary` では空配列 |
| `warnings` | array | 値チェックの警告。空なら `[]` |

### 持たないもの

- **生の機体シリアル番号** — `aircraft_id` に HMAC をかけた形でしか入りません
- **DJI の離陸地点の座標・高度** — MVP では使わないので含めません
- `user_id` — サーバー側で付けます

## 変換ルール

### GPX

| 項目 | 求め方 |
|---|---|
| 飛行時間 | 最後の点の時刻 − 最初の点の時刻 |
| 距離 | 各点の間を haversine 式（R = 6371000 m）で計算して合計 |
| 最高高度 | **最初の点の `<ele>` を 0 とした相対高度**の最大値 |

最高高度を相対にしているのは、DJI の「離陸地点からの高さ」と意味をそろえるためです。

- `<time>` が無い点が1つでもあると**エラー**になります（飛行時間が出せないため）
- `<ele>` が無い場合はエラーにせず、`max_height_m` を `null` にして警告を足します

### DJI

`djirecord --json` を呼び出し、概要（`details`）から変換します。
`track` は常に空配列、`parse_level` は常に `"summary"` です。

`video_time` は読みません。実ログで飛行時間より長い値が出ており、何を指しているのか
分かっていないためです（[`../samples/README.md`](../samples/README.md) 参照）。

## 値チェック

次に当てはまるとき、**エラーにはせず** `warnings` に文言を足します。
明らかにおかしい値を人が気付けるようにするためのものです。

| 条件 | |
|---|---|
| `duration_sec` が 0 以下 | 時刻が壊れている可能性 |
| `duration_sec` が 3600 秒超 | 1時間を超える飛行は想定外 |
| 平均速度（`distance_m / duration_sec`）が 25 m/s 超 | 座標か時刻が壊れている可能性 |
| `max_height_m` が 150 m 超 | 航空法の上限（150 m）を超えている |

## MVP の割り切り

### 経路は API キー発行後

DJI のログ本体（経路の時系列データ）は暗号化されていて、復号には DJI の API キーが要ります。
キー無しで読めるのは概要（開始時刻・飛行時間・距離・最高高度など）だけです。

MVP では陣地を経路ではなく**ポイントのクリック**で取る想定なので、概要だけで足ります。
研究室の API キーが発行されたら、経路を読む処理を `parse_level: "full"` として追加します。

### 改ざんは検出できない

**概要の値の改ざんは検出できません。**

- GPX はただの XML なので、時刻や座標を書き換えればポイントを水増しできます
- DJI も、概要だけでは本体データと照合できません

`warnings` の値チェックは「常識的な範囲に収まっているか」を見るだけで、改ざん対策ではありません。
API キー発行後に、概要と本体データ（経路）の照合を入れる予定です。

## テスト

```bash
go test ./...
```

- `internal/gpx` — `../samples/gpx/` の合成 GPX を使った変換テストと異常系
- `internal/djisummary` — `internal/djisummary/testdata/djirecord_dummy.json`（`djirecord --json` と
  同じ構造で、シリアルと座標を明らかなダミー値にしたもの）を使った変換テスト
- `internal/flightrecord` — 値チェックと HMAC のテスト

生ログを使う照合テスト（`TestParseRealLog`）は、`samples/private/` に生ログがあり、かつ
`DJIRECORD_BIN` が設定されているときだけ動きます。どちらか欠けていれば skip されるので、
生ログの無い環境（CI や他の人の手元）でも `go test ./...` は通ります。

```bash
# 生ログを使った照合も含めて実行する
DJIRECORD_BIN=~/.venvs/djirecord/bin/djirecord go test ./...
```

## ディレクトリ構成

変換処理と呼び出し方を分けてあります。

```
worker/
├── cmd/flightparse/      CLI（ファイルを読み、形式を判定して変換、JSON を出力）
└── internal/
    ├── flightrecord/     FlightRecord の型と値チェック
    ├── gpx/              GPX の読み込みと変換（encoding/xml）
    └── djisummary/       djirecord を呼び出して概要を変換
```

`djisummary` は、外部コマンドを実行する `Parse` と、その出力を変換する `Convert` に分けてあります。
`Convert` は JSON のバイト列を受け取るだけなので、djirecord 無しでテストできます。
