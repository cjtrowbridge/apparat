/*
 * Supplies private local dialogs and Android document access to the node identity workflow.
 * The Go setup thread waits for one answer while the Activity thread remains responsive.
 */
package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.app.AlertDialog;
import android.content.Intent;
import android.content.DialogInterface;
import android.net.Uri;
import android.text.InputType;
import android.util.Base64;
import android.view.WindowManager;
import android.widget.EditText;
import com.cjtrowbridge.apparat.apparatmobile.Apparatmobile;
import com.cjtrowbridge.apparat.apparatmobile.KeyDialogs;
import org.json.JSONArray;
import org.json.JSONObject;
import java.io.ByteArrayOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.concurrent.ArrayBlockingQueue;

final class NodeKeys implements KeyDialogs {
    static final int OPEN = 4871;
    static final int SAVE = 4872;
    private final Activity activity;
    private final AndroidKeyProtection protection;
    private ArrayBlockingQueue<String> pending;
    private byte[] export;
    private boolean closed;
    private boolean running;
    private AlertDialog visible;

    NodeKeys(Activity activity) {
        this.activity = activity;
        this.protection = new AndroidKeyProtection(activity);
    }

    synchronized void start() {
        if (running || closed) return;
        running = true;
        new Thread(new Runnable() {
            @Override public void run() {
                try {
                    Apparatmobile.setupKeys(NodeKeys.this);
                }
                finally {
                    synchronized (NodeKeys.this) {
                        running = false;
                    }
                }
            }
        }, "Apparat-key-setup").start();
    }

    @Override public String ask(String request) {
        final ArrayBlockingQueue<String> answer = new ArrayBlockingQueue<>(1);
        synchronized (this) {
            if (closed) return "null";
            pending = answer;
        }
        activity.runOnUiThread(new Runnable() {
            @Override public void run() {
                synchronized (NodeKeys.this) {
                    if (closed || pending != answer) return;
                }
                try {
                    show(new JSONObject(request), answer);
                }
                catch (Exception ignored) {
                    answer.offer("null");
                }
            }
        });
        try {
            return answer.take();
        }
        catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return "null";
        }
        finally {
            synchronized (this) {
                if (pending == answer) pending = null;
            }
        }
    }

    void close() {
        synchronized (this) {
            closed = true;
            if (pending != null) pending.offer("null");
        }
        if (visible != null) visible.dismiss();
        protection.cancel();
    }

    private static DialogInterface.OnClickListener respond(ArrayBlockingQueue<String> queue,String value) {
        return new DialogInterface.OnClickListener(){
            @Override public void onClick(DialogInterface dialog,int which) {
                queue.offer(value);
            }
        };
    }

    private void show(JSONObject r, ArrayBlockingQueue<String> answer) throws Exception {
        final String op = r.getString("Op");
        String title = r.getString("Title");
        String message = r.getString("Message");
        if (op.equals("available") || op.equals("store") || op.equals("load") || op.equals("remove")) {
            protection.handle(op,r.getString("Value"),answer);
            return;
        }
        if (op.equals("file") || op.equals("save")) {
            boolean save = op.equals("save");
            Intent intent = new Intent(save ? Intent.ACTION_CREATE_DOCUMENT : Intent.ACTION_OPEN_DOCUMENT);
            intent.addCategory(Intent.CATEGORY_OPENABLE);
            intent.setType(save ? "application/json" : "*/*");
            if (save) {
                export = Base64.decode(r.getString("Value"), Base64.DEFAULT);
                intent.putExtra(Intent.EXTRA_TITLE,"apparat-recovery.json");
            }
            activity.startActivityForResult(intent,save ? SAVE : OPEN);
            return;
        }
        AlertDialog.Builder builder = new AlertDialog.Builder(activity).setTitle(title).setMessage(message)
            .setNegativeButton("Cancel",respond(answer,"null"))
            .setOnCancelListener(new DialogInterface.OnCancelListener(){
                @Override public void onCancel(DialogInterface dialog){
                    answer.offer("null");
                }
            });
        if (op.equals("password")) {
            EditText input = new EditText(activity);
            input.setSingleLine(true);
            input.setInputType(InputType.TYPE_CLASS_TEXT | InputType.TYPE_TEXT_VARIATION_PASSWORD |
                InputType.TYPE_TEXT_FLAG_NO_SUGGESTIONS);
            if (android.os.Build.VERSION.SDK_INT >= 26) input.setImportantForAutofill(android.view.View.IMPORTANT_FOR_AUTOFILL_NO);
            builder.setView(input).setPositiveButton("Continue",new DialogInterface.OnClickListener(){
                @Override public void onClick(DialogInterface dialog,int which){
                    String response = JSONObject.quote(input.getText().toString());
                    input.getText().clear();
                    answer.offer(response);
                }
            });
        } else if (op.equals("choose")) {
            JSONArray items = r.getJSONArray("Options");
            String first = items.getString(0);
            builder.setPositiveButton(first,respond(answer,JSONObject.quote(first)));
            if (items.length()>1) {
                String other=items.getString(1);
                builder.setNeutralButton(other,respond(answer,JSONObject.quote(other)));
            }
        }
        visible = builder.create();
        if (op.equals("password")) visible.getWindow().addFlags(WindowManager.LayoutParams.FLAG_SECURE);
        visible.show();
    }

    void documentResult(int code, Uri uri) {
        final ArrayBlockingQueue<String> answer;
        synchronized (this) {
            answer = pending;
        }
        if (answer == null) return;
        final byte[] output=export;
        export=null;
        new Thread(new Runnable(){
            @Override public void run(){
                try {
                    if (uri == null) {
                        answer.offer("null");
                        return;
                    }
                    if (code == SAVE) {
                        try (OutputStream out=activity.getContentResolver().openOutputStream(uri,"wt")) {
                            out.write(output);
                        }
                        answer.offer(JSONObject.quote(""));
                    } else {
                        try (InputStream in=activity.getContentResolver().openInputStream(uri)) {
                            ByteArrayOutputStream data=new ByteArrayOutputStream();
                            byte[] part=new byte[2048];
                            int n;
                            n=in.read(part);
                            while (n!=-1) {
                                if (data.size()+n>16384) throw new java.io.IOException();
                                data.write(part,0,n);
                                n=in.read(part);
                            }
                            answer.offer(JSONObject.quote(Base64.encodeToString(data.toByteArray(),Base64.NO_WRAP)));
                        }
                    }
                }
                catch (Exception ignored) {
                    answer.offer("null");
                }
            }
        },"Apparat-key-document").start();
    }
}
