app-of-apps-providers-prod:
	argocd app create apps \
    --dest-namespace argocd \
    --dest-server https://kubernetes.default.svc \
    --repo https://github.com/sichan-vonage/argo-cd-demo-configs.git \
    --path app-of-apps/providers \
	--values values-prod-config.yaml

app-of-apps-providers-dev:
	argocd app create apps \
    --dest-namespace argocd \
    --dest-server https://kubernetes.default.svc \
    --repo https://github.com/sichan-vonage/argo-cd-demo-configs.git \
    --path app-of-apps/providers \
	--values values-dev-config.yaml