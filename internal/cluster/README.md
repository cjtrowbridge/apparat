# Cluster Package

This package owns local cluster-directory persistence.

It stores device profiles, roles, capabilities, reachability, and related metadata used by the Cluster and Routing surfaces.

The package must not decide UI presentation, network transport behavior, or workload scheduling policy. It provides repository behavior that higher layers can validate and render.
