---
plan_id: 2026-11-23-23-50-00_worker-inference-inventory
title: Advertise worker Ollama inventory and measured capacity
summary: Discover worker loopback Ollama, expose authorized inventory and show resource and capability sections in Cluster.
status: past
created_at: 2026-11-23-23-50-00
---

# Worker inference inventory

Key: `[ ]` pending task, `[x]` completed task, `[?]` needs validation, `[-]` closed task

User-authorized Phases 12 and 14 discovery and safe resource reporting. Headless workers probe the default or saved port at startup and refresh while running. Only authenticated members receive snapshots and only admins configure ports. Discovery does not submit inference, load models or install software.

- [x] Verify loopback identity and inventory, trial failures, persistence and freshness. Collect CPU, RAM, swap, disks and supported VRAM.
- [x] Connect worker API, Cluster detail, discovery forms, workload sections, selector icons and action directory.
- [x] Exercise authenticated access, inventory fixtures and wide/narrow rendering; leave live trashcan0 configuration to the operator.

Race-enabled worker API, inventory, persistence, runtime, GUI and HUD suites passed. Inspected the 1280×800 and 390×844 device detail renders. Live Ollama configuration on trashcan0 is reserved for the operator.
