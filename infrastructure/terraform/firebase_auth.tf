# Firebase Auth（Identity Platform）
# 前提: GCP プロジェクトに Firebase を「追加」済み（project_id と Firebase プロジェクト ID が同一）
# Web アプリの apiKey / appId は Console で作成し tfvars に記入

resource "google_identity_platform_config" "auth" {
  project = var.project_id

  sign_in {
    anonymous {
      enabled = true
    }
  }

  # Terraform 管理時はリスト全体を置き換えるため、デフォルト系も明示する
  authorized_domains = distinct(concat([
    "localhost",
    "${var.project_id}.firebaseapp.com",
    "${var.project_id}.web.app",
    var.domain,
  ], var.firebase_extra_authorized_domains))

  depends_on = [google_project_service.backend_apis]
}
