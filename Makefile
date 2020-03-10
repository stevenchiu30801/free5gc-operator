SHELL		:= /bin/bash
NAMESPACE	:= default

COLOR_WHITE			= \033[0m
COLOR_LIGHT_GREEN	= \033[1;32m
COLOR_LIGHT_RED		= \033[1;31m

define echo_green
	@echo -e "${COLOR_LIGHT_GREEN}${1}${COLOR_WHITE}"
endef

define echo_red
	@echo -e "${COLOR_LIGHT_RED}${1}${COLOR_WHITE}"
endef

.PHONY: install uninstall build

install: ## Install all resources (CR/CRD's, RBAC and Operator)
	$(call echo_green,"....... Creating namespace .......")
	-kubectl create namespace ${NAMESPACE}
	$(call echo_green," ....... Applying CRDs .......")
	kubectl apply -f deploy/crds/bans.io_free5gcslice_crd.yaml -n ${NAMESPACE}
	$(call echo_green," ....... Applying Rules and Service Account .......")
	kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
	kubectl apply -f deploy/role_binding.yaml -n ${NAMESPACE}
	kubectl apply -f deploy/cluster_role.yaml -n ${NAMESPACE}
	kubectl apply -f deploy/cluster_role_binding.yaml -n ${NAMESPACE}
	kubectl apply -f deploy/service_account.yaml -n ${NAMESPACE}
	$(call echo_green," ....... Applying Operator .......")
	kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}
	$(call echo_green," ....... Creating the CRs .......")
	kubectl apply -f deploy/crds/bans.io_v1alpha1_free5gcslice_cr.yaml -n ${NAMESPACE}

uninstall: ## Uninstall all that all performed in the $ make install
	$(call echo_red," ....... Uninstalling .......")
	$(call echo_red," ....... Deleting CRDs.......")
	-kubectl delete -f deploy/crds/bans.io_free5gcslice_crd.yaml -n ${NAMESPACE}
	$(call echo_red," ....... Deleting Rules and Service Account .......")
	-kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
	-kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
	-kubectl delete -f deploy/cluster_role.yaml -n ${NAMESPACE}
	-kubectl delete -f deploy/cluster_role_binding.yaml -n ${NAMESPACE}
	-kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}
	$(call echo_red," ....... Deleting Operator .......")
	-kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}
	$(call echo_red," ....... Deleting namespace ${NAMESPACE}.......")
	-kubectl delete namespace ${NAMESPACE}

build: ## Build Operator
	$(call echo_green,"...... Building Operator ......")
	operator-sdk build steven30801/free5gc-operator

reset-free5gc: ## Uninstall all free5GC functions along with CR except Mongo DB
	-helm uninstall free5gc
	-${SHELL} scripts/remove_slices.sh
	-${SHELL} scripts/clear_mongo.sh
	-${SHELL} scripts/remove_crs.sh
