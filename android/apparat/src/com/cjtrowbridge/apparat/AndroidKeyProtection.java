/*
 * Wraps a node credential with an Android Keystore AES key on supported Android 11+ devices.
 * BiometricPrompt authorizes each cipher operation with a strong biometric or device credential.
 */
package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.hardware.biometrics.BiometricManager;
import android.hardware.biometrics.BiometricPrompt;
import android.os.Build;
import android.os.CancellationSignal;
import android.security.keystore.KeyGenParameterSpec;
import android.security.keystore.KeyProperties;
import android.util.Base64;
import java.nio.charset.StandardCharsets;
import java.security.KeyStore;
import java.util.UUID;
import java.util.concurrent.ArrayBlockingQueue;
import javax.crypto.Cipher;
import javax.crypto.KeyGenerator;
import javax.crypto.spec.GCMParameterSpec;
import org.json.JSONObject;

final class AndroidKeyProtection {
    private final Activity activity;
    private CancellationSignal active;
    AndroidKeyProtection(Activity activity) {
        this.activity=activity;
    }
    void cancel() {
        if (active!=null) active.cancel();
    }
    private static final int ALLOWED=BiometricManager.Authenticators.BIOMETRIC_STRONG |
        BiometricManager.Authenticators.DEVICE_CREDENTIAL;

    void handle(String op,String value,ArrayBlockingQueue<String> answer) throws Exception {
        if (Build.VERSION.SDK_INT<30 ||
            activity.getSystemService(BiometricManager.class).canAuthenticate(ALLOWED)!=BiometricManager.BIOMETRIC_SUCCESS) {
            answer.offer("null");
            return;
        }
        if (op.equals("available")) {
            answer.offer(JSONObject.quote("Use Android device unlock"));
            return;
        }
        boolean create=op.equals("store");
        String[] parts=value.split(":",-1);
        String alias=create ? "apparat.node."+UUID.randomUUID() : parts[0];
        if (!alias.matches("apparat\\.node\\.[a-f0-9-]{36}")) throw new IllegalArgumentException();
        KeyStore store=KeyStore.getInstance("AndroidKeyStore");
        store.load(null);
        if (op.equals("remove")) {
            store.deleteEntry(alias);
            answer.offer(JSONObject.quote(""));
            return;
        }
        Cipher cipher=Cipher.getInstance("AES/GCM/NoPadding");
        if (create) {
            KeyGenerator generator=KeyGenerator.getInstance("AES","AndroidKeyStore");
            generator.init(new KeyGenParameterSpec.Builder(alias,KeyProperties.PURPOSE_ENCRYPT|KeyProperties.PURPOSE_DECRYPT)
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM).setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setUserAuthenticationRequired(true).setUserAuthenticationParameters(0,
                    KeyProperties.AUTH_BIOMETRIC_STRONG | KeyProperties.AUTH_DEVICE_CREDENTIAL).build());
            cipher.init(Cipher.ENCRYPT_MODE,generator.generateKey());
        } else {
            if (parts.length!=3) throw new IllegalArgumentException();
            cipher.init(Cipher.DECRYPT_MODE,store.getKey(alias,null),new GCMParameterSpec(128,Base64.decode(parts[1],Base64.NO_WRAP)));
        }
        active=new CancellationSignal();
        BiometricPrompt prompt=new BiometricPrompt.Builder(activity).setTitle("Unlock Apparat node key")
            .setAllowedAuthenticators(ALLOWED).build();
        prompt.authenticate(new BiometricPrompt.CryptoObject(cipher), active, activity.getMainExecutor(),
            new BiometricPrompt.AuthenticationCallback(){
                @Override public void onAuthenticationSucceeded(BiometricPrompt.AuthenticationResult result){
                    try {
                        Cipher authorized=result.getCryptoObject().getCipher();
                        byte[] data=authorized.doFinal(create ? value.getBytes(StandardCharsets.UTF_8) : Base64.decode(parts[2],Base64.NO_WRAP));
                        String response=create ? alias+":"+Base64.encodeToString(authorized.getIV(),Base64.NO_WRAP)+":"+
                            Base64.encodeToString(data,Base64.NO_WRAP) : new String(data,StandardCharsets.UTF_8);
                        answer.offer(JSONObject.quote(response));
                    }
                    catch(Exception ignored) {
                        answer.offer("null");
                    }
                }
                @Override public void onAuthenticationError(int code,CharSequence message){
                    if(create){
                        try {
                            store.deleteEntry(alias);
                        }
                        catch(Exception ignored) {
                        }
                    }
                    answer.offer("null");
                }
            });
    }
}
