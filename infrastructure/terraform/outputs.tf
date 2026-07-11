output "unity_build_bucket" {
  description = "Unity ビルド成果物用 GCS バケット名（手動: gsutil cp -r ... gs://BUCKET/latest/unity/）"
  value       = google_storage_bucket.unity_build.name
}

output "frontend_bucket" {
  description = "フロントエンド配信用 GCS バケット名"
  value       = google_storage_bucket.frontend.name
}

output "storage_bucket" {
  description = "公開アセット用 GCS バケット名（VRM: gs://BUCKET/models/ 等）"
  value       = google_storage_bucket.storage.name
}

output "frontend_url" {
  description = "フロントエンドのアクセス URL（CDN 経由）"
  value       = "https://${var.domain}"
}

output "lb_ip_address" {
  description = "Load Balancer の IP。DNS の A レコードでこの IP を指す"
  value       = google_compute_global_address.frontend.address
}

output "backend_secret_ids" {
  description = "Cloud Run 注入用 Secret Manager secret id 一覧（env 名 → secret id）"
  value       = local.backend_env_secrets
}

output "backend_env_plain" {
  description = "Cloud Run plain env（terraform apply で注入）"
  value       = local.backend_env_plain
}

output "backend_url" {
  description = "Cloud Run backend URL (enable_app=true のとき)"
  value       = var.enable_app ? google_cloud_run_v2_service.backend[0].uri : null
}

output "backend_artifact_registry" {
  description = "Backend Docker repository (enable_app=true のとき)"
  value       = var.enable_app ? "${var.region}-docker.pkg.dev/${var.project_id}/${google_artifact_registry_repository.backend[0].repository_id}" : null
}

output "backend_service_account" {
  description = "Cloud Run backend service account"
  value       = google_service_account.cloud_run_backend.email
}

output "firestore_database" {
  description = "Firestore database name"
  value       = google_firestore_database.default.name
}
