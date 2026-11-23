/*
 * Supplies installed package metadata to lightweight Settings checks.
 * Starts the APK transfer only after the explicit download action.
 */
package com.cjtrowbridge.apparat;
import android.content.pm.PackageInfo;
import android.os.Build;
import java.util.Arrays;
import android.content.pm.PackageManager;
import com.cjtrowbridge.apparat.apparatmobile.Updater;

final class AndroidUpdater implements Updater {
    private final MainActivity activity;
    AndroidUpdater(MainActivity activity) {
        this.activity = activity;
    }
    public String resolveSaved(String host) {
        return activity.lanBrowser.resolveSaved(host);
    }
    public String resolveWorker(String network,String host) {
 return activity.lanBrowser.resolve(network,host);
 }
    public String browseLAN(boolean open) {
        return activity.lanBrowser.browse(open);
    }
    public String installedRelease() {
        try {
            PackageInfo info = activity.getPackageManager().getPackageInfo(activity.getPackageName(),0);
            return info.versionCode + " " + info.versionName;
        } catch (PackageManager.NameNotFoundException e) {
            return "";
        }
    }
    public void checkForUpdate() {
        activity.checkForUpdate(true);
    }
    // TLS protects transport; installed signing identity and monotonic version
    // protect handoff. A public hash is equality evidence, not authenticity.
    @SuppressWarnings("deprecation")
    static String updateRejection(MainActivity activity, java.io.File apk) throws Exception {
        PackageManager manager = activity.getPackageManager();
        PackageInfo current = manager.getPackageInfo(activity.getPackageName(), PackageManager.GET_SIGNATURES);
        PackageInfo candidate = manager.getPackageArchiveInfo(apk.getPath(), PackageManager.GET_SIGNATURES);
        if (candidate == null || !activity.getPackageName().equals(candidate.packageName)) {
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

}
