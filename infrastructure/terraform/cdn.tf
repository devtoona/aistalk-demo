# Cloud CDN + Load Balancer（frontend バケットをオリジンに）

resource "google_compute_backend_bucket" "frontend" {
  name        = "aistalk-frontend-backend"
  project     = var.project_id
  bucket_name = google_storage_bucket.frontend.name
  enable_cdn  = true
}

resource "google_compute_url_map" "frontend" {
  name            = "aistalk-frontend-url-map"
  project         = var.project_id
  default_service = google_compute_backend_bucket.frontend.id

  host_rule {
    hosts        = ["*"]
    path_matcher = "frontend"
  }

  path_matcher {
    name            = "frontend"
    default_service = google_compute_backend_bucket.frontend.id

    route_rules {
      priority = 1
      match_rules {
        full_path_match = "/"
      }
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/index.html"
        }
      }
      service = google_compute_backend_bucket.frontend.id
    }

    # WebGL 成果物（public/unity/*）
    route_rules {
      priority = 2
      match_rules {
        prefix_match = "/unity/"
      }
      route_action {
        url_rewrite {
          path_prefix_rewrite = "unity/"
        }
      }
      service = google_compute_backend_bucket.frontend.id
    }
  }
}

resource "google_compute_target_http_proxy" "frontend" {
  name    = "aistalk-frontend-http-proxy"
  project = var.project_id
  url_map = google_compute_url_map.frontend.id
}

resource "google_compute_global_address" "frontend" {
  name    = "aistalk-frontend-ip"
  project = var.project_id
}

resource "google_compute_global_forwarding_rule" "frontend" {
  name                  = "aistalk-frontend-forwarding-rule"
  project               = var.project_id
  target                = google_compute_target_http_proxy.frontend.id
  port_range            = "80"
  load_balancing_scheme = "EXTERNAL"
  ip_address            = google_compute_global_address.frontend.address
}

resource "google_compute_managed_ssl_certificate" "frontend" {
  name = "aistalk-frontend-ssl-cert"

  managed {
    domains = [var.domain]
  }
}

resource "google_compute_target_https_proxy" "frontend" {
  name             = "aistalk-frontend-https-proxy"
  project          = var.project_id
  url_map          = google_compute_url_map.frontend.id
  ssl_certificates = [google_compute_managed_ssl_certificate.frontend.id]
}

resource "google_compute_global_forwarding_rule" "frontend_https" {
  name                  = "aistalk-frontend-https-forwarding-rule"
  project               = var.project_id
  target                = google_compute_target_https_proxy.frontend.id
  port_range            = "443"
  load_balancing_scheme = "EXTERNAL"
  ip_address            = google_compute_global_address.frontend.address
}
