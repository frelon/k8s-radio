# K8s radio

An operator that exposes SDR devices to your kubernetes cluster using device-plugin.

## Quickstart

This will:

- Spin up a kind cluster
- Build the container images
- Load the images into the kind cluster
- Deploy the manager, device-plugin and sample RtlSdrReceiver

```
make cluster
make docker-build
make load
make deploy
```

A sample RtlSdrReceiver that will tune to 101.9Mhz and expose the I/Q stream on the host-port 1234.

```yml
apiVersion: radio.frelon.se/v1beta1
kind: RtlSdrReceiver
metadata:
  name: rtlsdrreceiver-sample
spec:
  version: v3
  frequency: "101.9M"
  port:
    containerPort: 1234
    hostPort: 1234
    protocol: TCP
```
