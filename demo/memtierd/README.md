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

Total memory saved (G)
- Tells how much RAM is being saved by swapping out the idle workloads

Compressed (%)
- Tells how well the data is being compressed

RAM vs Swap
- Tells how the memory is being distributed between RAM and Swap

Page faults
- Tells how many new page faults happen in a certain time period