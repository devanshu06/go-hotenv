# Example: hotenv with Doppler + Kubernetes

This example shows how to ship the `hotenv` sample service from [`main.go`](./main.go) inside a container, then deploy it to Kubernetes with Doppler-powered configuration managed by External Secrets Operator (ESO).

---

## Prerequisites

- Go 1.25+ (for local builds) and Docker.
- Access to a container registry (Docker Hub, GHCR, ECR, etc.).
- A Kubernetes cluster with kubectl configured.
- External Secrets Operator installed in the cluster.
- A Doppler project + config with the secrets you want to inject.

---

## 1. Build and push the image

From the repository root, build the multi-stage Docker image:

```bash
docker build -t <registry>/<repo>/hotenv-example:<tag> -f example/Dockerfile example
```

Push it to your registry:

```bash
docker push <registry>/<repo>/hotenv-example:<tag>
```

Take note of the image reference; you will use it in the Kubernetes manifest.

---

## 2. Prepare Doppler integration

Open [`example/mainfest.yaml`](./mainfest.yaml) and update the placeholders:

- `Namespace` (optional) — change `demo` if you prefer a different namespace.
- `Secret` (`doppler-service-token`) — replace `<DOPPLER_TOKEN>` with a Doppler Service Token that has access to the target project/config.
- `SecretStore` — set `<DOPPLER_PROJECT>` and `<DOPPLER_CONFIG>` to match your Doppler environment.
- `ExternalSecret` — keep defaults unless you want a subset of keys.

The manifest renders all Doppler keys into a single `.env` file stored in an ESO-managed Kubernetes secret called `app-secrets`.

---

## 3. Point the Deployment at your image

In the same manifest, update the Deployment container image:

```yaml
          image: <registry>/<repo>/hotenv-example:<tag>
```

The deployment mounts the generated `.env` file at `/app/secrets/.env` and sets `SECRETS_FILE` so `hotenv` watches the correct path.

---

## 4. Apply the manifest

Apply all resources in one go (namespace, secrets, deployment, service):

```bash
kubectl apply -f example/mainfest.yaml
```

Wait for the pod to become Ready:

```bash
kubectl -n <namespace> get pods
```

(Optional) follow logs:

```bash
kubectl -n <namespace> logs deploy/hello-app -f
```

---

## 5. Test the service

Forward the service locally:

```bash
kubectl -n <namespace> port-forward svc/hello-app 8080:80
```

Hit the endpoints:

```bash
curl http://localhost:8080/health   # should return OK
curl http://localhost:8080/hi       # should reflect GREETING_TEXT from Doppler
```

Whenever Doppler updates the secret, `hotenv` reloads the `.env` file automatically—no redeploy needed.

---

## Clean up

Remove the sample resources when you are done:

```bash
kubectl delete -f example/mainfest.yaml
```

Delete the pushed image from your registry if desired.
