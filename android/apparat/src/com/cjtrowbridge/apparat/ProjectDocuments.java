/*
* Resolves a user-granted Android document tree and reads bounded Git metadata.
* Copy-out is restricted to a freshly created directory owned by this clone.
*/
package com.cjtrowbridge.apparat;
import android.app.Activity;
import android.content.ContentResolver;
import android.content.Intent;
import android.database.Cursor;
import android.net.Uri;
import android.provider.DocumentsContract;
import android.provider.DocumentsContract.Document;
import com.cjtrowbridge.apparat.apparatmobile.Documents;
import com.cjtrowbridge.apparat.apparatmobile.Transfer;
import java.io.*;
import java.util.concurrent.ArrayBlockingQueue;
import org.json.JSONObject;
final class ProjectDocuments implements Documents {
    static final int REQUEST = 7261;
    private final Activity activity;
    private final ContentResolver resolver;
    private final ArrayBlockingQueue<String> selection = new ArrayBlockingQueue<>(1);
    ProjectDocuments(Activity activity) {
        this.activity = activity;
        resolver = activity.getContentResolver();
    }
    public String chooseFolder() {
        selection.clear();
        activity.runOnUiThread(new Runnable() {
            @Override
            public void run() {
                try {
                Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT_TREE);
                intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION | Intent.FLAG_GRANT_WRITE_URI_PERMISSION |
                Intent.FLAG_GRANT_PERSISTABLE_URI_PERMISSION | Intent.FLAG_GRANT_PREFIX_URI_PERMISSION);
                activity.startActivityForResult(intent, REQUEST);
            } catch (RuntimeException ex) {
                closePicker();
            }
            }
        });
        try {
            return selection.take();
        }
        catch (InterruptedException ex) {
            Thread.currentThread().interrupt();
            return "";
        }
    }
    void closePicker() {
        selection.offer("");
    }
    void result(Intent data) {
        if (data == null || data.getData() == null) {
            closePicker();
            return;
        }
        Uri tree = data.getData();
        try {
            int flags = data.getFlags() & (Intent.FLAG_GRANT_READ_URI_PERMISSION | Intent.FLAG_GRANT_WRITE_URI_PERMISSION);
            resolver.takePersistableUriPermission(tree, flags);
            selection.offer(DocumentsContract.buildDocumentUriUsingTree(tree,
            DocumentsContract.getTreeDocumentId(tree)).toString());
        } catch (RuntimeException ex) {
            closePicker();
        }
    }
    private Uri child(Uri dir, String name) throws IOException {
        Uri children = DocumentsContract.buildChildDocumentsUriUsingTree(dir, DocumentsContract.getDocumentId(dir));
        String[] columns = {
            Document.COLUMN_DISPLAY_NAME, Document.COLUMN_DOCUMENT_ID
        };
        try (Cursor c = resolver.query(children, columns,null,null,null)) {
            if (c == null) throw new IOException();
            while (c.moveToNext()) if (name.equals(c.getString(0))) {
                return DocumentsContract.buildDocumentUriUsingTree(dir,c.getString(1));
            }
        }
        return null;
    }
    private String name(Uri document) throws IOException {
        String[] columns = {
            Document.COLUMN_DISPLAY_NAME
        };
        try (Cursor c = resolver.query(document,columns,null,null,null)) {
            if (c == null || !c.moveToFirst()) throw new IOException();
            return c.getString(0);
        }
    }
    private String read(Uri document, int limit) throws IOException {
        ByteArrayOutputStream out=new ByteArrayOutputStream();
        try(InputStream in=resolver.openInputStream(document)) {
            if (in==null) throw new IOException();
            byte[] buffer=new byte[4096];
            int n;
            n=in.read(buffer);
            while (n!=-1) {
                if (out.size()+n>limit) throw new IOException();
                out.write(buffer,0,n);
                n=in.read(buffer);
            }
            return out.toString("UTF-8");
        }
    }
    public String inspect(String location) {
        try {
            Uri root=Uri.parse(location);
            Uri git=child(root,".git");
            if (git==null || child(git,"objects")==null || child(git,"refs")==null) return "";
            Uri head=child(git,"HEAD");
            if(head==null) return "";
            String branch=read(head,4096).trim();
            if(branch.startsWith("ref: refs/heads/")) branch=branch.substring(16);
            else if(!branch.matches("[0-9a-fA-F]{40}|[0-9a-fA-F]{64}")) return "";
            Uri vscode=child(root,".vscode");
            Uri tasks=vscode==null?null:child(vscode,"tasks.json");
            String config="";
            if(tasks!=null) {
                try{
                    config=read(tasks,1<<20);
                } catch(IOException ignored) {
                    // A task file is optional.
                }
            }
            return new JSONObject().put("Name",name(root)).put("Head",branch).put("Tasks",config).toString();
        } catch(Exception ex) {
            return "";
        }
    }
    private Uri create(Uri parent,String mime,String name) throws IOException {
        Uri made=DocumentsContract.createDocument(resolver,parent,mime,name);
        if(made==null) throw new IOException();
        if(!name(made).equals(name)) {
            DocumentsContract.deleteDocument(resolver,made);
            throw new IOException();
        }
        return made;
    }
    private void copy(File source,Uri dest, Transfer transfer) throws IOException {
        File[] files=source.listFiles();
        if(files==null) throw new IOException();
        for(File file:files) {
            if (transfer.cancelled()) throw new InterruptedIOException();
            if(!file.getCanonicalFile().equals(file.getAbsoluteFile())) throw new IOException();
            if(file.isDirectory()) copy(file,create(dest,Document.MIME_TYPE_DIR,file.getName()),transfer);
            else {
                Uri doc=create(dest,"application/octet-stream",file.getName());
                try(InputStream in=new FileInputStream(file);
                OutputStream out=resolver.openOutputStream(doc)) {
                    if(out==null) throw new IOException();
                    byte[] buf=new byte[65536];
                    int n;
                    n=in.read(buf);
                    while(n!=-1){
                        if (transfer.cancelled()) throw new InterruptedIOException();
                        out.write(buf,0,n);
                        n=in.read(buf);
                    }
                }
            }
        }
    }
    public String export(String location,String repo,String source,Transfer transfer) {
        Uri created=null;
        try {
            Uri parent=Uri.parse(location);
            if(child(parent,repo)!=null) return "";
            created=create(parent,Document.MIME_TYPE_DIR,repo);
            copy(new File(source),created,transfer);
            return created.toString();
        } catch(Exception ex) {
            if(created!=null) try{
                DocumentsContract.deleteDocument(resolver,created);
            }catch(Exception ignored){}
            return "";
        }
    }
}
