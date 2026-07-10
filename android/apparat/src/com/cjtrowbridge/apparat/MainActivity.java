package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.content.Intent;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.provider.Settings;
import android.view.Gravity;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;
import android.widget.Button;
import android.widget.FrameLayout;
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
    private static final String UPDATE_URL = "https://raw.githubusercontent.com/cjtrowbridge/apparat/main/releases/android/arm64/apparat/latest.apk";
    private static final String UPDATE_AUTHORITY = "com.cjtrowbridge.apparat.update";
    private static final String APK_MIME_TYPE = "application/vnd.android.package-archive";
    private EbitenView view;
    private Button updateButton;
    private Handler mainHandler;
    private File downloadedUpdate;
    private final Runnable updateButtonVisibility = new Runnable() {
        @Override
        public void run() {
            refreshUpdateButtonVisibility();
            mainHandler.postDelayed(this, 250);
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Seq.setContext(getApplicationContext());
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(WindowManager.LayoutParams.FLAG_FULLSCREEN, WindowManager.LayoutParams.FLAG_FULLSCREEN);
        mainHandler = new Handler(Looper.getMainLooper());
        Apparatmobile.ready();
        view = new EbitenView(this);
        view.setFocusable(true);
        view.setFocusableInTouchMode(true);
        setContentView(view);
        updateButton = new Button(this);
        updateButton.setText("Check for update");
        updateButton.setAllCaps(false);
        updateButton.setMinHeight(dp(48));
        updateButton.setMinimumHeight(dp(48));
        updateButton.setVisibility(View.GONE);
        updateButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View button) {
                checkForUpdate();
            }
        });
        FrameLayout.LayoutParams updateLayout = new FrameLayout.LayoutParams(
                FrameLayout.LayoutParams.WRAP_CONTENT,
                dp(48),
                Gravity.TOP | Gravity.START);
        addContentView(updateButton, updateLayout);
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (view != null) {
            view.resumeGame();
        }
        if (mainHandler != null) {
            mainHandler.removeCallbacks(updateButtonVisibility);
            mainHandler.post(updateButtonVisibility);
        }
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (mainHandler != null) {
            mainHandler.removeCallbacks(updateButtonVisibility);
        }
        if (view != null) {
            view.suspendGame();
        }
    }

    private void refreshUpdateButtonVisibility() {
        if (updateButton == null) {
            return;
        }
        try {
            boolean settingsActive = "settings".equals(Apparatmobile.activeTab());
            if (settingsActive) {
                positionUpdateButton();
            }
            updateButton.setVisibility(settingsActive ? View.VISIBLE : View.GONE);
        } catch (Exception error) {
            updateButton.setVisibility(View.GONE);
        }
    }

    private void positionUpdateButton() {
        int width = view.getWidth();
        int height = view.getHeight();
        if (width <= 0 || height <= 0) {
            return;
        }
        FrameLayout.LayoutParams layout = (FrameLayout.LayoutParams) updateButton.getLayoutParams();
        layout.width = Math.max(dp(160), (int) Apparatmobile.updateButtonW(width, height));
        layout.height = Math.max(dp(48), (int) Apparatmobile.updateButtonH(width, height));
        layout.leftMargin = (int) Apparatmobile.updateButtonX(width, height);
        layout.topMargin = (int) Apparatmobile.updateButtonY(width, height);
        updateButton.setLayoutParams(layout);
    }

    private void checkForUpdate() {
        Toast.makeText(this, "Checking Apparat update...", Toast.LENGTH_SHORT).show();
        new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    File apk = downloadLatestApk();
                    downloadedUpdate = apk;
                    String installedHash = sha256(new File(getApplicationInfo().sourceDir));
                    String downloadedHash = sha256(apk);
                    if (installedHash.equals(downloadedHash)) {
                        showToast("Already current: installed and GitHub APK hashes match (" + shortHash(installedHash) + ")");
                        return;
                    }
                    runOnUiThread(new Runnable() {
                        @Override
                        public void run() {
                            requestPermissionOrInstall();
                        }
                    });
                } catch (Exception error) {
                    showToast("Update check failed: " + error.getMessage());
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
            Toast.makeText(this, "Allow Apparat to install updates, then tap Update again.", Toast.LENGTH_LONG).show();
            Intent intent = new Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES);
            intent.setData(Uri.parse("package:" + getPackageName()));
            startActivity(intent);
            return;
        }
        installDownloadedUpdate();
    }

    private void installDownloadedUpdate() {
        if (downloadedUpdate == null || !downloadedUpdate.exists()) {
            Toast.makeText(this, "No downloaded update is ready", Toast.LENGTH_LONG).show();
            return;
        }
        Uri uri = Uri.parse("content://" + UPDATE_AUTHORITY + "/latest.apk");
        Intent intent = new Intent(Intent.ACTION_VIEW);
        intent.setDataAndType(uri, APK_MIME_TYPE);
        intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
        intent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        try {
            startActivity(intent);
        } catch (Exception error) {
            Toast.makeText(this, "Could not open Android installer: " + error.getMessage(), Toast.LENGTH_LONG).show();
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
