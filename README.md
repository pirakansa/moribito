<p align="center">
  <!-- ロゴ画像を配置（正方形アイコン推奨） -->
  <img src="./docs/moribito-logo.png" width="160" alt="MORIBITO logo"/>
</p>

<h1 align="center">M.O.R.I.B.I.T.O.</h1>
<p align="center">
  <strong>Monitoring Orchestrator for Repository Issues and Build Integrity; Trustee Officer.</strong><br>
  A GitHub App appointed to oversee issues, reviews, and entrusted operations within your repository.
</p>

---

## About M.O.R.I.B.I.T.O.
**M.O.R.I.B.I.T.O.** is a GitHub App that acts as a repository’s appointed “Trustee Officer.”  
It monitors activity, supports discussions, assists reviews, and can execute entrusted tasks on your behalf.


GolangでGitHub Appを作るための雛形です。標準の`net/http`だけでWebhook受信と認証の骨組みを提供します。

## Prerequisites

- Go 1.21+

## Setup

```bash
go mod tidy
```

## Docs

- `docs/ARCHITECTURE.md`
- `docs/CONFIG.md`
- `docs/DEVELOPMENT.md`
- `docs/OPERATIONS.md`

## Build

```bash
make build
```

## Run

```bash
go run ./cmd/moribito
```

起動後は`/healthz`と`/webhook`が利用できます。

## Test

```bash
make test
```

## GitHub Appの基本

- App ID: GitHub Appを識別するIDです。
- Private Key: JWT署名に使う秘密鍵です（App側で発行）。
- Installation ID: インストールされた組織/リポジトリを識別します。
- Webhook Secret: Webhookの署名検証に使います。
- API Base URL: GitHub Enterprise Server向けに`https://<host>/api/v3`へ変更できます。

## 雛形の狙い（何が入っているか）

- Webhook受信: `/webhook`でイベントを受け取り、署名検証を行います。
- 認証: App ID + Private KeyでJWTを作成し、Installation tokenを取得できます。
- ルーティング: Webhookイベントごとにハンドラを分離できる構成です。
- ジョブキュー: 長い処理を非同期化するための簡易キューを持ちます。
- 拡張ポイント: Issue/PRイベントへの対応やAPI呼び出しを追加しやすくしています。

## 環境変数

```
APP_ADDR=:8080
GITHUB_WEBHOOK_PATH=/webhook
GITHUB_APP_ID=123
GITHUB_INSTALLATION_ID=456
GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
GITHUB_WEBHOOK_SECRET=secret
GITHUB_API_BASE_URL=https://api.github.com
```

### 使い分けの目安

- `GITHUB_APP_ID` / `GITHUB_PRIVATE_KEY_PATH`: **JWT作成**に必須。
- `GITHUB_INSTALLATION_ID`: **Installation token取得**に必須。
- `GITHUB_WEBHOOK_SECRET`: **Webhook署名検証**に使用（未設定なら検証をスキップ）。
- `GITHUB_API_BASE_URL`: **Enterprise Server**のときは必須。

## ローカル開発の流れ（例）

1. GitHub Appを作成し、Webhook URLを`http://localhost:8080/webhook`に設定します。
2. `GITHUB_PRIVATE_KEY_PATH`に秘密鍵を保存します。
3. `go run ./cmd/moribito`で起動します。

注: 本リポジトリのWebhook処理は骨組みのみです。イベント内容の処理やAPI呼び出しは今後追加します。

## smeeでWebhookをローカルに中継

1. GitHub AppのWebhook URLに`smee.io`のURLを設定します。
2. smeeクライアントを起動します。

```bash
npx -p smee-client@1.0.2 smee -u https://smee.io/<your-id> -p 8080 -P /webhook
```

3. アプリを起動します。

```bash
go run ./cmd/moribito
```

注: Node 18環境で動かない場合があるため、`smee-client@1.0.2`を指定しています。

## Installation tokenの取得（動作確認）

以下の環境変数を設定し、トークンを表示します。

```
export GITHUB_APP_ID=123
export GITHUB_INSTALLATION_ID=456
export GITHUB_PRIVATE_KEY_PATH=/path/to/private-key.pem
export GITHUB_API_BASE_URL=https://api.github.com
```

```bash
go run ./cmd/moribito --print-installation-token
```

注: これはローカル検証向けの機能です。運用時はログや出力先の扱いに注意してください。

## 認証フロー図（概要）

```mermaid
sequenceDiagram
    participant GitHub as GitHub
    participant App as App Server
    participant API as GitHub API

    GitHub->>App: Webhook (installation / installation_repositories)
    App->>App: Installation IDを取得
    App->>App: App ID + Private KeyでJWT生成
    App->>API: JWTでInstallation tokenを要求
    API-->>App: Installation token
    App->>API: TokenでPR/コメント等のAPI実行
```

## テスト方針

- `internal/config`: 環境変数の読み取りとバリデーション
- `internal/githubapp`: JWT生成、Webhook署名、Installation token取得、キャッシュ
- `internal/webhook`: イベントルーティングと基本のパース

### Fixtures

Webhookのテスト用payloadは`test/fixtures/webhook/`に配置しています。
イベント追加時は対応するfixtureを追加してください。

## 観測性（ログ）

- `X-GitHub-Delivery` と `X-GitHub-Request-Id` をログに出力します。
- エラー時も同じIDを出すため、再送判断やトラブルシュートがしやすくなります。

ローカル検証が必要な場合は、`smee`でWebhookを中継し実イベントで挙動確認できます。
