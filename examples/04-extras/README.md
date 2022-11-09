# 04 - Ports and Volumes

In advanced setups the workload might use files and volumes or serve incoming requests on selected ports.

Such requirements can be expressed in `score.yaml` file:

```yaml
apiVersion: score.sh/v1b1

metadata:
  name: web-app

service:
  ports:
    www:
      port: 80

containers:
  hello:
    image: nginx
    files:
      - target: /usr/share/nginx/html/index.html
        mode: "644"
        content: ["${resources.env.MESSAGE}"]

resources:
  env:
    type: environment
    properties:
      MESSAGE:
        type: string
  dns:
    type: dns
```

This example also uses an extensions file, called `humanitec.yaml`, that contains additional hints for `score-humanitec` CLI tool. This information would help the CLI tool to add proper routes to the deployment delta, so the service would be available to the outer world:

```yaml
apiVersion: humanitec.org/v1b1

service:
  routes:
    http:
      "/":
        type: prefix
        from: ${resources.dns}
        port: 80
```

To prepare a new Humanitec deployment delta from this `score.yaml` file, use `score-humanitec` CLI tool:

```console
$ score-humanitec run -f ./score.yaml --extensions ./humanitec.yaml --env test-env
```

Output JSON can be used as a payload for the [Create a new Delta](https://api-docs.humanitec.com/#tag/Delta/paths/~1orgs~1%7BorgId%7D~1apps~1%7BappId%7D~1deltas/post) Humanitec API call:

```json
{
  "metadata": {
    "env_id": "test-env",
    "name": "Auto-generated (SCORE)"
  },
  "modules": {
    "add": {
      "web-app": {
        "externals": {
          "dns": {
            "type": "dns"
          }
        },
        "profile": "humanitec/default-module",
        "spec": {
          "containers": {
            "hello": {
              "files": {
                "/usr/share/nginx/html/index.html": {
                  "mode": "644",
                  "value": "${values.MESSAGE}"
                }
              },
              "id": "hello",
              "image": "nginx"
            }
          },
          "ingress": {
            "rules": {
              "externals.dns": {
                "http": {
                  "/": {
                    "port": 80,
                    "type": "prefix"
                  }
                }
              }
            }
          },
          "service": {
            "ports": {
              "www": {
                "container_port": 80,
                "protocol": "TCP",
                "service_port": 80
              }
            }
          }
        }
      }
    }
  }
}
```

When deploying this service with `Humanitec`, make sure that the shared application value called `MESSAGE` is set for the target environment.