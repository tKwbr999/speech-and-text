data "local_file" "credentials" {
  filename = var.credentials_file
}

resource "google_secret_manager_secret" "cloud_run_sa_key" {
  secret_id = "cloud-run-sa-key"
  replication {
    auto {}
  }
}

resource "google_secret_manager_secret_version" "cloud_run_sa_key_version" {
  secret      = google_secret_manager_secret.cloud_run_sa_key.id
  secret_data = data.local_file.credentials.content
}