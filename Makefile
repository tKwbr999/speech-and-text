export GOOGLE_APPLICATION_CREDENTIALS ?= $(shell cat .env | grep GOOGLE_APPLICATION_CREDENTIALS | cut -d'=' -f2)
export PROJECT_ID ?= $(shell cat .env | grep PROJECT_ID | cut -d'=' -f2)
export TF_VAR_project_id := data-sunlight-445604-q4

.PHONY: terraform-init
terraform-init:
	terraform -chdir=terraform init


.PHONY: build-push-image
build-push-image:
	gcloud builds submit --tag gcr.io/$(PROJECT_ID)/speech-and-text .

.PHONY: terraform-apply
terraform-apply:
	terraform -chdir=terraform apply -auto-approve -replace="google_cloud_run_v2_service.default"

.PHONY: terraform-destroy
terraform-destroy:
	terraform -chdir=terraform destroy

.PHONY: image-list
image-list:
	gcloud container images list --repository=gcr.io/$(PROJECT_ID)

.PHONY: gcloud-auth-configure-docker
gcloud-auth-configure-docker:
	gcloud auth configure-docker

.PHONY: pull-image
pull-image:
	docker pull gcr.io/$(PROJECT_ID)/speech-and-text

.PHONY: compose-build-up
compose-build-up:
	docker-compose up -d --build
