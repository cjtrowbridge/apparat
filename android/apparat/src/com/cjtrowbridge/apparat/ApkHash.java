/*
 * Computes local APK digests for update verification.
 * Reads the package as a bounded stream.
 */
package com.cjtrowbridge.apparat;
import java.io.*;
import java.util.Locale;
import java.security.MessageDigest;
final class ApkHash {
    static String shortHash(String hash) {
        if (hash == null || hash.length() < 12) {
            return hash;
        }
        return hash.substring(0, 12);
    }

    static String sha256(File file) throws Exception {
        MessageDigest digest = MessageDigest.getInstance("SHA-256");
        try (InputStream in = new FileInputStream(file)) {
            byte[] buffer = new byte[64 * 1024];
            int read;
            while (true) {
                read = in.read(buffer);
                if (read == -1) {
                    break;
                }
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
