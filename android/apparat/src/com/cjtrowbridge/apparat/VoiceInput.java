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
import android.widget.FrameLayout;
import android.view.Gravity;
import android.graphics.drawable.GradientDrawable;
import android.widget.EditText;
import android.widget.TextView;
import com.cjtrowbridge.apparat.apparatmobile.Apparatmobile;
import org.json.JSONObject;

final class VoiceInput {
    private final Activity activity;
    private final Button button;
    static Button floatingButton(FrameLayout frame) {
        Button button = new Button(frame.getContext());
        float density = frame.getResources().getDisplayMetrics().density;
        int size = Math.round(56 * density);
        int margin = Math.round(12 * density);
        FrameLayout.LayoutParams bounds = new FrameLayout.LayoutParams(size,size,Gravity.RIGHT | Gravity.BOTTOM);
        bounds.setMargins(margin,margin,margin,margin);
        button.setPadding(0,0,0,0);
        button.setElevation(4 * density);
        button.setFocusable(false);
        frame.addView(button,bounds);
        return button;
    }
    private void paint(boolean held) {
        GradientDrawable circle=new GradientDrawable();
        circle.setShape(GradientDrawable.OVAL);
        circle.setColor(held ? 0xff506fa4 : 0xff303d55);
        button.setBackground(circle);
        button.setTextColor(0xffffffff);
    }
    private final EditText field;
    private final TextView preview;
    private final Handler handler = new Handler(android.os.Looper.getMainLooper());
    private String original;
    private int start;
    private int end;
    private boolean listening;
    private String lastState="";
    private void trace(String event) {
        android.util.Log.i("ApparatVoice","component=voice event="+event);
    }

    VoiceInput(Activity activity, Dialog dialog, EditText field, Button button, TextView preview) {
        this.activity = activity;
        this.field = field;
        this.button=button;
        paint(false);
        this.preview = preview;
        button.setText("PTT");
        button.setContentDescription("Hold to dictate or double tap to lock listening");
        button.setOnTouchListener(new View.OnTouchListener() {
            public boolean onTouch(View v, MotionEvent event) {
            switch (event.getActionMasked()) {
                case MotionEvent.ACTION_DOWN:
                    v.getParent().requestDisallowInterceptTouchEvent(true);
                    paint(true);
                    trace("touch_down");
                    begin("press-pointer");
                    break;
                case MotionEvent.ACTION_UP:
                    paint(false);
                    v.getParent().requestDisallowInterceptTouchEvent(false);
                    trace("touch_up");
                    stop("release-pointer");
                    break;
                case MotionEvent.ACTION_CANCEL:
                    v.getParent().requestDisallowInterceptTouchEvent(false);
                    trace("touch_cancel");
                    cancel();
                    break;
            }
            return true;
            }
        });
        dialog.setOnKeyListener(new android.content.DialogInterface.OnKeyListener() {
            public boolean onKey(android.content.DialogInterface d, int key, KeyEvent event) {
            if (key==KeyEvent.KEYCODE_CTRL_RIGHT || key==KeyEvent.KEYCODE_BUTTON_L2) {
                boolean trigger = key==KeyEvent.KEYCODE_BUTTON_L2;
                if (event.getAction()==KeyEvent.ACTION_UP) stop(trigger ? "release-l2" : "stop");
                else if (event.getRepeatCount()==0) begin(trigger ? "press-l2" : "start");
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
    private void begin(String action) {
        if (listening) {
            if (action.startsWith("press-")) receive(Apparatmobile.voice(action));
            return;
        }
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
        receive(Apparatmobile.voice(action));
    }
    private void stop(String action) {
        if (listening) receive(Apparatmobile.voice(action));
    }
    void cancel() {
        paint(false);
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
            trace("editing_target_changed");
            cancel();
            return;
        }
        try {
            JSONObject r=new JSONObject(json);
            String state=r.getString("State");
            String text=r.getString("Text");
            if (!state.equals(lastState)) {
                trace(state);
                lastState=state;
            }
            if (state.equals("locked") || state.equals("recording") || state.equals("starting") || state.equals("finishing")) {
                String label = state.equals("finishing") ? "Finishing" : "Listening";
                if (state.equals("locked")) {
                    label = "Listening locked · tap to finish";
                    paint(true);
                }
                show(text.isEmpty() ? label : label+" · "+text);
                handler.postDelayed(poll,120);
                return;
            }
            listening=false;
            paint(false);
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
