# Drone Territory（sandbox）

> 飛ぶほど、自分たちの領土が広がる。

> [!NOTE]
> これは **練習用の仮リポジトリ（sandbox）** です。Branch・PR・Review の流れを試すために使います。
> 正式なリポジトリ（`furuhashilab/drone-territory`）を作成したら、内容を移行します。

古橋研究室ドローン部の後期活動として開発している、ドローン練習をゲーム化するWebアプリです。

## Drone Territory とは

ドローンを飛ばした記録（飛行ログ）をポイントに変換し、そのポイントで地図上の陣地を取り合う2チーム対戦ゲームです。

```
ドローンを飛ばす → 飛行ログを登録 → 飛行時間などを解析 → ポイント化 → 陣地を取得 → 2チームで競う
```

## 開発目的

- ドローン部の飛行練習を継続しやすくする（練習のモチベーションをゲームで支える）
- 飛行ログ・地図・Webアプリ開発をチームで経験する
- 来年度のメンバーが開発を再開できる形で、コードと知見をGitHubに残す

## 今年度の目標（v1.0）

ドローン部7人で実際に遊べる状態にすること。最低限、次のゲームループが動くことを目指します。

```
飛行ログ → 解析 → ポイント → 陣取り → 状態保存
```

3D Voxel / Spatial ID は将来の挑戦とし、v1.0 はそれらに依存しない構成にします。

## 技術構成

| 領域 | 使用技術 | 状態 |
| --- | --- | --- |
| 地図・陣取り表示 | MapLibre GL JS | 決定 |
| 認証・データ保存 | Supabase（Auth / Database） | 決定 |
| フロントエンド | **未決定**（Issue #TBD で議論） | 未決定 |
| ホスティング | **未決定** | 未決定 |

### 飛行ログの扱い

ゲーム側はログの種類（GPX / DJI）を意識せず、共通の `FlightRecord` を受け取る構成を目指します。

```
GPX     → GPX Parser → FlightRecord ┐
DJI Log → DJI Parser → FlightRecord ┴→ ポイント計算 → 陣取り
```

`FlightRecord` の具体的な項目は **未決定** です（Issue #TBD）。決定後に `docs/flight-record.md` へまとめます。

## リポジトリ構成

```
drone-territory/
├── README.md            # このファイル
├── CONTRIBUTING.md      # 開発ルール（Issue / Branch / PR）
├── .github/             # Issue・PRテンプレート
├── docs/                # 仕様・設計メモ・引き継ぎ資料
├── src/                 # アプリのソースコード
├── public/              # 静的ファイル
└── supabase/            # Supabase の設定・マイグレーション
```

`src/` の中身はフロントエンド技術決定後に整理します。

## 開発環境構築

> フロントエンドの技術スタック決定後に記載します。

## 起動方法

> フロントエンドの技術スタック決定後に記載します。

## 担当

| 担当 | 内容 | メンバー | Label |
| --- | --- | --- | --- |
| ① MapLibre・陣取りマップ | MapLibre / Grid / Territory / チーム色表示 / 将来の3D Voxel | Tomoya, Reara | `map` |
| ② DJIログ・飛行時間データ | GPX / DJIログ / Parser / FlightRecord | Rito, Kanoko | `flight-log` |
| ③ UI・ゲームルール | ログアップロードUI / ゲーム画面 / ポイントルール / チームバランス | Aki, Yuto | `ui-game` |
| ④ GitHub・README・全体統合 | Repository / Issue / PR / Review / Merge / Supabase接続 / 引き継ぎ | Kosaku, Tomoya | `infra` |

## はじめての作業（Branch を切って PR を出すまで）

1. [Issues](../../issues) から自分の担当の Issue を開き、Assignees に自分を設定する
2. リポジトリを手元に持ってくる（初回のみ）

   ```bash
   git clone https://github.com/furuhashilab/uruhashilab-drone-territory-sandbox.git
   cd uruhashilab-drone-territory-sandbox
   ```

3. 最新の main から Branch を作る（Branch 名は Issue に書いてあります）

   ```bash
   git switch main
   git pull
   git switch -c feature/map-display
   ```

4. 作業して Commit・Push する

   ```bash
   git add .
   git commit -m "MapLibre で地図を表示"
   git push -u origin feature/map-display
   ```

5. GitHub 上に出る「Compare & pull request」から PR を作り、本文に `Closes #Issue番号` を書く
6. 火曜日に Review → Merge

## GitHub 開発フロー

```
Issue → Branch → 開発 → Commit → Push → Pull Request → Review → Merge
```

- `main` へ直接 Push しない
- 1つの Issue につき1つの Branch、1つの PR は1つの目的
- 毎週火曜日に進捗確認・PRレビュー・Merge・次週のIssue決定を行う

詳しくは [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。

## 実装済み機能

まだありません。

## 未実装機能

GitHub Issues を参照してください。

- v1.0 で実装するもの：Milestone `v1.0`
- 来年度以降に回したもの：Label `next-year`
- 提案段階のもの：Label `feature proposal`

## 来年度への引き継ぎ

> 年度末に記入します。最低限、次を残す予定です。
>
> - 現在 main で動く機能
> - 未完了の Issue と、その背景
> - Supabase など外部サービスの管理者・アクセス方法（パスワードやキーは書かない）
> - 開発中に分かったこと・ハマったこと

## メンバー

古橋研究室 ドローン部（2026年度）
