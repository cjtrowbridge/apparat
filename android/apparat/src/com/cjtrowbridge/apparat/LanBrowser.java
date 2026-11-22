/*
 * Browses Apparat DNS-SD records on the active Android Wi-Fi network.
 * Results live only while the discovery detail is open and foregrounded.
 */
package com.cjtrowbridge.apparat;

import android.content.Context;
import android.net.ConnectivityManager;
import android.net.Network;
import android.net.NetworkCapabilities;
import android.net.nsd.NsdManager;
import android.net.nsd.NsdServiceInfo;
import android.os.Build;
import java.util.TreeSet;

final class LanBrowser {
    private final MainActivity activity;
    private final NsdManager nsd;
    private final ConnectivityManager connectivity;
    private final TreeSet<String> names = new TreeSet<>();
    private NsdManager.DiscoveryListener listener;
    private Network network;
    private boolean foreground;
    private String status = "Searching · no nearby hosts";

    LanBrowser(MainActivity owner) {
        activity = owner;
        nsd = (NsdManager) owner.getSystemService(Context.NSD_SERVICE);
        connectivity = (ConnectivityManager) owner.getSystemService(Context.CONNECTIVITY_SERVICE);
    }
    synchronized void foreground(boolean value) {
        foreground = value;
        if (!value) stop();
    }
    private void stop() {
        if (listener != null) {
            try {
                nsd.stopServiceDiscovery(listener);
            }
            catch (IllegalArgumentException ignored) {}
        }
        listener = null;
        network = null;
        names.clear();
    }
    synchronized String browse(boolean open) {
        if (!open || !foreground) {
            stop();
            return "";
        }
        if (Build.VERSION.SDK_INT < 33) return "Scoped discovery requires Android 13 or later";
        Network current = connectivity.getActiveNetwork();
        NetworkCapabilities caps = connectivity.getNetworkCapabilities(current);
        if (caps == null || !caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI)) {
            stop();
            return "Select a Wi-Fi network to browse";
        }
        if (!current.equals(network)) {
            stop();
            network = current;
            status = "Searching · no nearby hosts";
            listener = new NsdManager.DiscoveryListener() {
                public void onDiscoveryStarted(String type) {}
                public void onDiscoveryStopped(String type) {}
                public void onStartDiscoveryFailed(String type, int error) {
                    synchronized (LanBrowser.this) {
                        if (listener == this) status = "Network discovery unavailable";
                    }
                }
                public void onStopDiscoveryFailed(String type, int error) {}
                public void onServiceFound(NsdServiceInfo info) {
                    synchronized (LanBrowser.this) {
                        if (listener == this && names.size() < 128) names.add(info.getServiceName());
                    }
                }
                public void onServiceLost(NsdServiceInfo info) {
                    synchronized (LanBrowser.this) {
                        if (listener == this) names.remove(info.getServiceName());
                    }
                }
            };
            try {
                nsd.discoverServices("_apparat._tcp", NsdManager.PROTOCOL_DNS_SD,
                    current, activity.getMainExecutor(), listener);
            } catch (RuntimeException error) {
                status = "Network discovery unavailable";
            }
        }
        return names.isEmpty() ? status : new org.json.JSONArray(names).toString();
    }
}
