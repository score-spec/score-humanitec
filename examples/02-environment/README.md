# 02 - Environment Variables

`Humanitec` provides a flexible application configuration mechanism through the [Shared Application Values](https://docs.humanitec.com/using-humanitec/work-with-apps/define-app-values-and-secrets). These values can be referenced through a special score `environment` resource type:

```yaml
apiVersion: score.sh/v1b1

metadata:
  name: hello-world

containers:
  hello:
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo Hello $${FRIEND}!; sleep 5; done"]
    variables:
      FRIEND: ${resources.env.NAME}

resources:
  env:
    type: environment
    properties:
      NAME:
        type: string
        default: World
```

To prepare a new Humanitec deployment delta from this `score.yaml` file, use `score-humanitec` CLI tool:

```console
$ score-humanitec run -f ./score.yaml --env test-env
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
      "hello-world": {
        "profile": "humanitec/default-module",
        "spec": {
          "containers": {
            "hello": {
              "args": [
                "-c",
                "while true; do echo Hello $${FRIEND}!; sleep 5; done"
              ],
              "command": [
                "/bin/sh"
              ],
              "id": "hello",
              "image": "busybox",
              "variables": {
                "FRIEND": "${values.NAME}"
              }
            }
          }
        }
      }
    }
  }
}
```

When deploying this service with `Humanitec`, make sure the shared application value called `NAME` is created and set for the target environment.
