resource "google_project_service" "backend_apis" {
  for_each = toset([
    "run.googleapis.com",
    "artifactregistry.googleapis.com",
    "secretmanager.googleapis.com",
    "firestore.googleapis.com",
    "identitytoolkit.googleapis.com",
    "firebase.googleapis.com",
    "firebaserules.googleapis.com",
  ])

  project            = var.project_id
  service            = each.value
  disable_on_destroy = false
}

resource "google_service_account" "cloud_run_backend" {
  account_id   = "aistalk-backend"
  display_name = "AISTalk Cloud Run backend"
  project      = var.project_id
}

resource "google_artifact_registry_repository" "backend" {
  count = var.enable_app ? 1 : 0

  location      = var.region
  repository_id = "aistalk-backend"
  description   = "AISTalk backend API images"
  format        = "DOCKER"

  depends_on = [google_project_service.backend_apis]
}

resource "google_cloud_run_v2_service" "backend" {
  count = var.enable_app ? 1 : 0

  name     = "aistalk-backend"
  location = var.region
  project  = var.project_id

  ingress = "INGRESS_TRAFFIC_ALL"

  template {
    service_account = google_service_account.cloud_run_backend.email
    timeout         = "${var.backend_request_timeout_seconds}s"

    containers {
      image = var.backend_bootstrap_image

      ports {
        container_port = 8080
      }

      dynamic "env" {
        for_each = local.backend_env_plain
        content {
          name  = env.key
          value = env.value
        }
      }

      dynamic "env" {
        for_each = local.backend_env_secrets
        content {
          name = env.key
          value_source {
            secret_key_ref {
              secret  = google_secret_manager_secret.backend_env[env.key].secret_id
              version = "latest"
            }
          }
        }
      }

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 1
    }
  }

  lifecycle {
    ignore_changes = [
      template[0].containers[0].image,
      client,
      client_version,
    ]
  }

  depends_on = [
    google_project_service.backend_apis,
    google_artifact_registry_repository.backend,
    google_secret_manager_secret_iam_member.backend_env_accessor,
  ]
}

# Firebase Admin SDK（ID token 検証）
resource "google_project_iam_member" "backend_firebase_admin" {
  count = var.enable_app ? 1 : 0

  project = var.project_id
  role    = "roles/firebase.admin"
  member  = "serviceAccount:${google_service_account.cloud_run_backend.email}"
}

# Firestore（quota 読み書き）。クライアント直アクセスは rules で全 deny
resource "google_project_iam_member" "backend_datastore_user" {
  count = var.enable_app ? 1 : 0

  project = var.project_id
  role    = "roles/datastore.user"
  member  = "serviceAccount:${google_service_account.cloud_run_backend.email}"
}

resource "google_cloud_run_v2_service_iam_member" "backend_public" {
  count = var.enable_app ? 1 : 0

  project  = var.project_id
  location = google_cloud_run_v2_service.backend[0].location
  name     = google_cloud_run_v2_service.backend[0].name
  role     = "roles/run.invoker"
  member   = "allUsers"
}

resource "google_project_iam_member" "cloud_build_run_admin" {
  count = var.enable_app ? 1 : 0

  project = var.project_id
  role    = "roles/run.admin"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_project_iam_member" "cloud_build_ar_writer" {
  count = var.enable_app ? 1 : 0

  project = var.project_id
  role    = "roles/artifactregistry.writer"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_project_iam_member" "cloud_build_sa_user" {
  count = var.enable_app ? 1 : 0

  project = var.project_id
  role    = "roles/iam.serviceAccountUser"
  member  = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

resource "google_service_account_iam_member" "cloud_build_uses_backend_sa" {
  count = var.enable_app ? 1 : 0

  service_account_id = google_service_account.cloud_run_backend.name
  role               = "roles/iam.serviceAccountUser"
  member             = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
