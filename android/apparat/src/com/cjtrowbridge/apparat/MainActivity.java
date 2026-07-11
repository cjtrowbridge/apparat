package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.content.Intent;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.provider.Settings;
import android.util.Log;
import android.view.Window;
import android.view.WindowManager;
import android.widget.Toast;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.security.MessageDigest;
import java.util.Locale;

import go.Seq;
import com.cjtrowbridge.apparat.apparatmobile.Apparatmobile;
import com.cjtrowbridge.apparat.apparatmobile.EbitenView;

public class MainActivity extends Activity {
    private static final String TAG = "ApparatUpdate";
    private static final String UPDATE_URL = "https://raw.githubusercontent.com/cjtrowbridge/apparat/main/releases/android/arm64/apparat/latest.apk";
    private static final String UPDATE_AUTHORITY = "com.cjtrowbridge.apparat.update";
    private static final String APK_MIME_TYPE = "application/vnd.android.package-archive";
    private EbitenView view;
    private File downloadedUpdate;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Seq.setContext(getApplicationContext());
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(WindowManager.LayoutParams.FLAG_FULLSCREEN, WindowManager.LayoutParams.FLAG_FULLSCREEN);
        Apparatmobile.ready();
        view = new EbitenView(this);
        view.setFocusable(true);
        view.setFocusableInTouchMode(true);
        setContentView(view);
        Apparatmobile.registerUpdater(new com.cjtrowbridge.apparat.apparatmobile.Updater() {
            @Override
            public void checkForUpdate() {
                Log.i(TAG, "checkForUpdate bridge invoked");
                MainActivity.this.checkForUpdate(true);
            }
        });
        Log.i(TAG, "Updater bridge registered");
        checkForUpdate(false);
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (view != null) {
            view.resumeGame();
        }

    }

    @Override
    protected void onPause() {
        super.onPause();

        if (view != null) {
            view.suspendGame();
        }
    }



    private void checkForUpdate(final boolean userInitiated) {
        Log.i(TAG, "Starting update check userInitiated=" + userInitiated);
        if (userInitiated) {
            reportUpdateStatus("Checking...");
            showToast("Checking Apparat update...");
        }
        new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    File apk = downloadLatestApk();
                    downloadedUpdate = apk;
                    String installedHash = sha256(new File(getApplicationInfo().sourceDir));
                    String downloadedHash = sha256(apk);
                    if (installedHash.equals(downloadedHash)) {
                        if (userInitiated) {
                            reportUpdateStatus("Already current");
                            showToast("Already current: installed and GitHub APK hashes match (" + shortHash(installedHash) + ")");
                        } else {
                            Log.i(TAG, "Startup update check found current APK " + shortHash(installedHash));
                        }
                        return;
                    }
                    runOnUiThread(new Runnable() {
                        @Override
                        public void run() {
                            requestPermissionOrInstall();
                        }
                    });
                } catch (Exception error) {
                    Log.w(TAG, "Update check failed", error);
                    if (userInitiated) {
                        reportUpdateStatus("Update failed");
                        showToast("Update check failed: " + error.getMessage());
                    }
                }
            }
        }).start();
    }

    private File downloadLatestApk() throws Exception {
        File apk = new File(getCacheDir(), "latest.apk");
        HttpURLConnection connection = (HttpURLConnection) new URL(UPDATE_URL + "?t=" + System.currentTimeMillis()).openConnection();
        connection.setUseCaches(false);
        connection.addRequestProperty("Cache-Control", "no-cache");
        connection.setConnectTimeout(15000);
        connection.setReadTimeout(60000);
        connection.setInstanceFollowRedirects(true);
        int status = connection.getResponseCode();
        if (status < 200 || status >= 300) {
            throw new IllegalStateException("HTTP " + status);
        }
        try (InputStream in = connection.getInputStream(); FileOutputStream out = new FileOutputStream(apk)) {
            byte[] buffer = new byte[64 * 1024];
            int read;
            while ((read = in.read(buffer)) != -1) {
                out.write(buffer, 0, read);
            }
        } finally {
            connection.disconnect();
        }
        return apk;
    }

    private void requestPermissionOrInstall() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && !getPackageManager().canRequestPackageInstalls()) {
            reportUpdateStatus("Permission needed");
            showToast("Allow Apparat to install updates, then tap Update again.");
            Intent intent = new Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES);
            intent.setData(Uri.parse("package:" + getPackageName()));
            startActivity(intent);
            return;
        }
        installDownloadedUpdate();
    }

    private void installDownloadedUpdate() {
        if (downloadedUpdate == null || !downloadedUpdate.exists()) {
            reportUpdateStatus("No update ready");
            showToast("No downloaded update is ready");
            return;
        }
        Uri uri = Uri.parse("content://" + UPDATE_AUTHORITY + "/latest.apk");
        Intent intent = new Intent(Intent.ACTION_VIEW);
        intent.setDataAndType(uri, APK_MIME_TYPE);
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        try {
            startActivity(intent);
            reportUpdateStatus("Installer opened");
        } catch (Exception error) {
            reportUpdateStatus("Installer failed");
            showToast("Could not open Android installer: " + error.getMessage());
        }
    }

    private void reportUpdateStatus(String message) {
        Log.i(TAG, "Update status: " + message);
        try {
            Apparatmobile.reportUpdateStatus(message);
        } catch (Exception error) {
            Log.w(TAG, "Could not report update status to HUD", error);
        }
    }

    private void showToast(final String message) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                Toast.makeText(MainActivity.this, message, Toast.LENGTH_LONG).show();
            }
        });
    }

    private int dp(int value) {
        return Math.round(value * getResources().getDisplayMetrics().density);
    }

    private static String shortHash(String hash) {
        if (hash == null || hash.length() < 12) {
            return hash;
        }
        return hash.substring(0, 12);
    }

    private static String sha256(File file) throws Exception {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        try (InputStream in = new FileInputStream(file)) {
            byte[] buffer = new byte[64 * 1024];
            int read;
            while ((read = in.read(buffer)) != -1) {
                digest.update(buffer, 0, read);
            }
        }
        StringBuilder builder = new StringBuilder();
        for (byte value : digest.digest()) {
            builder.append(String.format(Locale.ROOT, "%02x", value));
        }
        return builder.toString();
    }

}
