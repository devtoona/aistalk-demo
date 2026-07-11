resource "google_cloudbuild_trigger" "aistalk_backend" {
  count = var.enable_app ? 1 : 0

  name        = "aistalk-backend-build"
  project     = var.project_id
  location    = var.region
  description = "AISTalk backend: Docker build → Artifact Registry → Cloud Run"

  repository_event_config {
    repository = var.aistalk_repository
    push {
      branch = var.backend_build_branch
    }
  }

  filename = "cloudbuild-backend.yaml"

  service_account = "projects/${var.project_id}/serviceAccounts/${data.google_project.project.number}-compute@developer.gserviceaccount.com"

  substitutions = {
    _REGION          = var.region
    _AR_REPO         = google_artifact_registry_repository.backend[0].repository_id
    _SERVICE_NAME    = google_cloud_run_v2_service.backend[0].name
    _SERVICE_ACCOUNT = google_service_account.cloud_run_backend.email
  }
}
