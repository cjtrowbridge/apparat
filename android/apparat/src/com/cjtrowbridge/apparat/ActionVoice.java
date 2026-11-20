/*
 * Supplies offline spoken clarification and requests microphone access for canvas PTT.
 * The Activity owns the synthesizer and releases it when destroyed.
 */
package com.cjtrowbridge.apparat;

import android.Manifest;
import android.content.pm.PackageManager;
import android.speech.tts.TextToSpeech;
import android.speech.tts.Voice;
import com.cjtrowbridge.apparat.apparatmobile.ActionSpeech;

final class ActionVoice implements ActionSpeech, TextToSpeech.OnInitListener {
    private final MainActivity activity;
    private TextToSpeech synth;
    private volatile boolean ready;
    ActionVoice(MainActivity host) {
        activity = host;
        synth = new TextToSpeech(host, this);
    }
    public void onInit(final int status) {
        activity.runOnUiThread(new Runnable() {
            public void run() {
                if (status == TextToSpeech.SUCCESS && synth != null) {
                    Voice voice = synth.getVoice();
                    ready = voice != null && !voice.isNetworkConnectionRequired();
                }
            }
        });
    }
    public void say(final String prompt) {
        activity.runOnUiThread(new Runnable() {
            public void run() {
                if (ready && synth != null) {
                    if (prompt.isEmpty()) {
                        synth.stop();
                    } else {
                        synth.speak(prompt, TextToSpeech.QUEUE_FLUSH, null, "action-choice");
                    }
                }
            }
        });
    }
    public boolean microphone() {
        if (activity.checkSelfPermission(Manifest.permission.RECORD_AUDIO) == PackageManager.PERMISSION_GRANTED) {
            return true;
        }
        String[] permissions = new String[1];
        permissions[0] = Manifest.permission.RECORD_AUDIO;
        activity.runOnUiThread(new Runnable() {
            public void run() {
                activity.requestPermissions(permissions, 42);
            }
        });
        return false;
    }
    void close() {
        ready = false;
        synth.shutdown();
        synth = null;
    }
}
