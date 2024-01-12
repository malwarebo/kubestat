# kubestat

Simple k8s status cli tool

## Usage

Clone this repository and run:

```bash
make
sudo make install
```

Run the command `kubestat` from anywhere in the terminal.

## Namespace filter

Need to filter by namespace `toto` like you do with `kubectl get pods -n toto` ?

- modify `kubernetes.go`
  - replace `.Pods("")`
  - with `.Pods("toto")`
- run `make`
- then run `sudo make install`
