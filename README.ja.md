<p align="center">
  <img src="arcnem-logo.svg" alt="Arcnem AI" width="120" />
</p>

<h1 align="center">texvec</h1>

<p align="center">
  <strong>ローカルファーストのテキスト類似検索CLI。</strong>
</p>

<p align="center">
  <a href="README.md">English</a> ·
  <a href="#インストール">インストール</a> ·
  <a href="#クイックスタート">クイックスタート</a> ·
  <a href="#モデル">モデル</a> ·
  <a href="#開発">開発</a>
</p>

---

texvecは、テキストの類似検索のためのオープンソースCLIです。文書要約をONNXモデルでローカル実行し、要約埋め込みと本文チャンク埋め込みをローカルのlibsqlデータベースに保存し、コサイン距離で近い文書を見つけます。

Arcnem AIが開発しているtexvecは、私たちが好む実用的なAIツールの考え方を反映しています。ローカルファーストで、挙動を追いやすく、クラウド前提ではありません。

## こういう人に向いています

- 文書を手元に置いたまま類似検索したい
- テキスト要約とベクトル検索を小さなCLIとして扱いたい
- 外部のベクトルデータベースなしで試したい
- 文書全体の要旨と文書内の局所的な一致の両方を取りたい

## インストール

GitHub Releases からリリースアーカイブを取得するか、Goでインストールします。

```sh
go install github.com/arcnem-ai/texvec@latest
```

ソースからビルドする場合:

```sh
git clone https://github.com/arcnem-ai/texvec.git
cd texvec
go build -o texvec
```

リリースアーカイブは、`picvec` と同じく macOS (`arm64`) と Linux (`amd64`) を主な対象としています。

## クイックスタート

```sh
texvec init
texvec summarize test_texts/galaxies.md
texvec embed test_texts/galaxies.md
texvec search --text "dark matter in spiral galaxies"
texvec list
```

`texvec init` はONNX Runtimeをダウンロードし、`~/.texvec/` を作成し、ローカルデータベースを初期化し、デフォルトの要約モデルと埋め込みモデルを取得します。

## コマンド

| コマンド | 内容 |
|----------|------|
| `init` | ONNX Runtimeとデフォルトモデルをダウンロード |
| `summarize [document]` | 要約を生成して表示。データベースには書き込まない |
| `embed [document]` | 要約、チャンク化、埋め込み、保存をまとめて実行 |
| `search [document]` | 類似文書を検索 |
| `search --text "..."` | 生テキストから検索 |
| `list` | 登録済み文書を一覧表示 |
| `set-embedding-model [name]` | デフォルトの埋め込みモデルを変更 |
| `set-summary-model [name]` | デフォルトの要約モデルを変更 |
| `config` | 現在の設定を表示 |
| `clean` | `texvec` のデータをすべて削除 |

グローバルフラグ:

- `-v, --verbose` で詳細ログを表示

### よく使う例

要約を確認する:

```sh
texvec summarize notes.md
texvec summarize notes.md --summary-model flan-t5-small
```

`texvec summarize` はプレビュー用です。データベースには保存しません。

文書をインデックスする:

```sh
texvec embed notes.md
texvec embed notes.md -m bge-small-en-v1.5
texvec embed notes.md --summary-model flan-t5-small
```

文書の内容ハッシュが変わっていない場合は、既存の要約とチャンクデータを可能な範囲で再利用します。

類似文書を検索する:

```sh
texvec search notes.md
texvec search --text "barred spiral galaxy dark matter"
texvec search --text "barred spiral galaxy dark matter" -k 10
texvec search notes.md -m bge-small-en-v1.5
```

結果はコサイン距離順に並びます。検索対象の文書がすでにデータベースにある場合、その同じパスは結果から除外されます。

| フラグ | 内容 | デフォルト |
|--------|------|------------|
| `-k, --limit` | 返す件数 | 5 |
| `-m, --model` | 使用する埋め込みモデル | 設定値 |
| `--summary-model` | 長いクエリを要約するときの要約モデル | 設定値 |

登録済み文書を一覧表示する:

```sh
texvec list
texvec list -m all-minilm-l6-v2
texvec list -k 20
```

| フラグ | 内容 | デフォルト |
|--------|------|------------|
| `-k, --limit` | 表示上限 | すべて |
| `-m, --model` | 埋め込みモデルで絞り込み | すべて |

デフォルトモデルを変更する:

```sh
texvec set-embedding-model bge-small-en-v1.5
texvec set-summary-model flan-t5-small
```

## モデル

### 埋め込みモデル

| 名前 | 次元数 | メモ |
|------|--------|------|
| `all-minilm-l6-v2` | 384 | デフォルト。高速で汎用的。 |
| `bge-small-en-v1.5` | 384 | 検索向けで、クエリ用プレフィックスを持つ。 |

### 要約モデル

| 名前 | メモ |
|------|------|
| `flan-t5-small` | `1.0.0` のデフォルト要約モデル。小さく、ローカル実行しやすい。 |

モデルは初回利用時にHugging Faceからダウンロードされ、`~/.texvec/models/` に保存されます。

## 仕組み

1. `.txt`、`.md`、`.markdown` の文書を読み込みます。
2. 内容ハッシュを計算して、再インデックスが必要か判定します。
3. 選択した要約モデルで文書要約を生成します。
4. 選択した埋め込みモデルで要約と、本文の重なり付きチャンクを埋め込みます。
5. 検索では、クエリ埋め込みを要約埋め込みとチャンク埋め込みの両方に照合します。
6. それらのスコアを統合し、文書単位でコサイン距離順に返します。

## データ保存先

実行時データはすべて `~/.texvec/` に保存されます。

```text
~/.texvec/
  config.json       # デフォルトモデルなどの設定
  texvec.db         # libsqlデータベース
  models/           # ダウンロードしたONNXモデルとトークナイザ資産
  lib/              # ONNX Runtime共有ライブラリ
```

すべて削除したい場合は `texvec clean` を使います。

## リポジトリ構成

- `cmd/` Cobraコマンドとユーザー向け出力
- `core/` モデルレジストリ、ランタイム設定、ダウンロード、チャンク化、要約、埋め込み処理
- `store/` スキーマ作成、挿入、一覧取得、ベクトル検索クエリ
- `config/` `~/.texvec` のパス補助と設定初期化
- `test_texts/` 手動確認用のサンプル文書

## 対応プラットフォーム

| OS | 公開予定のリリース | ハードウェアアクセラレーション |
|----|--------------------|------------------------------|
| macOS | `arm64` | CPU |
| Linux | `amd64` | CPU |

texvecは、マシン間でCLI出力を安定させるため、現時点ではCPU実行をデフォルトにしています。

初回実行時にONNX Runtime 1.24.3が自動でダウンロードされます。

## 開発

```sh
go test ./...
go build -o texvec
```

コントリビュート手順は [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。AIエージェント向けのリポジトリ固有ルールは [AGENTS.md](AGENTS.md) にあります。

---

<p align="center">
  <a href="https://arcnem.ai">Arcnem AI</a> が開発。
</p>
