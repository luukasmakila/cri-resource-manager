# Memtierd demo

## What was showcased

The demo showcases the differenece between how low priority and high priority workloads are treated when using Memtierd as the memory manager. Low priority workloads and high priority workloads are defined by giving your deployments the following annotations:

```yaml
class.memtierd.nri: "high-prio"
# or
class.memtierd.nri: "low-prio"
```

This annotation defines whether the workload will be swapped agressively (low-prio) or more moderately (high-prio). More agressive swapping will lead to an increase in the number of page faults the process will have.

## About the metrics

Total Memory Saved (G)
- Tells how much RAM is being saved by swapping out the idle workloads.

Total Memory Saved (%)
- Tells how big the total memory saved is in comparsin to the overall memory of the system.

Compressed (%)
- Tells how well the data is being compressed.

RAM vs Swap
- Tells how the memory is being distributed between RAM and Swap.

Page faults
- Tells how many new page faults happen in between the requests from Grafana. This is a way to express the possible performance hit workloads experience if being tracked by Memtierd.

![alt text](https://github.com/luukasmakila/cri-resource-manager/blob/memtier-nri/demo/memtierd/memtierd-demo.png)