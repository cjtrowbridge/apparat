# Messaging Package

This package owns durable local messaging primitives.

It provides inbox, outbox, replay, duplicate tracking, cursor, and retry scheduling foundations used by future API and transport adapters.

Messaging code should preserve at-least-once delivery assumptions and expose idempotent state to higher layers.
