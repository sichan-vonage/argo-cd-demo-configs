.PHONY: app-of-apps-providers-dev
app-of-apps-providers-prod:
	argocd app create providers-prod \
    --dest-namespace argocd \
    --dest-server https://kubernetes.default.svc \
    --repo https://github.com/sichan-vonage/argo-cd-demo-configs.git \
    --path app-of-apps/providers \
	--values values-prod-config.yaml

.PHONY: app-of-apps-providers-dev
app-of-apps-providers-dev:
	argocd app create providers-dev \
    --dest-namespace argocd \
    --dest-server https://kubernetes.default.svc \
    --repo https://github.com/sichan-vonage/argo-cd-demo-configs.git \
    --path app-of-apps/providers \
	--values values-dev-config.yaml

.PHONY: docker-build-envpromoter
docker-build-envpromoter:
    cd cmd/envpromoter && docker build -t kinluek/envpromoter:1.0.5 .
    docker push kinluek/envpromoter:1.0.5

.PHONY: docker-build-automerger
docker-build-automerger:
    cd cmd/automerger && docker build -t kinluek/automerger:1.0.0 .
    docker push kinluek/automerger:1.0.0