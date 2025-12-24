# pghapp

GolangでGitHub Appを作るための雛形です。標準の`net/http`だけでWebhook受信と認証の骨組みを提供します。

## Prerequisites

- Go 1.21+

## Setup

```bash
go mod tidy
```

## Build

```bash
make build
```

## Run

```bash
go run ./cmd/app
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

## ローカル開発の流れ（例）

1. GitHub Appを作成し、Webhook URLを`http://localhost:8080/webhook`に設定します。
2. `GITHUB_PRIVATE_KEY_PATH`に秘密鍵を保存します。
3. `go run ./cmd/app`で起動します。

注: 本リポジトリのWebhook処理は骨組みのみです。イベント内容の処理やAPI呼び出しは今後追加します。

## smeeでWebhookをローカルに中継

1. GitHub AppのWebhook URLに`smee.io`のURLを設定します。
2. smeeクライアントを起動します。

```bash
npx -p smee-client@1.0.2 smee -u https://smee.io/<your-id> -p 8080 -P /webhook
```

3. アプリを起動します。

```bash
go run ./cmd/app
```

注: Node 18環境で動かない場合があるため、`smee-client@1.0.2`を指定しています。
