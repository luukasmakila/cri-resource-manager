# Memtierd NRI plugin

## Prerequisities
- NRI enabled on your container runtime

## Making your own image

```console
# Build the image
docker build . -t memtier-nri

# Then tag and push the image to your registry
docker tag memtier-nri <your registry>
docker push <your registry>
```

See [running Memtierd NRI plugin in a pod](#running-memtierd-nri-plugin-in-a-pod) on how to deploy it.

## Running a self compiled version locally

To compile your own version run:
```console
go build .
```

Then move the output to the plugin path specified in your /etc/containerd/config.toml file:
```toml
plugin_path = "/opt/nri/plugins"
````

You also need to specify an index for the plugin in the plugin name. The index is like an priority for the plugin to be executed in case you have multiple plugins.

For example:
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

Now the plugin is ready to recognize events happening in the cluster.

## <a name="running-memtierd-nri-plugin-in-a-pod"></a> Running Memtierd NRI plugin in a pod

To run Memtierd NRI plugin in a pod change the image in templates/pod-memtier-nri.yaml to point at your image and then deploy the pod to your cluster:

```console
kubectl apply -f templates/pod-memtier-nri.yaml
```

## Using memtierd with your deployments

Workload configurations are defined with the "class.memtierd.nri" annotation. Now for example the following annotation:

```yaml
class.memtierd.nri: "high-prio-configuration"
```

Starts memtierd for the workload with the configuration found in "templates/high-prio-configuration.yaml"

Currently the plugin supports "low-prio" and "high-prio" workloads. Low-prio means that the workload can be swapped out alot while high-prio means the exact opposite. You can obviously add whatever configurations you need in the code. The code looks in the templates/ directory for the configurations.
