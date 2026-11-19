/*
 * Supplies installed package metadata to lightweight Settings checks.
 * Starts the APK transfer only after the explicit download action.
 */
package com.cjtrowbridge.apparat;
import android.content.pm.PackageInfo;
import android.content.pm.PackageManager;
import com.cjtrowbridge.apparat.apparatmobile.Updater;

final class AndroidUpdater implements Updater {
    private final MainActivity activity;
    AndroidUpdater(MainActivity activity) {
        this.activity = activity;
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
}
