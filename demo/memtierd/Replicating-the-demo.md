# Replicating the Memtierd demo

## Prerequisites
- Ubuntu system with a Swap
- Python 3
- NRI enabled on your container runtime (Containerd/CRI-O)
- Grafana dashboard ready to go
- Grafana [infinity](https://grafana.com/grafana/plugins/yesoreyeram-infinity-datasource/) data source plugin downloaded

## Installing the Memtierd grafana dashboard

- Go to the "Data Sources" tab and apply the Infinity data source (needed to handle the showcase the json data)
- Go to the "Dashboards" section on Grafana
- Click "New" and then "Import"
- Download the "memtierd-demo-grafana-dashboard.json" file and import it
- Select Infinity data source as the data source

## Running the API

Edit the "path" variables found on the top of the main.py file to point to the correct data files in data/ aswell as zram and meminfo paths. When ran with the default workloads /tmp/memtierd directory will be created to read the output from so unless the workload configurations are changed, those paths won't need editing.

Install FastAPI:
```
console pip install fastapi
```

Start the API with:
```console
uvicorn main:app --reload
```

Make sure the files in data/ are in the correct format:

Page faults files:
```json
{
    "page_faults_highprio/lowprio_1": [
    ]
}
```

Time series files:
```json
{
    "time_series_highprio/lowprio_1": [
    ]
}
```

## Configuring the Memtierd NRI plugin:

Deploy the Memtier NRI plugin pod:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-memtier
  labels:
    app: pod-memtier
spec:
  hostPID: true
  containers:
  - name: pod-memtier
    image: # TODO
    imagePullPolicy: Always
    securityContext:
      privileged: true
    volumeMounts:
      - name: nri-socket
        mountPath: /var/run/nri/nri.sock
      - name: host-volume
        mountPath: /host
      - name: host-bitmap
        mountPath: /sys/kernel/mm/page_idle/bitmap
  volumes:
    - name: nri-socket
      hostPath:
        path: /var/run/nri/nri.sock
    - name: host-volume
      hostPath:
        path: /
    - name: host-bitmap
      hostPath:
        path: /sys/kernel/mm/page_idle/bitmap
```

## Deploying the workloads

Deploy the high priority workloads:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: meme-pod-highprio
  labels:
    app: meme-pod-highprio
  annotations:
    class.memtierd.nri: "high-prio" # <-- this tells memtierd that these workloads are "high priority"
spec:
  containers:
  - name: meme-pod-highprio-1-container
    image: # TODO
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
  - name: meme-pod-highprio-2-container
    image: # TODO
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

Deploy the low priority workloads:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: meme-pod-lowprio
  labels:
    app: meme-pod-lowprio
  annotations:
    class.memtierd.nri: "low-prio" # <-- This tells memtierd that these workloads are "low priority"
spec:
  containers:
  - name: meme-pod-lowprio-1-container
    image: # TODO
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
  - name: meme-pod-lowprio-2-container
    image: # TODO
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
