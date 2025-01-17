provider "google" {
  project = var.project_id == "" ? terraform.workspace : var.project_id
  region  = "us-central1"
}

provider "google-beta" {
  project = var.project_id == "" ? terraform.workspace : var.project_id
  region  = "us-central1"
}

resource "google_project_service_identity" "run_sa" {
  provider = google-beta
  project  = var.project_id == "" ? terraform.workspace : var.project_id
  service  = "run.googleapis.com"
}

resource "google_project_iam_member" "cloud_run_sa_secret_manager" {
  project = var.project_id == "" ? terraform.workspace : var.project_id
  role    = "roles/secretmanager.secretAccessor"
  member  = "serviceAccount:${google_project_service_identity.run_sa.email}"
}

resource "google_cloud_run_v2_service" "default" {
  name     = "speech-and-text-service"
  location = "us-central1"
  ingress = "INGRESS_TRAFFIC_ALL"
  
  template {
    containers {
      image = "gcr.io/${var.project_id == "" ? terraform.workspace : var.project_id}/speech-and-text"
      
      ports {
        container_port = 8080
      }

      env {
        name  = "PROJECT_ID"
        value = var.project_id == "" ? terraform.workspace : var.project_id
      }

      env {
        name = "GOOGLE_APPLICATION_CREDENTIALS"
        value_source {
          secret_key {
            secret = google_secret_manager_secret.cloud_run_sa_key.name
            version = "latest"
          }
        }
      }

      resources {
        limits = {
          cpu    = "1000m"
          memory = "512Mi"
        }
      }

      startup_probe {
        initial_delay_seconds = 10
        timeout_seconds       = 2
        period_seconds        = 3
        failure_threshold     = 3
        tcp_socket {
          port = 8080
        }
      }
    }

    scaling {
      min_instance_count = 0
      max_instance_count = 100
    }
  }

  traffic {
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
    percent = 100
  }

  depends_on = [google_project_iam_member.cloud_run_sa_secret_manager]
}

resource "google_cloud_run_service_iam_member" "noauth" {
  location = google_cloud_run_v2_service.default.location
  project  = var.project_id == "" ? terraform.workspace : var.project_id
  service  = google_cloud_run_v2_service.default.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
