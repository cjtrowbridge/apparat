/*
 * Binds one dialog's hold control to its editable selection.
 * A preview stays outside the document until the matching final result arrives.
 */
package com.cjtrowbridge.apparat;

import android.Manifest;
import android.app.Activity;
import android.app.Dialog;
import android.content.pm.PackageManager;
import android.os.Handler;
import android.view.KeyEvent;
import android.view.MotionEvent;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import com.cjtrowbridge.apparat.apparatmobile.Apparatmobile;
import org.json.JSONObject;

final class VoiceInput {
    private final Activity activity;
    private final EditText field;
    private final TextView preview;
    private final Handler handler = new Handler(android.os.Looper.getMainLooper());
    private String original;
    private int start;
    private int end;
    private boolean listening;

    VoiceInput(Activity activity, Dialog dialog, EditText field, Button button, TextView preview) {
        this.activity = activity;
        this.field = field;
        this.preview = preview;
        button.setText("PTT");
        button.setContentDescription("Hold to dictate");
        button.setOnTouchListener(new View.OnTouchListener() {
            public boolean onTouch(View v, MotionEvent event) {
            switch (event.getActionMasked()) {
                case MotionEvent.ACTION_DOWN:
                    begin();
                    break;
                case MotionEvent.ACTION_UP:
                    stop();
                    break;
                case MotionEvent.ACTION_CANCEL:
                    cancel();
                    break;
            }
            return true;
            }
        });
        dialog.setOnKeyListener(new android.content.DialogInterface.OnKeyListener() {
            public boolean onKey(android.content.DialogInterface d, int key, KeyEvent event) {
            if (key==KeyEvent.KEYCODE_CTRL_RIGHT || key==KeyEvent.KEYCODE_BUTTON_R2) {
                if (event.getAction()==KeyEvent.ACTION_UP) stop();
                else if (event.getRepeatCount()==0) begin();
                return true;
            }
            if (listening && (key==KeyEvent.KEYCODE_ESCAPE || key==KeyEvent.KEYCODE_BACK)) {
                cancel();
                return true;
            }
            return false;
            }
        });
    }
    private void begin() {
        if (listening) return;
        if (activity.checkSelfPermission(Manifest.permission.RECORD_AUDIO)!=PackageManager.PERMISSION_GRANTED) {
            activity.requestPermissions(new String[] {
                Manifest.permission.RECORD_AUDIO
            }, 219);
            show("Allow microphone access, then hold PTT again.");
            return;
        }
        field.requestFocus();
        start=field.getSelectionStart();
        end=field.getSelectionEnd();
        if (start<0 || end<0) return;
        original=field.getText().toString();
        listening=true;
        receive(Apparatmobile.voice("start"));
    }
    private void stop() {
        if (listening) receive(Apparatmobile.voice("stop"));
    }
    void cancel() {
        if (listening) Apparatmobile.voice("cancel");
        listening=false;
        handler.removeCallbacks(poll);
        show("");
    }
    private void show(String value) {
        preview.setText(value);
        preview.setVisibility(value.isEmpty() ? View.GONE : View.VISIBLE);
    }
    private final Runnable poll=new Runnable() {
        public void run() {
            if (listening) receive(Apparatmobile.voice("poll"));
        }
    };
    private void receive(String json) {
        handler.removeCallbacks(poll);
        if (!field.isAttachedToWindow() || !field.hasFocus() || !original.contentEquals(field.getText())
                || start!=field.getSelectionStart() || end!=field.getSelectionEnd()) {
            cancel();
            return;
        }
        try {
            JSONObject r=new JSONObject(json);
            String state=r.getString("State");
            String text=r.getString("Text");
            if (state.equals("recording") || state.equals("starting") || state.equals("finishing")) {
                show(text.isEmpty() ? state : text);
                handler.postDelayed(poll,120);
                return;
            }
            listening=false;
            if (state.equals("final") && !text.isEmpty()) {
                field.getText().replace(Math.min(start,end),Math.max(start,end),text);
            }
            show(state.equals("error") ? text : "");
        } catch (org.json.JSONException error) {
            cancel();
            show("Dictation session unavailable");
        }
    }
}
