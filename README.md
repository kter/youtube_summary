# YouTube Hashtag Summary Service

特定のハッシュタグがついたYouTube動画を定期的にスキャンし、Gemini 2.5 Flashを用いて要約を生成・表示するWebサービスです。

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────────┐
│                         CloudFront                              │
│  youtube-summarize.dev.devtools.site                           │
│  youtube-summarize.devtools.site                                │
└──────────────┬─────────────────────────────────────┬───────────┘
               │                                     │
               ▼                                     ▼
        ┌─────────────┐                    ┌────────────────┐
        │   S3 (SPA)  │                    │  API Gateway   │
        │   React     │                    │  HTTP API      │
        └─────────────┘                    └───────┬────────┘
                                                   │
                                                   ▼
                                          ┌────────────────┐
                                          │  Lambda (API)  │
                                          └───────┬────────┘
                                                   │
        ┌──────────────────────────────────────────┼──────────────┐
        │                                          │              │
        ▼                                          ▼              │
┌───────────────┐                         ┌────────────────┐      │
│  EventBridge  │──(3h)───────────────▶   │ Lambda (Batch) │      │
│  (Scheduler)  │                         └───────┬────────┘      │
└───────────────┘                                 │               │
                                                  ▼               │
                           ┌─────────────────────────────────────┴─────┐
                           │              DynamoDB                      │
                           │  PK: hashtag  SK: processedAt             │
                           └───────────────────────────────────────────┘
                                                  ▲
                                                  │
        ┌─────────────────────────────────────────┴─────────────────────┐
        │                                                               │
        ▼                              ▼                                ▼
┌───────────────┐           ┌─────────────────┐              ┌─────────────────┐
│ YouTube API   │           │ Transcript API  │              │  Gemini 2.5     │
│  v3           │           │ (字幕取得)       │              │  Flash          │
└───────────────┘           └─────────────────┘              └─────────────────┘
```

## 前提条件

- AWS CLI (プロファイル `dev` / `prd` 設定済み)
- Terraform >= 1.0.0
- Node.js >= 18
- Python >= 3.12
- make

## セットアップ

### 1. APIキーの準備と登録

本サービスでは、以下の2つのAPIキーが必要です。

#### A. YouTube Data API v3 Key (Google Cloud)

1. [Google Cloud Console](https://console.cloud.google.com/) にアクセスし、プロジェクトを作成（または選択）します。
2. 左側のメニューから「APIとサービス」>「ライブラリ」を選択します。
3. "YouTube Data API v3" を検索し、「有効にする」をクリックします。
4. 「認証情報」>「認証情報を作成」>「APIキー」を選択します。
5. 作成されたAPIキーをコピーします。

#### B. Gemini API Key (Google AI Studio)

1. [Google AI Studio](https://aistudio.google.com/) にアクセスします。
2. "Get API key" をクリックします。
3. "Create API key" をクリックし、Google Cloudプロジェクトに関連付けてキーを作成します。
4. 作成されたAPIキーをコピーします。

#### C. Secrets Manager への登録

取得したAPIキーを AWS Secrets Manager に登録します。以下のコマンドの `YOUR_...` 部分を実際のキーに置き換えて実行してください。

**Development環境 (`dev` プロファイル)**

```bash
# YouTube API Key
aws secretsmanager create-secret \
  --name youtube-summary/youtube-api-key \
  --secret-string "YOUR_YOUTUBE_API_KEY" \
  --profile dev

# Gemini API Key
aws secretsmanager create-secret \
  --name youtube-summary/gemini-api-key \
  --secret-string "YOUR_GEMINI_API_KEY" \
  --profile dev
```

**Production環境 (`prd` プロファイル)**

```bash
# YouTube API Key
aws secretsmanager create-secret \
  --name youtube-summary/youtube-api-key \
  --secret-string "YOUR_YOUTUBE_API_KEY" \
  --profile prd

# Gemini API Key
aws secretsmanager create-secret \
  --name youtube-summary/gemini-api-key \
  --secret-string "YOUR_GEMINI_API_KEY" \
  --profile prd
```

### 2. Terraform State バケットの作成

```bash
aws s3 mb s3://terraform-state-youtube-summary --region ap-northeast-1 --profile dev
aws dynamodb create-table \
  --table-name terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --profile dev
```



## デプロイ手順

### Development 環境

```bash
# Lambda Layer のビルド
make build-layer

# Terraform 初期化
make init-dev

# プラン確認
make plan-dev

# インフラデプロイ
make apply-dev

# フロントエンドビルド
make build-frontend

# フロントエンドデプロイ
make deploy-frontend-dev
```

### Production 環境

```bash
# Lambda Layer のビルド
make build-layer

# Terraform 初期化
make init-prd

# プラン確認
make plan-prd

# インフラデプロイ
make apply-prd

# フロントエンドビルド
make build-frontend

# フロントエンドデプロイ
make deploy-frontend-prd
```

## ローカル開発

### フロントエンド

```bash
cd frontend
npm install
npm run dev
```

ブラウザで http://localhost:5173 を開きます。

### Lambda 関数のテスト

```bash
# バッチ処理のローカル実行（要 API キー設定）
cd backend/batch
python -c "from handler import lambda_handler; print(lambda_handler({}, None))"
```

## 設定変更

### ハッシュタグの変更

`terraform/variables.tf` の `hashtags` 変数を編集：

```hcl
variable "hashtags" {
  default = ["プログラミング", "エンジニア", "Python", "AI"]
}
```

### フィルタリング閾値の変更

```hcl
variable "min_view_count" {
  default = 1000  # 最低再生数
}

variable "min_like_count" {
  default = 50    # 最低高評価数
}
```

## Makefile コマンド一覧

| コマンド | 説明 |
|---------|------|
| `make init-dev` | Terraform 初期化 (dev) |
| `make init-prd` | Terraform 初期化 (prd) |
| `make plan-dev` | Terraform plan (dev) |
| `make plan-prd` | Terraform plan (prd) |
| `make apply-dev` | Terraform apply (dev) |
| `make apply-prd` | Terraform apply (prd) |
| `make destroy-dev` | Terraform destroy (dev) |
| `make destroy-prd` | Terraform destroy (prd) |
| `make build-layer` | Lambda Layer ビルド |
| `make build-frontend` | フロントエンドビルド |
| `make deploy-frontend-dev` | S3 へデプロイ (dev) |
| `make deploy-frontend-prd` | S3 へデプロイ (prd) |
| `make invoke-batch-dev` | バッチ Lambda 手動実行 (dev) |
| `make logs-batch-dev` | バッチ Lambda ログ確認 (dev) |

## トラブルシューティング

### 字幕が取得できない動画がスキップされる

YouTube の自動生成字幕が無効な動画や、字幕がアップロードされていない動画はスキップされます。これは仕様通りの動作です。

### API Gateway でCORSエラーが発生する

CloudFront 経由でアクセスしているか確認してください。ローカル開発時は Vite のプロキシ機能を使用します。

### DynamoDB にデータが保存されない

1. Lambda の実行ログを確認: `make logs-batch-dev`
2. Secrets Manager に API キーが正しく設定されているか確認
3. IAM ロールに必要な権限があるか確認

## ライセンス

MIT License
