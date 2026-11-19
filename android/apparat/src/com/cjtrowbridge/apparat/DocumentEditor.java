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
import com.cjtrowbridge.apparat.apparatmobile.EditorSession;

public final class DocumentEditor implements Editor {
    private final Activity activity;
    private Dialog active;
    private final java.util.Map<String,int[]> positions = new java.util.HashMap<>();
    DocumentEditor(Activity activity) {
        this.activity = activity;
    }
    public void close() {
        if (active != null) active.dismiss();
    }
    @Override public String open(String project, String title, String text, EditorSession session) {
        CountDownLatch done = new CountDownLatch(1);
        activity.runOnUiThread(new Runnable() {
            public void run() {
                if (activity.isFinishing() || activity.isDestroyed()) {
                    done.countDown();
                    return;
                }
                Dialog dialog = new Dialog(activity, android.R.style.Theme_Material_NoActionBar);
                active = dialog;
                int inset = (int)(4 * activity.getResources().getDisplayMetrics().density);
                LinearLayout column = new LinearLayout(dialog.getContext());
                column.setBackgroundColor(0xff0f121c);
                column.setPadding(0,0,0,0);
                column.setOrientation(LinearLayout.VERTICAL);
                LinearLayout bar = new LinearLayout(dialog.getContext());
                Button back = new Button(dialog.getContext());
                back.setText("Back");
                back.setOnClickListener(new View.OnClickListener() {
                    public void onClick(View v) {
                        dialog.dismiss();
                    }
                });
                bar.addView(back);
                TextView name = new TextView(dialog.getContext());
                name.setTextColor(0xffe6e8ed);
                name.setText(title);
                name.setSingleLine(true);
                name.setEllipsize(android.text.TextUtils.TruncateAt.START);
                bar.addView(name, new LinearLayout.LayoutParams(0,-2,1));
                Button undo = new Button(dialog.getContext());
                undo.setText("Undo");
                bar.addView(undo);
                Button save = new Button(dialog.getContext());
                save.setText("Save");
                bar.addView(save);
                column.addView(bar);
                TextView status = new TextView(dialog.getContext());
                status.setTextColor(0xffb7c5d9);
                status.setVisibility(View.GONE);
                column.addView(status);
                EditText body = new EditText(dialog.getContext());
                body.setTextColor(0xfff0f2f6);
                body.setTextSize(16);
                body.setPadding(inset,inset,inset,inset);
                body.setGravity(Gravity.TOP | Gravity.START);
                body.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_FLAG_MULTI_LINE | InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS);
                body.setTypeface(Typeface.MONOSPACE);
                body.setText(text);
                body.setContentDescription("Document text");
                undo.setOnClickListener(new View.OnClickListener() {
                    public void onClick(View v) {
                        body.onTextContextMenuItem(android.R.id.undo);
                    }
                });
                FrameLayout workspace=new FrameLayout(dialog.getContext());
                workspace.addView(body);
                Button ptt=VoiceInput.floatingButton(workspace);
                body.setPadding(inset,inset,inset,ptt.getLayoutParams().height+6*inset);
                VoiceInput voice=new VoiceInput(activity,dialog,body,ptt,status);
                int[] position = positions.get(project + "/" + title);
                if (position != null) {
                    body.setSelection(Math.min(body.length(),position[0]),Math.min(body.length(),position[1]));
                    body.post(new Runnable() {
                        public void run() {
                            body.scrollTo(0,position[2]);
                        }
                    });
                }
                column.addView(workspace, new LinearLayout.LayoutParams(-1,0,1));
                body.addTextChangedListener(new TextWatcher() {
                    public void beforeTextChanged(CharSequence s,int start,int count,int after) {}
                    public void onTextChanged(CharSequence s,int start,int before,int count) {}
                    public void afterTextChanged(Editable value) {
                        String error = session.apply("draft",value.toString());
                        name.setText(error.isEmpty() ? title + " •" : title);
                        status.setVisibility(error.isEmpty() ? View.GONE : View.VISIBLE);
                        status.setText(error);
                    }
                });
                save.setOnClickListener(new View.OnClickListener() {
                    public void onClick(View v) {
                        String draft = body.getText().toString();
                        String reason = session.apply("save", draft);
                        if (reason.isEmpty()) {
                            dialog.dismiss();
                            return;
                        }
                        status.setVisibility(View.VISIBLE);
                        status.setText(reason);
                        new AlertDialog.Builder(activity).setTitle("Save needs attention").setMessage(reason)
                        .setNegativeButton("Back",null).setPositiveButton("Review checkout",new DialogInterface.OnClickListener() {
                            public void onClick(DialogInterface a,int b) {
                                String current = session.apply("compare",draft);
                                new AlertDialog.Builder(activity).setTitle("Current checkout").setMessage(current)
                                .setNegativeButton("Back",null).setPositiveButton("Keep my draft against this version",new DialogInterface.OnClickListener() {
                                    public void onClick(DialogInterface x,int y) {
                                        String error = session.apply("rebase",draft);
                                        status.setText(error.isEmpty() ? "Draft retained. Review before saving." : error);
                                    }
                                }).show();
                            }
                        }).show();
                    }
                });
                dialog.setContentView(column);
                dialog.setOnDismissListener(new DialogInterface.OnDismissListener() {
                    public void onDismiss(DialogInterface d) {
                        int[] bookmark = new int[3];
                        bookmark[0] = Math.max(0,body.getSelectionStart());
                        bookmark[1] = Math.max(0,body.getSelectionEnd());
                        bookmark[2] = body.getScrollY();
                        positions.put(project + "/" + title,bookmark);
                        voice.cancel();
                        active = null;
                        done.countDown();
                    }
                });
                dialog.getWindow().setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_ADJUST_RESIZE);
                dialog.show();
            }
        });
        try {
            done.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return "Editor interrupted";
        }
        return "";
    }

    @Override public String prompt(String title, String value) {
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
                input.setText(value);
                input.setSelection(input.length());
                LinearLayout panel=new LinearLayout(activity);
                panel.setOrientation(LinearLayout.VERTICAL);
                FrameLayout workspace=new FrameLayout(activity);
                workspace.addView(input);
                Button ptt=VoiceInput.floatingButton(workspace);
                int space=ptt.getLayoutParams().height*3/2;
                input.setMinHeight(space);
                input.setPadding(input.getPaddingLeft(), input.getPaddingTop(),space,input.getPaddingBottom());
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
