TERRAFORM_DIR := terraform

.PHONY: terraform-init terraform-plan terraform-apply terraform-destroy backend-build check-tfvars

terraform-apply: check-tfvars
	cd $(TERRAFORM_DIR) && terraform apply -var-file=terraform.tfvars

terraform-plan: check-tfvars
	cd $(TERRAFORM_DIR) && terraform plan -var-file=terraform.tfvars

terraform-init:
	cd $(TERRAFORM_DIR) && terraform init

terraform-destroy: check-tfvars
	cd $(TERRAFORM_DIR) && terraform destroy -var-file=terraform.tfvars

check-tfvars:
	@if [ ! -f $(TERRAFORM_DIR)/terraform.tfvars ]; then \
		echo "Error: $(TERRAFORM_DIR)/terraform.tfvars not found."; \
		echo "Copy terraform.tfvars.example to terraform.tfvars and fill in your values:"; \
		echo "  cp $(TERRAFORM_DIR)/terraform.tfvars.example $(TERRAFORM_DIR)/terraform.tfvars"; \
		exit 1; \
	fi

# enable_app=true かつ terraform apply 後: バックエンドイメージをビルドして Cloud Run に反映
backend-build: check-tfvars
	@PROJECT=$$(grep '^project_id' $(TERRAFORM_DIR)/terraform.tfvars | sed 's/.*= *"\(.*\)".*/\1/'); \
	REGION=$$(grep '^region' $(TERRAFORM_DIR)/terraform.tfvars | sed 's/.*= *"\(.*\)".*/\1/'); \
	cd .. && gcloud builds submit --config cloudbuild-backend.yaml \
	  --substitutions=_REGION=$$REGION,_AR_REPO=aistalk-backend,_SERVICE_NAME=aistalk-backend,_SERVICE_ACCOUNT=aistalk-backend@$$PROJECT.iam.gserviceaccount.com
