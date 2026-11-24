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
    private String expected;
    private int expectedStart;
    private int expectedEnd;
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
        bindDrag(dialog);
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
    private Button homeTarget;
    private Button lockTarget;
    private Button agentTarget;
    private boolean slideLocked;
    private boolean agentDrag;
    private float startX;
    private float startY;

    private Button target(String label, float x, float y) {
        FrameLayout frame = (FrameLayout)button.getParent();
        Button target = floatingButton(frame);
        GradientDrawable circle = new GradientDrawable();
        circle.setShape(GradientDrawable.OVAL);
        circle.setColor(0xff303d55);
        target.setBackground(circle);
        target.setTextColor(0xffffffff);
        target.setText(label);
        target.setTranslationX(x);
        target.setTranslationY(y);
        target.setClickable(false);
        target.setVisibility(View.INVISIBLE);
        return target;
    }

    private void bindDrag(Dialog dialog) {
        float travel = button.getLayoutParams().width * 0.75f;
        homeTarget = target("PTT", 0, 0);
        lockTarget = target("Lock", 0, -travel);
        agentTarget = target("Agent", -travel, 0);
        button.bringToFront();
        button.setOnTouchListener(new View.OnTouchListener() {
            @Override public boolean onTouch(View view, MotionEvent event) {
                float x = event.getRawX();
                float y = event.getRawY();
                switch (event.getActionMasked()) {
                    case MotionEvent.ACTION_DOWN:
                        startX = x;
                        startY = y;
                        view.getParent().requestDisallowInterceptTouchEvent(true);
                        if (!slideLocked) begin("press-pointer");
                        paint(true);
                        homeTarget.setVisibility(View.VISIBLE);
                        lockTarget.setVisibility(View.VISIBLE);
                        agentTarget.setVisibility(View.VISIBLE);
                        break;
                    case MotionEvent.ACTION_MOVE:
                        float dx = x - startX;
                        float dy = y - startY;
                        if (slideLocked && Math.hypot(dx, dy-travel) < travel/2) {
                            resetDrag();
                            stop("stop");
                        } else if (!slideLocked && !agentDrag) {
                            if (Math.hypot(dx+travel,dy) < travel/2) {
                                agentDrag = true;
                                button.setTranslationX(-travel);
                            } else if (listening && Math.hypot(dx,dy+travel) < travel/2) {
                                slideLocked = true;
                                startX = x;
                                startY = y;
                                button.setTranslationY(-travel);
                                receive(Apparatmobile.voice("slide-lock"));
                            }
                        }
                        break;
                    case MotionEvent.ACTION_UP:
                        view.getParent().requestDisallowInterceptTouchEvent(false);
                        if (agentDrag) {
                            String origin = Apparatmobile.activeTab();
                            restoreDraft();
                            listening = false;
                            handler.removeCallbacks(poll);
                            resetDrag();
                            show("");
                            dialog.hide();
                            Apparatmobile.voice("global-prompt");
                            restoreOnReturn(dialog, origin);
                        } else if (!slideLocked) {
                            resetDrag();
                            stop("release-pointer");
                        }
                        break;
                    case MotionEvent.ACTION_CANCEL:
                        view.getParent().requestDisallowInterceptTouchEvent(false);
                        cancel();
                        break;
                }
                return true;
            }
        });
    }

    // Keep the native editor and its unconfirmed text until the user returns.
    private void restoreOnReturn(Dialog dialog, String origin) {
        handler.postDelayed(new Runnable() {
            boolean left;
            public void run() {
                if (!field.isAttachedToWindow() || activity.isDestroyed()) return;
                String current = Apparatmobile.activeTab();
                if (!origin.equals(current)) left = true;
                if (left && origin.equals(current)) {
                    dialog.show();
                    field.requestFocus();
                } else {
                    handler.postDelayed(this, 200);
                }
            }
        }, 200);
    }

    private void resetDrag() {
        slideLocked = false;
        agentDrag = false;
        button.setTranslationX(0);
        button.setTranslationY(0);
        homeTarget.setVisibility(View.INVISIBLE);
        lockTarget.setVisibility(View.INVISIBLE);
        agentTarget.setVisibility(View.INVISIBLE);
        paint(false);
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
        expected=original;
        expectedStart=start;
        expectedEnd=end;
        listening=true;
        receive(Apparatmobile.voice(action));
    }
    private void stop(String action) {
        if (listening) receive(Apparatmobile.voice(action));
    }
    void cancel() {
        restoreDraft();
        resetDrag();
        if (listening) Apparatmobile.voice("cancel");
        listening=false;
        handler.removeCallbacks(poll);
        show("");
    }
    private void revise(String text) {
        int pos = Math.min(start,end);
        String next = original.substring(0,pos)+text+original.substring(Math.max(start,end));
        if (next.equals(expected)) return;
        field.getText().replace(pos,expected.length()-(original.length()-Math.max(start,end)),text);
        expected=next;
        expectedStart=pos+text.length();
        expectedEnd=expectedStart;
        field.setSelection(expectedStart);
    }
    private void restoreDraft() {
        if (expected!=null && !expected.equals(original) && expected.contentEquals(field.getText())) {
            field.setText(original);
            field.setSelection(start,end);
        }
        expected=null;
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
        if (!field.isAttachedToWindow() || !field.hasFocus() || !expected.contentEquals(field.getText())
                || expectedStart!=field.getSelectionStart() || expectedEnd!=field.getSelectionEnd()) {
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
                    label = slideLocked ? "Listening locked · slide down to finish" : "Listening locked · tap to finish";
                    paint(true);
                }
                if (!text.isEmpty()) revise(text);
                show(label);
                handler.postDelayed(poll,120);
                return;
            }
            listening=false;
            resetDrag();
            if (state.equals("final")) {
                if (!text.trim().isEmpty()) revise(text);
            } else {
                restoreDraft();
            }
            expected=null;
            show(state.equals("error") ? text : "");
        } catch (org.json.JSONException error) {
            cancel();
            show("Dictation session unavailable");
        }
    }
}
