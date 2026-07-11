locals {
  # GCP project_id = Firebase project ID（既存 GCP に Firebase を追加する運用）
  firebase_auth_domain = var.firebase_auth_domain != "" ? var.firebase_auth_domain : "${var.project_id}.firebaseapp.com"

  # Secret Manager（値はコンソール / gcloud で投入。Terraform は箱のみ）
  backend_env_secrets = {
    OPENAI_API_KEY = "aistalk-openai-api-key"
    AIVIS_API_KEY  = "aistalk-aivis-api-key"
  }

  # Cloud Run plain env（terraform apply で反映）
  backend_env_plain = {
    FIREBASE_PROJECT_ID    = var.project_id
    CHAT_MODEL             = "gpt-4.1-nano"
    MOTION_MODEL           = "gpt-4.1-nano"
    CHAT_TRIM_MAX_TOKENS   = "3000"
    CHAT_SYSTEM_PROFILE    = "single_chat"
    AIVIS_MODE             = "cloud"
    AIVIS_BASE_URL         = "https://api.aivis-project.com"
    AIVIS_SYNTHESIZE_PATH  = "/v1/tts/synthesize"
    QUOTA_DAILY_CHAT_LIMIT = tostring(var.quota_daily_chat_limit)
    QUOTA_DAILY_TTS_LIMIT  = tostring(var.quota_daily_tts_limit)
    QUOTA_DISABLED         = var.quota_disabled ? "1" : "0"
  }
}
