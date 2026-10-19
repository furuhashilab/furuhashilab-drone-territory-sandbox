# 開発への参加方法（CONTRIBUTING）

Drone Territory の共同開発ルールです。迷ったら火曜日に相談してください。

## 基本の流れ

```
Issue → Branch → 開発 → Commit → Push → Pull Request → Review → Merge
```

1. **Issue** を確認する（なければ作る）。自分が担当する Issue には Assignee を設定する
2. 最新の `main` から **Branch** を作る
3. Branch 上で開発し、こまめに **Commit / Push** する
4. **Pull Request**（PR）を作り、本文に `Closes #Issue番号` を書く
5. 1人以上の **Review**（Approve）をもらう
6. **Merge** する（基本は火曜日にまとめて確認）

`main` へ直接 Push はしません。

## Branch 命名ルール

```
<種類>/<内容>
```

| 種類 | 用途 | 例 |
| --- | --- | --- |
| `feature/` | 新機能 | `feature/map-grid`, `feature/gpx-parser` |
| `fix/` | バグ修正 | `fix/grid-click-offset` |
| `docs/` | README・仕様書などドキュメントのみ | `docs/flight-record-spec` |
| `chore/` | 設定・依存関係・リポジトリ整備 | `chore/project-scaffold` |

- 英小文字とハイフン（`-`）のみ使う
- 内容は短く、何をするか分かる名前にする
- Issue 番号を入れたい場合は `feature/12-map-grid` のように書いてもよい

```bash
git switch main
git pull
git switch -c feature/map-grid
```

## Commit メッセージ

厳密な形式は決めませんが、何をしたか分かるように書いてください。

```
Grid をクリックしたときに座標を取得する処理を追加
```

## Pull Request のルール

- **1つの PR は1つの目的**。複数の機能を1つの PR に混ぜない
- できるだけ小さく（目安：レビューが15分で終わる量）
- PR テンプレートの項目を埋める
- 作業途中で見てほしいときは **Draft PR** にする
- `main` の変更を取り込んで、コンフリクトを解消してから Review を依頼する

### PR タイトル

```
[担当領域] 変更内容
```

例：`[map] Grid 表示を追加`、`[flight-log] GPX Parser を追加`、`[docs] FlightRecord の仕様を追加`

## Review で確認すること

- Issue の目的に合っているか
- 他の担当の機能を壊していないか（`main` で動いていたものが動くか）
- データ形式（特に `FlightRecord`）が仕様と一致しているか
- README や `docs/` の更新が必要ではないか
- パスワード・APIキー・個人の飛行ログなど、公開してはいけないものが含まれていないか
- Merge して問題ないか

気になる点はコメントで伝え、直してほしい場合は「Request changes」、軽い提案なら「Comment」を使います。

## 秘密情報の扱い

- Supabase のキーなどは `.env` に書き、**Git に含めない**（`.env.example` に項目名だけ書く）
- 実際の飛行ログには位置情報が含まれます。リポジトリに置くサンプルは、公開してよいものだけにする

## 新しい機能のアイデア

開発中に「この機能があったら面白そう」と思ったら、すぐに実装せず、

1. `Feature proposal` テンプレートで Issue を作る
2. 火曜日に共有し、次の観点で議論する
   - 面白いか / 技術的に可能か / 工数はどれくらいか
   - v1.0 の完成を遅らせないか
   - 今年実装するか、来年度へ回すか
3. 実装する → Label `feature`、Milestone `v1.0` などを付けて通常の Issue にする
4. 今はやらない → Label `next-year` を付けて残す

## 火曜日の進め方

1. 現在 `main` で何が動くか確認
2. 各担当の進捗共有
3. Pull Request 確認
4. Review
5. Merge
6. 困っていること共有
7. 新しい機能案（`feature proposal`）の議論
8. 次週の Issue 決定・担当割り当て
