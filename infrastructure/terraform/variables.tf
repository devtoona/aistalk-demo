# variable: 入力変数。terraform.tfvars や -var で上書き可能。
variable "project_id" {
  description = "GCP / Firebase project ID（同一であること）"
  type        = string
}

variable "region" {
  description = "GCP region (e.g. asia-northeast1)"
  type        = string
  default     = "asia-northeast1"
}

variable "firestore_location" {
  description = "Firestore location (multi-region: nam5/eur3, or regional e.g. asia-northeast1)"
  type        = string
  default     = "asia-northeast1"
}

variable "connection" {
  description = "Cloud Build connection resource name"
  type        = string
}

variable "aistalk_repository" {
  description = "aistalk repository resource name"
  type        = string
}

variable "domain" {
  description = "Frontend domain. Used for SSL cert, HTTPS, and Firebase authorized domains."
  type        = string
}

variable "app_name" {
  description = "NEXT_PUBLIC_APP_NAME"
  type        = string
  default     = "AISTalk"
}

# false: Secret の箱 + SA + Firestore/Auth。true: Cloud Run / AR / Build トリガー
variable "enable_app" {
  description = "Cloud Run 等のアプリ基盤を作成する。初回は false → Secret 投入後に true"
  type        = bool
  default     = false
}

variable "backend_request_timeout_seconds" {
  description = "Cloud Run request timeout (SSE 用に長め推奨)"
  type        = number
  default     = 3600
}

variable "backend_build_branch" {
  description = "Git branch that triggers backend Cloud Build"
  type        = string
  default     = "^main$"
}

variable "backend_bootstrap_image" {
  description = "Placeholder image until first cloudbuild-backend run"
  type        = string
  default     = "us-docker.pkg.dev/cloudrun/container/hello"
}

variable "quota_daily_chat_limit" {
  description = "Daily chat API quota per anonymous UID"
  type        = number
  default     = 30
}

variable "quota_daily_tts_limit" {
  description = "Daily TTS API quota per anonymous UID"
  type        = number
  default     = 30
}

variable "quota_disabled" {
  description = "Quota disabled (QUOTA_DISABLED)"
  type        = bool
  default     = false
}

# Firebase Web アプリ（Console で作成後に記入。クライアント公開設定）
variable "firebase_api_key" {
  description = "Firebase Web API key (NEXT_PUBLIC_FIREBASE_API_KEY)"
  type        = string
  default     = ""
}

variable "firebase_auth_domain" {
  description = "Firebase authDomain。空なら {project_id}.firebaseapp.com"
  type        = string
  default     = ""
}

variable "firebase_app_id" {
  description = "Firebase Web appId (NEXT_PUBLIC_FIREBASE_APP_ID)"
  type        = string
  default     = ""
}

variable "firebase_extra_authorized_domains" {
  description = "Identity Platform authorized_domains に追加するドメイン（domain / localhost 以外）"
  type        = list(string)
  default     = []
}
