# samples — 飛行ログのサンプル

GPX / DJI ログのパーサー開発とテストに使うサンプルを置く場所です。

> [!CAUTION]
> **このリポジトリは public です。** DJI の生ログ（`FlightRecord_*.txt` / `DJIFlightRecord_*.txt`）には
> 機体シリアル番号と離陸地点の座標が含まれます。生ログ、および機体シリアル・緯度経度・離陸地点の高度の
> 実値を、コミット・PR・Issue のどこにも置かないでください。

## 構成

| ディレクトリ | 内容 | コミット |
|---|---|---|
| `gpx/` | 合成 GPX（手で作ったダミーデータ） | ✅ する |
| `dji/` | DJI ログから抽出した**概要だけ**の JSON | ✅ する |
| `private/` | DJI の生ログ置き場 | ❌ `.gitignore` 済み |

---

## `gpx/` — 合成 GPX

実際の飛行ログではありません。すべて手で作ったダミーデータで、座標は 35.0°N / 135.0°E ちょうどを
起点にした明らかな合成値です。実在の飛行経路・離陸地点ではありません。

### `synthetic_normal.gpx`

正常系。13 点、10 秒間隔。東へ 6 区間 → 北へ 6 区間 の L 字の経路。

| 項目 | 期待値 |
|---|---|
| `duration_sec` | `120` |
| `distance_m` | `121.368…`（haversine, R=6371000 m） |
| `max_height_m` | `30.0`（最初の点の `ele=100.0` を 0 とした相対高度の最大） |
| 平均速度 | 約 1.01 m/s（警告の閾値 25 m/s 未満） |
| `warnings` | 空 |

### `synthetic_no_time.gpx`

異常系。`<time>` を持たない GPX です。パーサーがエラーを返すことの確認に使います。

---

## `dji/` — DJI ログの概要 JSON

生ログから**概要部分だけ**を抜き出したものです。機体シリアルと座標は含めていません。

### `neo_2026-06-20_summary.json`

DJI Neo の飛行（UTC 2026-06-20 23:51:37 開始 / JST 2026-06-21 08:51:37）の概要。

| 項目 | 値 |
|---|---|
| `duration_sec` | `241.0` |
| `distance_m` | `103.23981940746307` |
| `max_height_m` | `5.0`（離陸地点からの相対高度） |
| `max_horizontal_speed_mps` | `1.1045360565185547` |
| `record_count` | `1894` |

小数は**丸めずに djirecord の出力値をそのまま**入れてあります。パーサーのテストの正解データとして
使うためです。`parser` に使用したツールのバージョン、`source_sha256` に元の生ログの SHA-256 を
記録してあります（ハッシュから中身は復元できないので公開して問題ありません）。

経路（トラックポイント）は入っていません。DJI ログの本体は暗号化されており、復号には
DJI の API キーが必要なためです。MVP では陣地を経路ではなくポイントのクリックで取る想定なので、
概要だけで足ります。経路は研究室の API キー発行後に `parse_level: full` として追加します。

### 概要の値について（`video_time` の注意）

概要に含まれる値がすべて正確とは限りません。`samples/private/` に置いた DJI Neo の生ログでは、
概要の `video_time` が **1552 秒（25 分 52 秒）** で、同じ概要の飛行時間 **241 秒** より長い値でした。
原因は分かっていません（録画の扱いや集計方法の違いの可能性があります）。

このため FlightRecord では、飛行時間・距離・最高高度など**必要な項目だけ**を使い、
`video_time` は使いません。

---

## `private/` — DJI の生ログ

`.gitignore` で除外済みです。手元で動作確認するときだけ、ここに生ログを置いてください。

```
samples/private/FlightRecord_YYYY-MM-DD_HH-MM-SS.txt
```

`.gitignore` では次の 3 つを除外しています。`private/` の外にうっかり置いた場合も拾えるようにしています。

```gitignore
samples/private/
FlightRecord_*.txt
DJIFlightRecord_*.txt
```

---

## djirecord（pydjirecord）のインストール

DJI ログの概要を読むのに [pydjirecord](https://pypi.org/project/pydjirecord/) を使います。
リポジトリ内に venv を作らないよう、ホーム配下の `~/.venvs/djirecord` に入れてください。

```bash
python3 -m venv ~/.venvs/djirecord
~/.venvs/djirecord/bin/pip install pydjirecord==1.3.0
```

インストールできたか確認します。

```bash
~/.venvs/djirecord/bin/djirecord --help
```

生ログの概要を JSON で見る場合:

```bash
~/.venvs/djirecord/bin/djirecord --json samples/private/FlightRecord_2026-06-20_xx-xx-xx.txt
```

> [!WARNING]
> この出力には**機体シリアル番号と離陸地点の座標がそのまま入ります。**
> 出力をファイルに保存してコミットしたり、Issue / PR に貼ったりしないでください。
> リポジトリに残すのは `dji/*_summary.json` のように、シリアルと座標を取り除いたものだけです。

パーサー（`worker/`）からは環境変数 `DJIRECORD_BIN` でこのパスを指定します。

```bash
export DJIRECORD_BIN=~/.venvs/djirecord/bin/djirecord
```

## 関連 Issue

- #3 公開してよいサンプル GPX を用意する
- #6 FlightRecord の項目を決める
