# sidetable SPEC (Draft)

## 1. 目的・スコープ

- プロジェクト直下の私的領域（例: `.sushichan044/`）を、外部コマンドに安全・簡潔に委譲するための CLI を提供する。
- サブコマンドは基本的に設定ファイルで定義し、`sidetable <subcmd> ...` を任意のコマンドに委譲できる。
- 本フェーズのスコープは **delegate + config** の最小実装。
- Completion、権限制御の詳細、安全強化は後続フェーズ。

## 2. 用語

- **subcmd**: `sidetable <subcmd> ...` の `<subcmd>` 部分。
- **delegate**: subcmd を外部コマンドへ引き渡して実行すること。

## 3. 設定ファイル仕様

### 3.1 パス解決

- 設定ファイルは `SIDETABLE_CONFIG_DIR/config.{yml,yaml}`。
  - `SIDETABLE_CONFIG_DIR` が未設定の場合は `XDG_CONFIG_HOME/sidetable` を利用する。
  - `XDG_CONFIG_HOME` が未設定の場合は `~/.config` を利用する。
  - 拡張子は `yml` または `yaml` のみ許容する。
  - 両方存在する場合はエラーとする。

### 3.2 フォーマット

YAML を採用する。

### 3.3 スキーマ (草案)

```yaml
directory: ".sushichan044"
commands:
  ghq:
    command: "ghq"
    args:
      prepend:
        - "--root={{.CommandDir}}"
      append:
        - "get"
    env:
      GHQ_ROOT: ".sushichan044/ghq"
    description: "ghq wrapper"
    alias: "q"
```

#### top-level

- `directory` (required): プロジェクト内の私的領域のディレクトリ名。
  - 必須指定（デフォルトなし）。
  - cwd からの相対パスとして解決する。

#### commands.<name>

- `command` (required): 委譲先の実行ファイル名を Go template で組み立てる文字列。
- `args` (optional): 追加の引数を Go template で組み立てる設定。
  - `args.prepend`: userArgs の前に追加する配列。
  - `args.append`: userArgs の後に追加する配列。
  - 何も指定していない場合は `sidetable <subcmd> ...` の `...` 部分をそのまま引き渡す。
- `env` (optional): 追加/上書きする環境変数の map。
- `description` (optional): `list` コマンドで表示する説明。
- `alias` (optional): 1つだけ許容する別名。`sidetable <alias>` は `<name>` と同義。

### 3.5 テンプレート展開

- `command` / `args` / `env` / `description` の各文字列は Go template として評価する。
- テンプレートは **文字列単位** で評価する（配列全体の結合などはしない）。
- 未定義キーはエラーにする（`missingkey=error`）。

#### テンプレート変数

- `.ProjectDir`: `sidetable` 実行時のカレントディレクトリ（絶対パス）
- `.PrivateDir`: `.ProjectDir/<directory>`（絶対パス）
- `.CommandDir`: `.PrivateDir/<commandName>`（絶対パス）
- `.ConfigDir`: `config.yml` が存在するディレクトリ（絶対パス）

例:

```yaml
commands:
  ghq:
    command: "ghq"
    args:
      prepend:
        - "get"
    env:
      GHQ_ROOT: "{{.CommandDir}}"
      CONFIG_DIR: "{{.ConfigDir}}"
```

### 3.4 バリデーション

- `commands` が存在しない場合はエラー。
- 各 command は `command` を必須とする。
- `alias` が重複する場合はエラー。
- `alias` が既存の `commands.<name>` と衝突する場合はエラー。
- `command` が空文字の場合はエラー。

## 4. CLI 仕様

### 4.1 グローバルフラグ

- `--version`: バージョン表示。
- `--help`: ヘルプ表示。

### 4.2 組み込みサブコマンド

- `list`: 設定済みの subcmd 一覧を表示する。
- `help`: 標準のヘルプ表示（cobra 等に委譲可）。
- `version`: バージョン表示。

### 4.3 delegate 対象の解決

- `sidetable <subcmd> ...` で `<subcmd>` を解決する。
- 解決順:
  1. `commands.<name>` の直指定
  2. `commands.<name>` の `alias`
- それ以外は **未解決** とし、エラー終了する。

## 5. delegate 仕様

### 5.1 実行コマンド組み立て

- `command` をテンプレート評価し、実行ファイル名として利用する。
- `args` が指定されている場合:
  - `args.prepend` / `args.append` の各要素をテンプレート評価する。
  - 実行引数は `prepend + userArgs + append` の順で組み立てる。
- `args` が指定されていない場合:
  - `exec.Command(command, userArgs...)` として実行する。

### 5.2 実行環境

- `env` は親プロセス環境を引き継いだ上で、`env` の値で上書きする。
- 標準入出力は `stdin/stdout/stderr` をそのまま透過する。
- `exec.Command` を用いて直接実行する。

### 5.3 終了コード

- 委譲先の終了コードをそのまま返す。
- 直接起動に失敗した場合は `1` を返す。

## 6. エラーハンドリング

- 設定ファイルが読めない/パースできない場合はエラー終了。
- `subcmd` が未解決の場合はエラー終了。
- `list` 実行時に設定ファイルが壊れている場合は、理由を表示してエラー終了。

## 7. 非スコープ

- Completion の実装（bash/zsh/fish は将来対応）
- セキュリティ強化（許可リスト、cwd 制限、署名等）
