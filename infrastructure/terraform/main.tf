# terraform ブロック: バージョン制約とプロバイダ指定。
#   required_providers: 使用するプロバイダとバージョン（~> 5.0 = 5.x）
terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region

  # Identity Toolkit 等は ADC 利用時に quota project が必須
  # gcloud の set-quota-project だけでは Terraform に効かないことがあるため明示する
  user_project_override = true
  billing_project       = var.project_id
}
