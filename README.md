# openshift-prometheus-grafana-ambassador

## Why?

Red Hat OpenShift Container Platform provides a playbook to install Prometheus,
a time series database to collect metrics on infrastructure and applications.

Since Prometheus doesn't provide authentication, the OpenShift template enables
an OAuth proxy: this sounds cool, except for the side effect that the Grafana
Prometheus data source doesn't handle this kind of authorization.

A solution could be rebuilding the source image but this is not agile.
Another one is using the `contrib` playbook provided by the OpenShift Origin
repository to install Grafana, but the version provided is outdated

Reading the book `Designing Distributed Systems` I discovered the *Ambassador*
pattern: this is an implementation to solve this problem in order to achieve a
reusable and modular component, also to get rid-off of patching over versioning.

## Variables

The ambassador uses two mandatory variables:
* `PROMETHEUS_SERVICE`: the Prometheus service (according to a regular installation
  using the official Ansible playbook: `https://prometheus.openshift-metrics.svc`)
* `TOKEN`: the authorized Service Account token

If you have troubles with your unsigned CA:
* `SKIP_INSECURE_VERIFY`: boolean, skip certificate checking.

## How to use

Deploy it on your DeploymentConfig as a sidecar container, just a little snippet:

```
apiVersion: v1
kind: DeploymentConfig
metadata:
  labels:
    app: grafana
  name: grafana
spec:
  replicas: 1
  selector:
    app: grafana
    deploymentconfig: grafana
  template:
    metadata:
      labels:
        app: grafana
        deploymentconfig: grafana
    spec:
      containers:
      - env:
        - name: PROMETHEUS_SERVICE
          value: <FIXME:PROMETHEUS_SERVICE>
        - name: TOKEN
          value: <FIXME:TOKEN>
        - name: SKIP_INSECURE_VERIFY
          value: <FIXME:BOOL>
        image: prometherion/grafana-ambassador
        imagePullPolicy: Always
        name: ambassador
      - image: grafana/grafana@<FIXME:GRAFANA_VERSION>
        name: grafana
        ports:
        - containerPort: 3000
          protocol: TCP
        volumeMounts:
        - mountPath: /var/lib/grafana
          name: grafana-2
        - mountPath: /var/log/grafana
          name: grafana-3
      serviceAccount: <FIXME:GRAFANA_SA>
      serviceAccountName: <FIXME:GRAFANA_SA>
      volumes:
      - emptyDir: {}
        name: grafana-2
      - emptyDir: {}
        name: grafana-3
```

## Backlog

Still a lot to do: the image is build via Docker multi-stage, but we want to migrate to Source2Image.

No Go tests (my shame).
