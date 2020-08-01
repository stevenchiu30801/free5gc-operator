SHELL	:= /bin/bash
MAKEDIR	:= $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
R		?= /tmp
DEPLOY	?= $(MAKEDIR)/deploy

NAMESPACE	:= default

GO_VERSION	?= 1.13.5

SRIOV_INTF		?=
SRIOV_VF_NUM	?= 16

COLOR_WHITE			= \033[0m
COLOR_LIGHT_GREEN	= \033[1;32m
COLOR_LIGHT_RED		= \033[1;31m

define echo_green
	@echo -e "${COLOR_LIGHT_GREEN}${1}${COLOR_WHITE}"
endef

define echo_red
	@echo -e "${COLOR_LIGHT_RED}${1}${COLOR_WHITE}"
endef

.PHONY: sriovdp

sriovdp: $(R)/sriov-network-device-plugin/build/sriovdp

# https://golang.org/doc/install#install
/usr/local/go:
	curl -O -L https://dl.google.com/go/go${GO_VERSION}.linux-amd64.tar.gz
	sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
	echo -e '\nexport PATH=$$PATH:/usr/local/go/bin' >> $(HOME)/.profile
	rm go${GO_VERSION}.linux-amd64.tar.gz
	@echo -e "Please reload your shell or source $$HOME/.profile to apply the changes:\n\
		source $$HOME/.profile"

# https://github.com/intel/sriov-cni.git
/opt/cni/bin/sriov: | /usr/local/go
	-git clone https://github.com/intel/sriov-cni.git $(R)/sriov-cni
	export PATH=$$PATH:/usr/local/go/bin; cd $(R)/sriov-cni; make
	mkdir -p /opt/cni/bin
	sudo cp $(R)/sriov-cni/build/sriov $@

# https://github.com/intel/sriov-network-device-plugin
$(R)/sriov-network-device-plugin/build/sriovdp: | /usr/local/go
	-git clone https://github.com/intel/sriov-network-device-plugin.git $(R)/sriov-network-device-plugin
	export PATH=$$PATH:/usr/local/go/bin; cd $(R)/sriov-network-device-plugin; make && make image

.PHONY: sriov-server-setup sriov-init

sriov-server-setup:
	@if [[ -z "${SRIOV_INTF}" ]]; \
	then \
		echo "Invalid value: SRIOV_INTF must be provided"; \
		exit 1; \
	fi
	${SHELL} scripts/sriov_setup.sh ${SRIOV_INTF} ${SRIOV_VF_NUM}

sriov-init: | /opt/cni/bin/sriov $(R)/sriov-network-device-plugin/build/sriovdp sriov-server-setup
	sed 's/PF_NAME/${SRIOV_INTF}/g' $(DEPLOY)/sriov-configmap.yaml | sed "s/LAST_VF/$$(( ${SRIOV_VF_NUM} - 1 ))/g" | kubectl apply -f -
	kubectl apply -f $(R)/sriov-network-device-plugin/deployments/k8s-v1.16/sriovdp-daemonset.yaml

/nfsshare:
	$(call echo_green," ...... Setup NFS Server ......")
	sudo apt update
	sudo apt install -y nfs-kernel-server
	echo "/nfsshare   localhost(rw,sync,no_root_squash)" | sudo tee /etc/exports
	sudo mkdir $@
	sudo exportfs -r
	# Check if /etc/exports is properly loaded
	# showmount -e localhost

.PHONY: setup install uninstall build reset-free5gc

setup: /nfsshare sriov-init ## Setup environment
	$(call echo_green," ...... Setup Environment ......")
	kubectl apply -f https://raw.githubusercontent.com/intel/multus-cni/master/images/multus-daemonset.yml

install: setup ## Install all resources (CR/CRD's, RBAC and Operator)
	$(call echo_green," ....... Creating namespace .......")
	-kubectl create namespace ${NAMESPACE}
	$(call echo_green," ....... Applying CRDs .......")
	kubectl apply -f $(DEPLOY)/crds/bans.io_free5gcslice_crd.yaml -n ${NAMESPACE}
	$(call echo_green," ....... Applying Rules and Service Account .......")
	kubectl apply -f $(DEPLOY)/role.yaml -n ${NAMESPACE}
	kubectl apply -f $(DEPLOY)/role_binding.yaml -n ${NAMESPACE}
	kubectl apply -f $(DEPLOY)/cluster_role.yaml -n ${NAMESPACE}
	kubectl apply -f $(DEPLOY)/cluster_role_binding.yaml -n ${NAMESPACE}
	kubectl apply -f $(DEPLOY)/service_account.yaml -n ${NAMESPACE}
	$(call echo_green," ....... Applying Operator .......")
	kubectl apply -f $(DEPLOY)/operator.yaml -n ${NAMESPACE}
	# ${SHELL} scripts/wait_pods_running.sh ${NAMESPACE}
	# $(call echo_green," ....... Creating the CRs .......")
	# kubectl apply -f $(DEPLOY)/crds/bans.io_v1alpha1_free5gcslice_cr1.yaml -n ${NAMESPACE}

uninstall: ## Uninstall all that all performed in the $ make install
	$(call echo_red," ....... Uninstalling .......")
	$(call echo_red," ....... Deleting CRDs.......")
	-kubectl delete -f $(DEPLOY)/crds/bans.io_free5gcslice_crd.yaml -n ${NAMESPACE}
	$(call echo_red," ....... Deleting Rules and Service Account .......")
	-kubectl delete -f $(DEPLOY)/role.yaml -n ${NAMESPACE}
	-kubectl delete -f $(DEPLOY)/role_binding.yaml -n ${NAMESPACE}
	-kubectl delete -f $(DEPLOY)/cluster_role.yaml -n ${NAMESPACE}
	-kubectl delete -f $(DEPLOY)/cluster_role_binding.yaml -n ${NAMESPACE}
	-kubectl delete -f $(DEPLOY)/service_account.yaml -n ${NAMESPACE}
	$(call echo_red," ....... Deleting Operator .......")
	-kubectl delete -f $(DEPLOY)/operator.yaml -n ${NAMESPACE}
	$(call echo_red," ....... Deleting namespace ${NAMESPACE}.......")
	-kubectl delete namespace ${NAMESPACE}

build: ## Build Operator
	$(call echo_green," ...... Building Operator ......")
	operator-sdk build steven30801/free5gc-operator:latest
	$(call echo_green," ...... Pushing image ......")
	docker push steven30801/free5gc-operator:latest

reset-free5gc: ## Uninstall all free5GC functions along with CR except Mongo DB
	-helm uninstall free5gc
	-${SHELL} scripts/remove_slices.sh
	-${SHELL} scripts/clear_mongo.sh
	-${SHELL} scripts/remove_crs.sh
	${SHELL} scripts/wait_pods_terminating.sh ${NAMESPACE}
