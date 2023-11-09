# kubestat

Simple k8s status cli tool

Clone the repository

Run locally within the package scope

```bash
./kubestat
```

or

Clone this repository and run:

```bash
cd kubestat
go build
sudo cp kubestat /usr/local/bin
```

Run the command `kubestat` from anywhere in the CLI.

# namespace filter

Need to filter by namespace `toto` like you do with `kubectl get pods -n toto` ?

- install go
- modify `main.go`
  - replace `.Pods("")`
  - with `.Pods("toto")`
- run `go build`
