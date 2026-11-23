/*
 * Exercises the production update copy loop using controlled streams and time.
 * Covers progressing slow downloads, stalled reads and finite resource bounds.
 */
package com.cjtrowbridge.apparat;

import java.io.InputStream;
import java.io.OutputStream;
import java.io.IOException;
import java.net.SocketTimeoutException;

public final class UpdateTransferTest {
    static final class Clock implements UpdateTransfer.State {
        long time;
        boolean cancelled;
        long received;
        int reports;
        public void progress(long bytes) {
            if (bytes < received) throw new AssertionError("progress regressed");
            received = bytes;
            reports++;
        }
        public long now() {
            return time;
        }
        public boolean cancelled() {
            return cancelled;
        }
    }

    static final class Stream extends InputStream {
        final Clock clock;
        final long delay;
        long remaining;
        Stream(Clock clock, long bytes, long delay) {
            this.clock = clock;
            this.remaining = bytes;
            this.delay = delay;
        }
        public int read() {
            throw new AssertionError("Expected buffered reads");
        }
        public int read(byte[] buffer) {
            clock.time += delay;
            if (remaining == 0) {
                return -1;
            }
            int count = (int) Math.min(remaining, buffer.length);
            remaining -= count;
            return count;
        }
    }

    static final class Sink extends OutputStream {
        long count;
        public void write(int value) {
            count++;
        }
        public void write(byte[] buffer, int offset, int length) {
            count += length;
        }
    }

    static void expect(String code, InputStream input, Clock clock) throws Exception {
        try {
            UpdateTransfer.copy(input, new Sink(), clock);
            throw new AssertionError("Expected " + code);
        } catch (UpdateTransfer.Failure error) {
            if (!code.equals(error.code)) {
                throw new AssertionError(error.code);
            }
        }
    }

    public static void main(String[] args) throws Exception {
        Clock slow = new Clock();
        Sink sink = new Sink();
        UpdateTransfer.copy(new Stream(slow, 256 * 1024, 40000), sink, slow);
        if (sink.count != 256 * 1024 || slow.time <= 120000 || slow.received != sink.count || slow.reports < 3) {
            throw new AssertionError("Slow progressing transfer must finish");
        }
        Clock deadline = new Clock();
        expect("unable_timeout", new Stream(deadline, 1, UpdateTransfer.MAX_DURATION_MS + 1), deadline);
        Clock cancelled = new Clock();
        cancelled.cancelled = true;
        expect("update_cancelled", new Stream(cancelled, 1, 0), cancelled);
        Clock size = new Clock();
        expect("unable_size", new Stream(size, UpdateTransfer.MAX_BYTES + 1, 0), size);
        expect("unable_timeout", new InputStream() {
            public int read() throws IOException {
                throw new SocketTimeoutException();
            }
        }, new Clock());
        System.out.println("PASS: slow progress, total timeout, cancellation, size bound and stalled read");
    }
}
