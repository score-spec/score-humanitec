# 03 - Dependencies

`Humanitec` makes sure that all the dependencies that the workload needs are up and available during the deployment. These could be other workloads and services, resources (such as database, or DNS records), or even whole environments, scripted with Terraform.

In this example `score.yaml` file describes a workload that needs a `backend` (another workload), a PostgreSQL database instance, and a shared DNS record:

```yaml
apiVersion: score.sh/v1b1

metadata:
  name: service-a

containers:
  service-a:
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo service-a: Hello $${FRIEND}! Connecting to $${CONNECTION_STRING}...; sleep 10; done"]
    variables:
      FRIEND: ${resources.backend.name}
      CONNECTION_STRING: postgresql://${resources.db.user}:${resources.db.password}@${resources.db.host}:${resources.db.port}/${resources.db.name}

resources:
  db:
    type: postgres
  dns:
    type: dns
  backend:
    type: service
```

This example also uses an extensions file, called `humanitec.yaml`, that contains additional hints for `score-humanitec` CLI tool. This information would help the CLI tool to resolve the resources properly.

```yaml
apiVersion: humanitec.org/v1b1

resources:
    db:
        scope: external
    dns:
        scope: shared
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
      "service-a": {
        "externals": {
          "db": {
            "type": "postgres"
          }
        },
        "profile": "humanitec/default-module",
        "spec": {
          "containers": {
            "service-a": {
              "args": [
                "-c",
                "while true; do echo service-a: Hello $${FRIEND}! Connecting to $${CONNECTION_STRING}...; sleep 10; done"
              ],
              "command": [
                "/bin/sh"
              ],
              "id": "service-a",
              "image": "busybox",
              "variables": {
                "CONNECTION_STRING": "postgresql://${externals.db.user}:${externals.db.password}@${externals.db.host}:${externals.db.port}/${externals.db.name}",
                "FRIEND": "${modules.backend.name}"
              }
            }
          }
        }
      }
    }
  },
  "shared": [
    {
      "path": "/dns",
      "op": "add",
      "value": {
        "type": "dns"
      }
    }
  ]
}
```

When deploying this service with `Humanitec`, make sure that all the dependencies are properly defined and configured for the target environment.
