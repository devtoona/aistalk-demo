# Infrastructure（デモ公開）

GCP + Firebase + Cloud Run / Cloud Build / LB で AISTalk デモを立てる手順。

## 前提

- GCP プロジェクト ID = Firebase プロジェクト ID
- Firebase は **「既存の Google Cloud プロジェクトに追加」** で紐づける（新規作成だと別 ID になる）
- Cloud Build 用の GitHub 接続（Cloud Build 接続 + リポジトリリンク）が済んでいること

## セットアップ手順

### 1. GCP / Firebase

1. GCP プロジェクトを作成する
2. Firebase Console →「既存の Google Cloud プロジェクトに Firebase を追加」→ `project_id` を選択

### 2. tfvars 準備

```bash
cd infrastructure
cp terraform/terraform.tfvars.example terraform/terraform.tfvars
# project_id / domain / connection / aistalk_repository などを記入
# enable_app は最初 false のまま
```

`terraform.tfvars` は gitignore 済み（秘密・環境固有値用）。

### 3. 基盤 apply（アプリはまだ上げない）

```bash
make terraform-init
make terraform-apply   # enable_app=false
```

作られるもの（抜粋）:

- Firestore + rules（クライアント全 deny）
- Identity Platform: 匿名認証 ON、authorized domains（`localhost` / `{project}.firebaseapp.com` / `domain`）
- Secret Manager の箱、SA、LB、buckets、Cloud Build トリガー（substitutions 含む）

`terraform output lb_ip_address` をメモする。

ADC で Identity Toolkit が quota project を要求したら:

```bash
gcloud auth application-default set-quota-project YOUR_PROJECT_ID
```

### 4. Firebase Web アプリ + Secret

1. Firebase Console（**同じ project_id**）で Web アプリを作成
2. `firebase_api_key` / `firebase_app_id` を tfvars に記入して再 `make terraform-apply`（トリガー substitutions 更新）
3. Secret Manager に値を投入
   - `aistalk-openai-api-key`
   - `aistalk-aivis-api-key`

### 5. Unity / VRM（手動）

Unity Build は有料アセット等の都合でリポジトリに含めない。GCS へ手動アップロードする。

```bash
# Unity WebGL（必須）
gsutil -m cp -r path/to/unity/* \
  gs://$(cd terraform && terraform output -raw unity_build_bucket)/latest/unity/

# VRM（任意）
gsutil cp avatar.vrm \
  gs://$(cd terraform && terraform output -raw storage_bucket)/models/avatar.vrm
```

### 6. アプリ有効化

1. tfvars で `enable_app = true`
2. `make terraform-apply`
3. `make backend-build`
4. フロントは Cloud Build（`main` push または手動）

### 7. DNS

`terraform output lb_ip_address` をドメインの **A レコード** に設定する。SSL 証明書の発行完了を待つ。

### 8. 確認

```bash
curl -I https://YOUR_DOMAIN
curl https://BACKEND_URL/healthz
```

## Make ターゲット

| コマンド | 内容 |
|----------|------|
| `make terraform-init` | `terraform init` |
| `make terraform-plan` | plan（tfvars 必須） |
| `make terraform-apply` | apply（tfvars 必須） |
| `make terraform-destroy` | destroy |
| `make backend-build` | バックエンド Cloud Build → Cloud Run |

## git に含めないもの

`infrastructure/.gitignore` 参照:

- `terraform/terraform.tfvars`
- `terraform/.terraform/`
- `terraform/*.tfstate` / `*.tfstate.*`

`.terraform.lock.hcl` は含めてよい（プロバイダ版ロック）
