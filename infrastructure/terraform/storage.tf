# Unity ビルド成果物用（手動アップロード: gs://BUCKET/latest/unity/）
resource "google_storage_bucket" "unity_build" {
  name     = "${var.project_id}-aistalk-unity-build"
  location = var.region
  project  = var.project_id

  uniform_bucket_level_access = true
}

# 公開アセット（VRM 等）。ブラウザから直接 fetch
resource "google_storage_bucket" "storage" {
  name     = "${var.project_id}-aistalk-storage"
  location = var.region
  project  = var.project_id

  uniform_bucket_level_access = true

  cors {
    origin          = ["https://${var.domain}", "http://localhost:3000"]
    method          = ["GET", "HEAD"]
    response_header = ["Content-Type", "Content-Length", "Content-Encoding"]
    max_age_seconds = 3600
  }
}

resource "google_storage_bucket_iam_member" "storage_public_viewer" {
  bucket = google_storage_bucket.storage.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Next.js 静的配信用
resource "google_storage_bucket" "frontend" {
  name     = "${var.project_id}-aistalk-frontend"
  location = var.region
  project  = var.project_id

  uniform_bucket_level_access = true
}

data "google_project" "project" {
  project_id = var.project_id
}

resource "google_storage_bucket_iam_member" "frontend_public_viewer" {
  bucket = google_storage_bucket.frontend.name
  role   = "roles/storage.objectViewer"
  member = "allUsers"
}

# Cloud Build が unity-build を読めるように
resource "google_storage_bucket_iam_member" "unity_build_cb_viewer" {
  bucket = google_storage_bucket.unity_build.name
  role   = "roles/storage.objectViewer"
  member = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}

# Cloud Build が frontend バケットに書き込めるように
resource "google_storage_bucket_iam_member" "frontend_cb_admin" {
  bucket = google_storage_bucket.frontend.name
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${data.google_project.project.number}-compute@developer.gserviceaccount.com"
}
