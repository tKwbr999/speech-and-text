
provider "google" {
  project = var.project_id == "" ? terraform.workspace : var.project_id
  region  = "us-central1"
}

resource "google_cloud_run_v2_service" "default" {
  name     = "speech-and-text-service"
  location = "us-central1"
  deletion_protection = false

  template {
    containers {
      image = "gcr.io/${var.project_id == "" ? terraform.workspace : var.project_id}/speech-and-text"
      
      # ポート設定
      ports {
        container_port = 8080
      }

      # プロジェクトID環境変数
      env {
        name  = "PROJECT_ID"
        value = var.project_id
      }

      # リソース制限
      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }

      # ヘルスチェック設定
      startup_probe {
        initial_delay_seconds = 10
        timeout_seconds      = 2
        period_seconds       = 3
        failure_threshold    = 3
        tcp_socket {
          port = 8080
        }
      }
    }

    # スケーリング設定
    scaling {
      min_instance_count = 0
      max_instance_count = 100
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }
}

# パブリックアクセス設定
resource "google_cloud_run_service_iam_binding" "noauth" {
  location = google_cloud_run_v2_service.default.location
  project  = var.project_id == "" ? terraform.workspace : var.project_id
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  members  = ["allUsers"]
}

