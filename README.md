# External Secrets Transformer

Transform kubernetes secrets to external secrets automatically.

## Idea

I had an issue where I wanted to use third party helm charts (Argo CD, Prometheus, etc). But I had all my secrets in a keyvault. I didn't want to update or maintain all those charts by myself so therefore I made this tool to help me out with that.

When I started to use Kustomize, I wanted this tool to be able to run as a Kustomize transformer through an Exec KRM Function.

## How it works

The program reads yaml documents from stdin and transforms the secrets which includes go template variables in their `.data` or `.stringData` fields where it converts the ones containing `{{ .someVariable }}` to external secrets.

## Usage

To use it, you must set these environment variables:
| Name               | Required | Default Value |
| ------------------ | -------- | ------------- |
| `STORE_NAME`       | X        |               |
| `STORE_KIND`       | X        |               |
| `REFRESH_INTERVAL` |          | `1h`          |

## Example

Example with `STORE_NAME=gcp-store` and ` STORE_KIND=ClusterSecretStore`

**Before**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: dotfile-secret
data:
  apiToken: e3sgYXBpLXRva2VuIH19
stringData:
  config-file.json: |
    {
      "username": "admin123",
      "password": "{{ password }}",
    }
```

**After**

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: dotfile-secret
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: gcp-store
    kind: ClusterSecretStore
  target:
    template:
      data:
        apiToken: "{{ .api-token }}"
        config-file.json: |
          {
            "username": "admin123",
            "password": "{{ .password }}",
          }
  data:
    - secretKey: password
      remoteRef:
        key: password
    - secretKey: api-token
      remoteRef:
        key: api-token

---
```

## Build
Just run `make all` and use the binary for your OS.

## Binary
```bash
cat k8s-manifest.yaml | ./external-secrets-transformer-linux-amd64
```

## Helm

```bash
helm template <CHART> --post-renderer ./external-secrets-transformer-linux-amd64
```

## Kustomize
1. Create a directory in your kustomize project called `plugins`. In that directory, create a file called `external-secrets-transformer.sh`.
2. Add the following content to the file, remember to specify the correct path to the EXT binary.

    ```bash
    #!/bin/bash
    resourceList=$(cat) # read the `kind: ResourceList` from stdin
    storeName=$(echo "$resourceList" | yq e '.functionConfig.spec.storeName' - )
    storeKind=$(echo "$resourceList" | yq e '.functionConfig.spec.storeKind' - )

    export STORE_NAME="$storeName"
    export STORE_KIND="$storeKind"

    echo "$resourceList" | (/PATH/TO/external-secrets-transformer-macos-amd64) #> bananer
    ```
3. Create a file in your project directory called `ESTransformer.yaml` and set the correct `spec`.
      ```yaml
      # $MYAPP/cmGenerator.yaml
      apiVersion: sjoedin.se
      kind: ExternalSecretsTransformer
      metadata:
        name: external-secrets-transformer
        annotations:
          config.kubernetes.io/function: |
            exec: 
              path: ./plugins/external-secrets-transformer.sh
      spec:
        storeKind: ClusterSecretStore
        storeName: my-keyvault
      ```
4. Add this to your `kustomization.yaml`.

    ```yaml
    transformers:
      - ./ESTransformer.yaml
    ```
5. Run kustomize `kustomize build .  --enable-exec --enable-alpha-plugins`.