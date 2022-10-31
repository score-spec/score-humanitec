# 01 - Hello World!

In this basic example there is a simple compose service based on `busybox` Docker image described in a `score.yaml` file:

```yaml
apiVersion: score.dev/v1b1

metadata:
  name: hello-world

containers:
  hello:
    image: busybox
    command: ["/bin/sh"]
    args: ["-c", "while true; do echo Hello World!; sleep 5; done"]
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
                "while true; do echo Hello World!; sleep 5; done"
              ],
              "command": [
                "/bin/sh"
              ],
              "id": "hello",
              "image": "busybox"
            }
          }
        }
      }
    }
  }
}
```
