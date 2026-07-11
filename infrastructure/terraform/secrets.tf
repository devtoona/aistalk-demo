# Secret Manager の箱のみ（Terraform は値を読まない → state に載らない）

resource "google_secret_manager_secret" "backend_env" {
  for_each = local.backend_env_secrets

  secret_id = each.value
  project   = var.project_id

  replication {
    auto {}
  }

  depends_on = [google_project_service.backend_apis]
}

resource "google_secret_manager_secret_iam_member" "backend_env_accessor" {
  for_each = google_secret_manager_secret.backend_env

  secret_id = each.value.id
  role      = "roles/secretmanager.secretAccessor"
  member    = "serviceAccount:${google_service_account.cloud_run_backend.email}"
}
