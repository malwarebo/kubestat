# kubestat

A CLI that scans Kubernetes pods and prints only the unhealthy ones. It reads
the current context from your kubeconfig (`~/.kube/config`).

## Install

```bash
make
sudo make install
```

## Usage

Scan all namespaces:

```bash
kubestat
```

Scan a single namespace:

```bash
kubestat -n <namespace>
```

Set the restart count above which a pod is flagged (default: 5):

```bash
kubestat --restarts 10
```

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-n`, `--namespace` | all namespaces | Limit the scan to one namespace |
| `--restarts` | `5` | Flag pods whose total restart count exceeds this value |

## What is reported

A pod is shown when any of these is true:

- A container is waiting in a failure state (`CrashLoopBackOff`,
  `ImagePullBackOff`, `ErrImagePull`, `CreateContainerConfigError`, and similar).
- A container terminated with a non-zero exit code (`Error`, `OOMKilled`, etc.).
- The pod phase is `Failed` or `Pending`.
- The pod is `Running` but a container is not `Ready`.
- The total restart count exceeds `--restarts`.

`Running`-and-ready pods and `Succeeded` pods are hidden. Hard failures are
printed in red, warnings in yellow.

## Output

```
Context: prod-cluster  |  Scanning all namespaces for unhealthy pods...

  NAMESPACE | POD                  | STATUS           | RESTARTS | AGE
------------+----------------------+------------------+----------+-------
  default   | api-7d9f8c-abc12     | CrashLoopBackOff |       12 | 3h20m
  payments  | worker-5f4b2a-xyz98  | OOMKilled        |        4 | 1d2h

Scanned 142 pod(s) across 9 namespace(s): 2 unhealthy.
```

## Test

```bash
go test ./...
```
