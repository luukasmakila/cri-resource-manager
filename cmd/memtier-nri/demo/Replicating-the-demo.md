## Replicating the demo

### Prerequisites
- NRI enabled on your container runtime
- Grafana dashboard ready to go
- Grafana [infinity](https://grafana.com/grafana/plugins/yesoreyeram-infinity-datasource/) data source plugin downloaded.

### Configuring the Memtierd NRI plugin:

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

### Configuring the workloads

Deploy the high priority workloads:
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: meme-pod-highprio
  labels:
    app: meme-pod-highprio
  annotations:
    class.memtierd.nri: "high-prio"
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
    class.memtierd.nri: "low-prio"
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
