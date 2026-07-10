# Root Makefile - delegates to subdirectories

.PHONY: terraform-init terraform-plan terraform-apply terraform-destroy

terraform-init:
	$(MAKE) -C infrastructure terraform-init

terraform-plan:
	$(MAKE) -C infrastructure terraform-plan

terraform-apply:
	$(MAKE) -C infrastructure terraform-apply

terraform-destroy:
	$(MAKE) -C infrastructure terraform-destroy
