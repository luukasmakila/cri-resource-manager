# Memtierd NRI plugin

### Making your own image

To add your own memtierd configuration templates, move the template files to the /templates directory. These templates can then be used by specifying the name of the template in the pod annotations of the pod you want memtierd to track.

Run:
```console
docker build . -t memtierd-nri
```

### Using memtierd with your deployments

The following yaml file give the pod annotations "use-memtierd" and "template-memtierd". The "use-memtierd" is required to tell the plugin to tell the plugin to run memtierd and can be set to any string (ex. "true"). The template annotaion "template-memtierd" is where you can specify the name of the configuration template you would like to use with this deployment. If you have made your own configurations and are using your own image, you can specify the name of your configuration here.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: some-pod
  labels:
    app: some-pod
  annotations:
    use-memtierd: "true"
    template-memtierd: "memtierd-age-swapidle.yaml"
spec:
  containers:
  - name: some-pod
    image: some-image
    imagePullPolicy: Always
    securityContext:
      privileged: true
    volumeMounts:
      - name: nri-socket
        mountPath: /var/run/nri.sock
      - name: host-volume
        mountPath: /host
      - name: host-bitmap
        mountPath: /sys/kernel/mm/page_idle/bitmap
  volumes:
    - name: nri-socket
      hostPath:
        path: /var/run/nri.sock
    - name: host-volume
      hostPath:
        path: /
    - name: host-bitmap
      hostPath:
        path: /sys/kernel/mm/page_idle/bitmap
```
