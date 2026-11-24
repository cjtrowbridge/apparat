/*
* Presents a full-window Android document with IME, selection handles and clipboard.
* Scoped callbacks persist drafts and check the base before writing a checkout.
*/
package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.app.AlertDialog;
import android.app.Dialog;
import android.content.DialogInterface;
import android.view.View;
import android.graphics.Typeface;
import android.text.Editable;
import android.text.InputType;
import android.text.TextWatcher;
import android.view.Gravity;
import android.view.WindowManager;
import android.widget.*;
import java.util.concurrent.CountDownLatch;
import com.cjtrowbridge.apparat.apparatmobile.Editor;

public final class DocumentEditor implements Editor {
    private final Activity activity;
    private Dialog active;
    DocumentEditor(Activity activity) {
        this.activity = activity;
    }
    @Override public void copyText(String value) {
        android.content.ClipboardManager clipboard = activity.getSystemService(android.content.ClipboardManager.class);
        clipboard.setPrimaryClip(android.content.ClipData.newPlainText("Apparat", value));
    }
    @Override public String pasteText() {
        android.content.ClipboardManager clipboard = activity.getSystemService(android.content.ClipboardManager.class);
        android.content.ClipData clip = clipboard.getPrimaryClip();
        if (clip == null || clip.getItemCount() == 0) return "";
        CharSequence text = clip.getItemAt(0).getText();
        return text == null ? "" : text.toString();
    }
    public void close() {
        if (active != null) active.dismiss();
    }
    @Override public String prompt(String title, String value, boolean singleLine) {
        CountDownLatch done = new CountDownLatch(1);
        String[] answer = new String[1];
        answer[0] = "null";
        activity.runOnUiThread(new Runnable() {
            public void run() {
                if (activity.isFinishing() || activity.isDestroyed()) {
                    done.countDown();
                    return;
                }
                EditText input = new EditText(activity);
                if (singleLine) {
                    input.setSingleLine(true);
                    input.setImeOptions(android.view.inputmethod.EditorInfo.IME_ACTION_DONE);
                }
                input.setText(value);
                input.setSelection(input.length());
                LinearLayout panel=new LinearLayout(activity);
                panel.setOrientation(LinearLayout.VERTICAL);
                FrameLayout workspace=new FrameLayout(activity);
                workspace.addView(input);
                Button ptt=VoiceInput.floatingButton(workspace);
                int space=ptt.getLayoutParams().height*2;
                workspace.setMinimumHeight(space+Math.round(12*activity.getResources().getDisplayMetrics().density));
                if (singleLine) {
                    FrameLayout.LayoutParams field = new FrameLayout.LayoutParams(-1,-2,Gravity.CENTER_VERTICAL);
                    field.rightMargin=space;
                    input.setLayoutParams(field);
                } else {
                    input.setMinHeight(space);
                    input.setPadding(input.getPaddingLeft(), input.getPaddingTop(),space,input.getPaddingBottom());
                }
                panel.addView(workspace);
                TextView preview=new TextView(activity);
                panel.addView(preview);
                AlertDialog dialog = new AlertDialog.Builder(activity).setTitle(title).setView(panel)
                .setNegativeButton("Cancel",null).setPositiveButton("Done",new DialogInterface.OnClickListener() {
                    public void onClick(DialogInterface d,int w) {
                        answer[0] = org.json.JSONObject.quote(input.getText().toString());
                    }
                }).create();
                VoiceInput voice=new VoiceInput(activity,dialog,input,ptt,preview);
                active = dialog;
                dialog.setOnDismissListener(new DialogInterface.OnDismissListener() {
                    public void onDismiss(DialogInterface d) {
                        voice.cancel();
                        active = null;
                        done.countDown();
                    }
                });
                dialog.create();
                dialog.getWindow().setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_STATE_ALWAYS_VISIBLE | WindowManager.LayoutParams.SOFT_INPUT_ADJUST_RESIZE);
                dialog.show();
                input.requestFocus();
            }
        });
        try {
            done.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
        return answer[0];
    }
}
