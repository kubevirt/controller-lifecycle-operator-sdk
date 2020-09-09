# Operator reference implementation

 `sample-operator` controls deployment of a simple, not customized nginx HTTP server.

 ## Deployment
 `sample-operator` is deployed in `kubevirt` namespace and deployment manifests will create it, if it's not present.

 To deploy the operator execute following command:
 ```shell script
kubectl apply -f manifests/v0.0.3/operator.yaml
```

Deployment of the HTTP server is controlled by the presence of a `sampleconfig` CR; to deploy it, execute:
```shell script
kubectl apply -f manifests/v0.0.3/sampleconfig_types.go
```

The HTTP service listens on port `8081`, which in turn is exposed as a node port `30080`.

In case of deployments like CRC or kubevirtci, port forwarding might be required:
```shell script
kubectl -n kubevirt port-forward service/http-server 8081:8081
```

To upgrade the HTTP server deployment, change configuration of the `sample-operator` container in the `spec.template.spec.containers` array of `sample-operator` deployment.

## Building the sample operator
For convenience Makefile has been provided that helps with testing, building and publishing of the operator:
- `make docker-build` will build both operator and HTTP server images (by default: `quay.io/$USER/sample-operator:latest` and `quay.io/$USER/sample-http-server:latest`)
- `make docker-push` will push both images to the remote repository;