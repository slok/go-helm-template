apiVersion: batch/v1
kind: Job
metadata:
  name: eg-{{.Values.cfg.tenant}}-gateway-helm-certgen
  namespace: {{ .Values.cfg.controllerNamespace }}
  labels:
    helm.sh/chart: gateway-helm-latest
    app.kubernetes.io/name: gateway-helm
    app.kubernetes.io/instance: eg-{{.Values.cfg.tenant}}
    app.kubernetes.io/version: "latest"
    app.kubernetes.io/managed-by: Helm
  annotations:
    "helm.sh/hook": pre-install, pre-upgrade
spec:
  backoffLimit: 1
  completions: 1
  parallelism: 1
  template:
    metadata:
      labels:
        app: certgen
    spec:
      containers:
      - command:
        - envoy-gateway
        - certgen
        env:
        - name: ENVOY_GATEWAY_NAMESPACE
          valueFrom:
            fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
        - name: KUBERNETES_CLUSTER_DOMAIN
          value: cluster.local
        image: {{ .Values.eg.certgen.image }}
        imagePullPolicy: IfNotPresent
        name: envoy-gateway-certgen
      imagePullSecrets: []
      restartPolicy: Never
      securityContext:
        runAsGroup: 65534
        runAsNonRoot: true
        runAsUser: 65534
      serviceAccountName: eg-{{.Values.cfg.tenant}}-gateway-helm-certgen
  ttlSecondsAfterFinished: 30