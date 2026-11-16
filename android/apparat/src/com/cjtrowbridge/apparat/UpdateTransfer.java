/*
 * Copies update bytes with bounded time, size and cancellation.
 * Uses injectable monotonic time for deterministic slow-transfer tests.
 */
package com.cjtrowbridge.apparat;

import java.io.InputStream;
import java.io.OutputStream;
import java.io.IOException;
import java.net.SocketTimeoutException;

final class UpdateTransfer {
    static final int READ_TIMEOUT_MS = 60000;
    static final long MAX_DURATION_MS = 15L * 60 * 1000;
    static final long MAX_BYTES = 128L * 1024 * 1024;

    interface State {
        long now();
        boolean cancelled();
    }

    static final class Failure extends IOException {
        final String code;
        final long bytes;
        final long elapsed;

        Failure(String code, long bytes, long elapsed) {
            super(code);
            this.code = code;
            this.bytes = bytes;
            this.elapsed = elapsed;
        }
    }

    static void copy(InputStream input, OutputStream output, State state) throws IOException {
        long started = state.now();
        long total = 0;
        byte[] buffer = new byte[64 * 1024];
        while (true) {
            check(state, started, total);
            int count;
            try {
                count = input.read(buffer);
            } catch (SocketTimeoutException error) {
                throw new Failure("unable_timeout", total, state.now() - started);
            }
            check(state, started, total);
            if (count == -1) {
                return;
            }
            total += count;
            if (total > MAX_BYTES) {
                throw new Failure("unable_size", total, state.now() - started);
            }
            output.write(buffer, 0, count);
        }
    }

    private static void check(State state, long started, long total) throws Failure {
        long elapsed = state.now() - started;
        if (state.cancelled()) {
            throw new Failure("update_cancelled", total, elapsed);
        }
        if (elapsed > MAX_DURATION_MS) {
            throw new Failure("unable_timeout", total, elapsed);
        }
    }
}
