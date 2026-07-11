resource "google_cloudbuild_trigger" "aistalk" {
  name        = "aistalk-frontend-build"
  project     = var.project_id
  location    = var.region
  description = "フロントエンドビルド → GCS アップロード（Unity は GCS から取得）"

  repository_event_config {
    repository = var.aistalk_repository
    push {
      branch = "^main$"
    }
  }

  filename = "cloudbuild.yaml"

  service_account = "projects/${var.project_id}/serviceAccounts/${data.google_project.project.number}-compute@developer.gserviceaccount.com"

  substitutions = {
    _UNITY_BUILD_BUCKET   = google_storage_bucket.unity_build.name
    _FRONTEND_BUCKET      = google_storage_bucket.frontend.name
    _API_BASE_URL         = var.enable_app ? google_cloud_run_v2_service.backend[0].uri : ""
    _APP_NAME             = var.app_name
    _FIREBASE_API_KEY     = var.firebase_api_key
    _FIREBASE_AUTH_DOMAIN = local.firebase_auth_domain
    _FIREBASE_PROJECT_ID  = var.project_id
    _FIREBASE_APP_ID      = var.firebase_app_id
  }
}
