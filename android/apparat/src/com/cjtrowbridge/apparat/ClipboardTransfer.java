/*
 * Performs editor clipboard operations with Android's foreground window.
 * Retries a transient focus or access failure during the original user gesture.
 */
package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.content.ClipData;
import android.content.ClipboardManager;
import android.net.Uri;
import android.os.Handler;
import android.os.Looper;
import android.os.SystemClock;

import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;

final class ClipboardTransfer implements Runnable {
    private final Activity activity;
    private final String copy;
    private final CountDownLatch done = new CountDownLatch(1);
    private final long deadline = SystemClock.uptimeMillis() + 900;
    private String result = "";
    private Exception failure;

    private ClipboardTransfer(Activity activity, String copy) {
        this.activity = activity;
        this.copy = copy;
    }

    static String run(Activity activity, String copy) throws Exception {
        ClipboardTransfer operation = new ClipboardTransfer(activity, copy);
        activity.runOnUiThread(operation);
        if (!operation.done.await(2, TimeUnit.SECONDS)) {
            throw new Exception("Clipboard timed out waiting for the Android window. Return to Apparat and retry.");
        }
        if (operation.failure != null) throw operation.failure;
        return operation.result;
    }

    @Override public void run() {
        boolean retry = SystemClock.uptimeMillis() + 100 < deadline;
        try {
            if (activity.isFinishing() || activity.isDestroyed() || SystemClock.uptimeMillis() >= deadline) {
                throw new Exception("Clipboard request expired while Apparat was losing focus. Return to the field and retry.");
            }
            if (!activity.hasWindowFocus()) {
                if (retry) {
                    new Handler(Looper.getMainLooper()).postDelayed(this, 80);
                    return;
                }
                throw new Exception("Android withheld clipboard access while another window had focus. Return to the Apparat field and retry.");
            }
            ClipboardManager manager = activity.getSystemService(ClipboardManager.class);
            if (manager == null) throw new Exception("The Android clipboard service is unavailable. Reopen Apparat and retry.");
            if (copy != null) {
                manager.setPrimaryClip(ClipData.newPlainText("Apparat", copy));
            } else {
                result = read(manager);
            }
        } catch (SecurityException denied) {
            if (retry) {
                new Handler(Looper.getMainLooper()).postDelayed(this, 100);
                return;
            }
            failure = new Exception("Android denied clipboard access after a foreground retry. Copy plain text from an app in the same Android profile.");
        } catch (RuntimeException unavailable) {
            failure = new Exception("The Android clipboard service failed during the transfer. Copy the text again and retry.");
        } catch (Exception explained) {
            failure = explained;
        }
        done.countDown();
    }

    private String read(ClipboardManager manager) throws Exception {
        ClipData clip = manager.getPrimaryClip();
        if (clip == null || clip.getItemCount() == 0) return "";
        ClipData.Item item = clip.getItemAt(0);
        if (item.getText() != null) return item.getText().toString();
        Uri uri = item.getUri();
        if (uri != null && ("https".equals(uri.getScheme()) || "http".equals(uri.getScheme()))) {
            return uri.toString();
        }
        throw new Exception("Nothing to paste. This Android clipboard item contains no text. Select text in the source app and copy it.");
    }
}
