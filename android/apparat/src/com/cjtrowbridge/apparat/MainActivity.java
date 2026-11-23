/*
 * Hosts the Android Ebitengine view and bounded APK-update UI.
 * Uses Android lifecycle and file APIs; application data remains on the device.
 */
package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.content.Intent;
import android.content.pm.PackageInfo;
import android.content.pm.PackageManager;
import java.util.Arrays;
import java.util.concurrent.atomic.AtomicBoolean;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.provider.Settings;
import android.util.Log;
import android.view.Window;
import android.view.WindowManager;
import android.view.WindowInsets;
import android.view.WindowInsetsController;
import android.view.View;
import android.graphics.Insets;
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
    private static final String TAG = "ApparatUpdate";
    private static final String UPDATE_URL = "https://raw.githubusercontent.com/cjtrowbridge/apparat/main/releases/android/arm64/apparat/latest.apk";
    private static final String UPDATE_AUTHORITY = "com.cjtrowbridge.apparat.update";
    private static final String APK_MIME_TYPE = "application/vnd.android.package-archive";
    private EbitenView view;
    private ProjectDocuments documents;
    private DocumentEditor editor;
    private ActionVoice voice;
    LanBrowser lanBrowser;
    private NodeKeys nodeKeys;
    private File downloadedUpdate;
    private final AtomicBoolean checking = new AtomicBoolean();
    private volatile boolean destroyed;
    private volatile HttpURLConnection activeConnection;
    private Thread updateThread;

    private static class UpdateHttpException extends Exception {
        UpdateHttpException() {
            super();
        }
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Seq.setContext(getApplicationContext());
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(WindowManager.LayoutParams.FLAG_FULLSCREEN, WindowManager.LayoutParams.FLAG_FULLSCREEN);
        if (!Apparatmobile.initialize(getFilesDir().getAbsolutePath())) {
            Log.e(TAG, "Local runtime initialization failed");
            Toast.makeText(this, "Apparat could not initialize its local data. See app diagnostics.", Toast.LENGTH_LONG).show();
            finish();
            return;
        }
        view = new EbitenView(this);
        view.setFocusable(true);
        view.setFocusableInTouchMode(true);
        FrameLayout content = new FrameLayout(this);
        int background = (int) Apparatmobile.backgroundARGB();
        content.setBackgroundColor(background);
        getWindow().setNavigationBarColor(background);
        if (Build.VERSION.SDK_INT >= 29) {
            getWindow().setNavigationBarContrastEnforced(false);
        }
        content.addView(view, new FrameLayout.LayoutParams(-1, -1));
        if (Build.VERSION.SDK_INT >= 35) {
            // Inset the parent: EbitenView lays out its surface without padding.
            content.setOnApplyWindowInsetsListener(new View.OnApplyWindowInsetsListener() {
                @Override
                public WindowInsets onApplyWindowInsets(View target, WindowInsets insets) {
                    int types = WindowInsets.Type.systemBars() | WindowInsets.Type.displayCutout();
                    Insets safe = insets.getInsets(types);
                    target.setPadding(safe.left, safe.top, safe.right, safe.bottom);
                    return insets;
                }
            });
        }
        setContentView(content);
        nodeKeys = new NodeKeys(this);
        voice = new ActionVoice(this);
        Apparatmobile.registerActionSpeech(voice);
        editor = new DocumentEditor(this);
        Apparatmobile.registerEditor(editor);
        documents = new ProjectDocuments(this);
        Apparatmobile.registerDocuments(documents);
        if (Build.VERSION.SDK_INT >= 30) {
            getWindow().getInsetsController().setSystemBarsAppearance(0,
                WindowInsetsController.APPEARANCE_LIGHT_NAVIGATION_BARS);
        } else if (Build.VERSION.SDK_INT >= 26) {
            View decor = getWindow().getDecorView();
            decor.setSystemUiVisibility(decor.getSystemUiVisibility() & ~View.SYSTEM_UI_FLAG_LIGHT_NAVIGATION_BAR);
        }
        content.requestApplyInsets();
        if (Build.VERSION.SDK_INT >= 33) {
            getOnBackInvokedDispatcher().registerOnBackInvokedCallback(
                android.window.OnBackInvokedDispatcher.PRIORITY_DEFAULT,
                new android.window.OnBackInvokedCallback() {
                    @Override
                    public void onBackInvoked() {
                        MainActivity.this.onBackPressed();
                    }
                });
        }
        lanBrowser = new LanBrowser(this);
        Apparatmobile.registerUpdater(new AndroidUpdater(this));
        Log.i(TAG, "Updater bridge registered");
        // Downloads and installer handoff require an explicit Update action.
    }

    @Override
    protected void onActivityResult(int request, int result, Intent data) {
        super.onActivityResult(request,result,data);
        if ((request==NodeKeys.OPEN || request==NodeKeys.SAVE) && nodeKeys!=null) nodeKeys.documentResult(request,result==RESULT_OK && data!=null ? data.getData() : null);
        if (request == ProjectDocuments.REQUEST && documents != null) documents.result(result == RESULT_OK ? data : null);
    }

    @Override
    public void onBackPressed() {
        if (Apparatmobile.workspaceBackAvailable()) {
            Apparatmobile.workspaceBack();
            return;
        }
        if ("stmpck".equals(Apparatmobile.activeTab())) {
            Apparatmobile.arcadeBack();
            return;
        }
        finish();
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (nodeKeys!=null) nodeKeys.start();
        if (lanBrowser != null) lanBrowser.foreground(true);
        if (view != null) {
            view.resumeGame();
        }

    }

    @Override
    protected void onStop() {
        super.onStop();
        Apparatmobile.lockKeys();
    }

    @Override
    protected void onPause() {
        super.onPause();
        if (lanBrowser != null) lanBrowser.foreground(false);
        Apparatmobile.suspendInput();

        if (view != null) {
            view.suspendGame();
        }
    }



    @Override
    protected void onDestroy() {
        if (nodeKeys!=null) nodeKeys.close();
        destroyed = true;
        Apparatmobile.registerActionSpeech(null);
        if (voice != null) voice.close();
        Apparatmobile.registerEditor(null);
        if (editor != null) editor.close();
        HttpURLConnection connection = activeConnection;
        if (connection != null) {
            connection.disconnect();
        }
        if (updateThread != null) {
            updateThread.interrupt();
        }
        Apparatmobile.registerUpdater(null);
        if(documents!=null) documents.closePicker();
        super.onDestroy();
    }

    void checkForUpdate(final boolean userInitiated) {
        if (destroyed || !checking.compareAndSet(false, true)) {
            return;
        }
        Log.i(TAG, "Starting update check userInitiated=" + userInitiated);
        if (userInitiated) {
            reportUpdateStatus("checking");
            showToast("Checking Apparat update...");
        }
        updateThread = new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    File apk;
                    try {
                        apk = downloadLatestApk();
                    } catch (UpdateHttpException error) {
                        reportUpdateFailure("unable_http", userInitiated);
                        return;
                    } catch (UpdateTransfer.Failure error) {
                        Log.w(TAG, "event=download_failed reason=" + error.code + " bytes=" + error.bytes + " elapsed_ms=" + error.elapsed);
                        reportUpdateFailure(error.code, userInitiated);
                        return;
                    } catch (java.net.SocketTimeoutException error) {
                        reportUpdateFailure("unable_timeout", userInitiated);
                        return;
                    } catch (java.io.IOException error) {
                        Log.w(TAG, "event=download_failed error=" + error.getClass().getSimpleName());
                        reportUpdateFailure("unable_network", userInitiated);
                        return;
                    } catch (Exception error) {
                        reportUpdateFailure("unable_download", userInitiated);
                        return;
                    }
                    downloadedUpdate = apk;
                    String installedHash;
                    String downloadedHash;
                    try {
                        installedHash = ApkHash.sha256(new File(getApplicationInfo().sourceDir));
                        downloadedHash = ApkHash.sha256(apk);
                    } catch (Exception error) {
                        reportUpdateFailure("unable_verification", userInitiated);
                        apk.delete();
                        return;
                    }
                    if (installedHash.equals(downloadedHash)) {
                        if (userInitiated) {
                            reportUpdateStatus("up_to_date");
                            showToast("This download matches the installed APK.");
                        } else {
                            Log.i(TAG, "Startup update check found current APK " + ApkHash.shortHash(installedHash));
                        }
                        return;
                    }
                    String rejection = updateRejection(apk);
                    if (rejection != null) {
                        reportUpdateFailure(rejection, userInitiated);
                        apk.delete();
                        return;
                    }
                    runOnUiThread(new Runnable() {
                        @Override
                        public void run() {
                            if (!destroyed) {
                                requestPermissionOrInstall();
                            }
                        }
                    });
                } catch (Exception error) {
                    reportUpdateFailure("unable_download", userInitiated);
                } finally {
                    checking.set(false);
                }
            }
        }, "apparat-update");
        updateThread.start();
    }

    // TLS protects transport; installed signing identity and monotonic version
    // protect handoff. A public hash is equality evidence, not authenticity.
    @SuppressWarnings("deprecation")
    private String updateRejection(File apk) throws Exception {
        PackageManager manager = getPackageManager();
        PackageInfo current = manager.getPackageInfo(getPackageName(), PackageManager.GET_SIGNATURES);
        PackageInfo candidate = manager.getPackageArchiveInfo(apk.getPath(), PackageManager.GET_SIGNATURES);
        if (candidate == null || !getPackageName().equals(candidate.packageName)) {
            return "unable_package";
        }
        long oldVersion = Build.VERSION.SDK_INT >= 28 ? current.getLongVersionCode() : current.versionCode;
        long newVersion = Build.VERSION.SDK_INT >= 28 ? candidate.getLongVersionCode() : candidate.versionCode;
        if (current.signatures == null || candidate.signatures == null
                || !Arrays.equals(current.signatures, candidate.signatures)) {
            return "unable_signer";
        }
        if (newVersion <= oldVersion) {
            return "no_newer_update";
        }
        return null;
    }

    private File downloadLatestApk() throws Exception {
        File partial = new File(getCacheDir(), "latest.apk.part");
        File apk = new File(getCacheDir(), "latest.apk");
        HttpURLConnection connection = (HttpURLConnection) new URL(UPDATE_URL).openConnection();
        activeConnection = connection;
        boolean complete = false;
        try {
            connection.setUseCaches(false);
            connection.setConnectTimeout(15000);
            connection.setReadTimeout(UpdateTransfer.READ_TIMEOUT_MS);
            connection.setInstanceFollowRedirects(false);
            int status = connection.getResponseCode();
            if (status != 200) {
                throw new UpdateHttpException();
            }
            try (InputStream in = connection.getInputStream();
                 FileOutputStream out = new FileOutputStream(partial)) {
                UpdateTransfer.copy(in, out, new UpdateTransfer.State() {
                    public long now() {
                        return android.os.SystemClock.elapsedRealtime();
                    }
                    public boolean cancelled() {
                        return destroyed || Thread.currentThread().isInterrupted();
                    }
                });
                out.getFD().sync();
            }
            if (!partial.renameTo(apk)) {
                throw new java.io.IOException("cannot finalize download");
            }
            complete = true;
            return apk;
        } finally {
            connection.disconnect();
            activeConnection = null;
            if (!complete) {
                partial.delete();
            }
        }
    }

    private void requestPermissionOrInstall() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O && !getPackageManager().canRequestPackageInstalls()) {
            reportUpdateStatus("permission_needed");
            showToast("Allow Apparat to install updates, then tap Update again.");
            Intent intent = new Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES);
            intent.setData(Uri.parse("package:" + getPackageName()));
            startActivity(intent);
            return;
        }
        installDownloadedUpdate();
    }

    private void installDownloadedUpdate() {
        try {
            String rejection = downloadedUpdate == null ? "unable_verification" : updateRejection(downloadedUpdate);
            if (rejection != null) {
                reportUpdateStatus(rejection);
                return;
            }
        } catch (Exception error) {
            reportUpdateStatus("unable_verification");
            return;
        }
        if (downloadedUpdate == null || !downloadedUpdate.exists()) {
            reportUpdateStatus("unable_install");
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
            reportUpdateStatus("installer_opened");
        } catch (Exception error) {
            reportUpdateStatus("unable_install");
            showToast("Unable to open the system installer.");
        }
    }

    private void reportUpdateFailure(String code, boolean userInitiated) {
        Log.w(TAG, "Update check failed: " + code);
        if (userInitiated) {
            reportUpdateStatus(code);
            showToast("Unable to check for updates. Check status in Settings for details.");
        }
    }

    private void reportUpdateStatus(String message) {
        Log.i(TAG, "Update status: " + message);
        try {
            Apparatmobile.reportUpdateStatus(message);
        } catch (Exception error) {
            Log.w(TAG, "Could not report update status to HUD");
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

}
