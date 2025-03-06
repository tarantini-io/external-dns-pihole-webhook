# ExternalDNS Webhook Provider for PiHole v6

<div align="center">

[![GitHub Release](https://img.shields.io/github/v/release/tarantini-io/external-dns-pihole-webhook?style=for-the-badge)](https://github.com/tarantini-io/external-dns-pihole-webhook/releases)
&nbsp;&nbsp;

</div>

[ExternalDNS](https://github.com/kubernetes-sigs/external-dns) is a Kubernetes add-on for automatically managing DNS
records for Kubernetes ingresses and services by using different DNS providers. This webhook provider allows you to
automate DNS records from your Kubernetes clusters into your self-hosted PiHole instance.

## 🎯 Requirements

- ExternalDNS >= v0.14.0
- PiHole >= 6.x

## ⛵ Deployment

### Installing the provider

1. Add the ExternalDNS Helm repository to your cluster.

    ```sh
    helm repo add external-dns https://kubernetes-sigs.github.io/external-dns/
    ```

2. Deploy your `pihole-password` secret that holds your password.

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: pihole-password
    stringData:
      password: <PiHole password>
    ```

3. Create the helm values file, for example `external-dns-pihole-values.yaml`:

    ```yaml
    fullnameOverride: external-dns-pihole
    logLevel: &logLevel debug
    provider:
      name: webhook
      webhook:
        image:
          repository: ghcr.io/tarantini-io/external-dns-pihole-webhook
          tag: main # or specific tag
        env:
          - name: PIHOLE_PASSWORD
            valueFrom:
              secretKeyRef:
                name: pihole-password
                key: password
          - name: LOG_LEVEL
            value: *logLevel
        livenessProbe:
          httpGet:
            path: /healthz
            port: http-webhook
          initialDelaySeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /readyz
            port: http-webhook
          initialDelaySeconds: 10
          timeoutSeconds: 5
    policy: sync
    sources: ["ingress", "gateway-httproute", "service"] # or whatever you need
    txtOwnerId: default
    txtPrefix: k8s.
    domainFilters: ["example.com"] # your domain
    ```

4. Install the Helm chart

    ```sh
    helm install external-dns-pihole external-dns/external-dns -f external-dns-pihole-values.yaml
    ```

## Configuration

### PiHole Controller Configuration

| Environment Variable  | Description                                                  | Default Value       |
|-----------------------|--------------------------------------------------------------|---------------------|
| `PIHOLE_PASSWORD`     | The PiHole password                                          | N/A                 |
| `PIHOLE_SERVER`       | The full path of your PiHole instance.                       | `http://pi.hole:80` |
| `PIHOLE_TLS_INSECURE` | Whether to allow insecure TLS verification (true or false).  | `false`             |
| `PIHOLE_DRY_RUN`      | Whether to not applied but just log changes                  | `false`             |
| `LOG_LEVEL`           | Change the verbosity of logs (used when making a bug report) | `info`              |

### Server Configuration

| Environment Variable             | Description                                                      | Default Value |
|----------------------------------|------------------------------------------------------------------|---------------|
| `SERVER_HOST`                    | The host address where the server listens.                       | `localhost`   |
| `SERVER_PORT`                    | The port where the server listens.                               | `8888`        |
| `SERVER_READ_TIMEOUT`            | Duration the server waits before timing out on read operations.  | N/A           |
| `SERVER_WRITE_TIMEOUT`           | Duration the server waits before timing out on write operations. | N/A           |
| `DOMAIN_FILTER`                  | List of domains to include in the filter.                        | Empty         |
| `EXCLUDE_DOMAIN_FILTER`          | List of domains to exclude from filtering.                       | Empty         |
| `REGEXP_DOMAIN_FILTER`           | Regular expression for filtering domains.                        | Empty         |
| `REGEXP_DOMAIN_FILTER_EXCLUSION` | Regular expression for excluding domains from the filter.        | Empty         |

---

## 🤝 Gratitude and Thanks

Thanks to @kashalls for their work on the [Unifi Webhook](https://github.com/kashalls/external-dns-unifi-webhook/tree/main) which I used as a base
