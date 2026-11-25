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
import java.util.TreeMap;

final class LanBrowser {
    private final MainActivity activity;
    private final NsdManager nsd;
    private final ConnectivityManager connectivity;
    private final TreeMap<String,NsdServiceInfo> names = new TreeMap<>();
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
    String resolveSaved(String host) {
        if (Build.VERSION.SDK_INT < 33) return "";
        Network active = connectivity.getActiveNetwork();
        NetworkCapabilities caps = connectivity.getNetworkCapabilities(active);
        if (caps == null || !caps.hasTransport(NetworkCapabilities.TRANSPORT_WIFI)) return "";
        synchronized (this) {
            if (!foreground) return "";
        }
        NsdServiceInfo saved = new NsdServiceInfo();
        saved.setServiceName(host);
        saved.setServiceType("_apparat._tcp");
        saved.setNetwork(active);
        return resolveEndpoint(active, saved, true);
    }
    String resolve(String device, String name) {
        final Network expected;
        final NsdServiceInfo selected;
        synchronized (this) {
            expected = network;
            selected = names.get(name);
            android.net.LinkProperties link = connectivity.getLinkProperties(expected);
            if (!foreground || expected == null || selected == null || link == null ||
                !device.equals(link.getInterfaceName()) || !expected.equals(connectivity.getActiveNetwork())) return "";
        }
        return resolveEndpoint(expected, selected, false);
    }
    private String resolveEndpoint(Network expected, NsdServiceInfo selected, boolean saved) {
        String name = selected.getServiceName();
        java.util.concurrent.ArrayBlockingQueue<String> answer = new java.util.concurrent.ArrayBlockingQueue<>(1);
        try {
            nsd.resolveService(selected, new NsdManager.ResolveListener() {
                public void onResolveFailed(NsdServiceInfo ignored, int code) {
                    answer.offer("");
                }
                public void onServiceResolved(NsdServiceInfo result) {
                    String address = "";
                    synchronized (LanBrowser.this) {
                        if (foreground && expected.equals(connectivity.getActiveNetwork()) &&
                            (saved || (expected.equals(network) && names.containsKey(name))) && name.equals(result.getServiceName()) &&
                            result.getHost() != null && result.getPort() > 0) {
                            String host = result.getHost().getHostAddress();
                            address = (host.contains(":") ? "[" + host + "]" : host) + ":" + result.getPort();
                        }
                    }
                    answer.offer(address);
                }
            });
            String address = answer.poll(5, java.util.concurrent.TimeUnit.SECONDS);
            return address == null ? "" : address;
        } catch (InterruptedException error) {
            Thread.currentThread().interrupt();
            return "";
        } catch (RuntimeException error) {
            return "";
        }
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
                        if (listener == this && names.size() < 128) names.put(info.getServiceName(), info);
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
        if (names.isEmpty()) return status;
        android.net.LinkProperties link = connectivity.getLinkProperties(current);
        if (link == null || link.getInterfaceName() == null) return "Wi-Fi interface unavailable";
        try {
            return new org.json.JSONObject().put("network",link.getInterfaceName()).put("hosts",new org.json.JSONArray(names.keySet())).toString();
        } catch (org.json.JSONException e) {
            return "Network discovery unavailable";
        }
    }
}
