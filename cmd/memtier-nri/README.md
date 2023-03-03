# Memtierd NRI plugin

## Making your own image

To add your own memtierd configuration templates, move the template files to the /templates directory. These templates can then be used by specifying the name of the template in the pod annotations of the pod you want memtierd to track.

Run:
```console
docker build . -t memtierd-nri
```

## Running a self compiled version locally

Prerequisites:
- NRI server enabled in containerd see [here](#enabling-nri-in-containerd)

To compile your own version run:
```console
go build .
```

Then move the output to the plugin path specified in your /etc/containerd/config.toml file:
```toml
plugin_path = "/opt/nri/plugins"
````

You also need to specify an index for the plugin in the plugin name. The index is kind of like a priority for the plugin to be executed in case you have multiple plugins.

Here is an example:
```console
mv memtier-nri /opt/nri/plugins/10-memtier-nri
```

Then just run:
```console
/opt/nri/plugin/10-memtier-nri
```

After that you should see something like the following:
```console
INFO   [0000] Created plugin 10-memtier-nri (10-memtier-nri, handles RunPodSandbox,StopPodSandbox,RemovePodSandbox,CreateContainer,PostCreateContainer,StartContainer,PostStartContainer,UpdateContainer,PostUpdateContainer,StopContainer,RemoveContainer)
INFO   [0000] Registering plugin 10-memtier-nri...
...
```

Now the plugin is ready to answer to events happening in your the cluster.

## <a name="enabling-nri-in-containerd"></a> Enabling NRI in containerd

Replace the containerd version in the system with 1.7-beta.

In the config.toml look for the "io.containerd.nri.v1.nri" and replace "disable = true" with "disable = false".

```console
vim /etc/containerd/config.toml
```

Create a configuration or the NRI server will not start:
```console
sudo sh -c "mkdir -p /etc/nri; touch /etc/nri/nri.conf; systemctl restart containerd"
```

## Using memtierd with your deployments

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
