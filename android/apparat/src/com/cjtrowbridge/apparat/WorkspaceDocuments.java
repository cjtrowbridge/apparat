/*
 * Maps repository-relative names to documents beneath one persisted tree grant.
 * Provides seekable file descriptors and document operations for the local Git adapter.
 */
package com.cjtrowbridge.apparat;
import android.content.ContentResolver;
import android.content.Context;
import android.content.SharedPreferences;
import java.util.UUID;
import android.database.Cursor;
import android.net.Uri;
import android.os.ParcelFileDescriptor;
import android.provider.DocumentsContract;
import android.provider.DocumentsContract.Document;
import java.io.*;
import org.json.JSONArray;
import org.json.JSONObject;

final class WorkspaceDocuments {
    private final ContentResolver resolver;
    private final Uri root;
    private final SharedPreferences recovery;
    WorkspaceDocuments(Context context, String root) {
        this.resolver = context.getContentResolver();
        recovery=context.getSharedPreferences("workspace-document-replacement",Context.MODE_PRIVATE);
        this.root = Uri.parse(root);
    }
    private static final String[] COLUMNS = {
        Document.COLUMN_DOCUMENT_ID, Document.COLUMN_DISPLAY_NAME, Document.COLUMN_MIME_TYPE,
        Document.COLUMN_SIZE, Document.COLUMN_LAST_MODIFIED
    };
    private Uri children(Uri dir) {
        return DocumentsContract.buildChildDocumentsUriUsingTree(dir, DocumentsContract.getDocumentId(dir));
    }
    private static String parent(String path) {
        int split = path.lastIndexOf('/');
        return split < 0 ? "." : path.substring(0,split);
    }
    private static String base(String path) {
        return path.substring(path.lastIndexOf('/')+1);
    }
    private static void validate(String part) throws IOException {
        if(part.isEmpty() || part.equals(".") || part.equals("..") || part.contains("/") || part.contains("\\")) throw new IOException();
    }
    private Uri child(Uri dir,String name) throws IOException {
        validate(name);
        try(Cursor rows=resolver.query(children(dir),COLUMNS,null,null,null)) {
            if(rows==null) throw new IOException();
            Uri match=null;
            while(rows.moveToNext()) {
                if(!name.equals(rows.getString(1))) continue;
                if(match!=null) throw new IOException();
                match=DocumentsContract.buildDocumentUriUsingTree(root,rows.getString(0));
            }
            return match;
        }
    }
    private Uri create(Uri dir,String name,boolean directory) throws Exception {
        validate(name);
        if(child(dir,name)!=null) throw new IOException("exists");
        String type=directory ? Document.MIME_TYPE_DIR : "application/octet-stream";
        Uri doc=DocumentsContract.createDocument(resolver,dir,type,name);
        if(doc==null) throw new IOException();
        if(!doc.equals(child(dir,name))) {
            DocumentsContract.deleteDocument(resolver,doc);
            throw new IOException("exists");
        }
        return doc;
    }
    private Uri resolve(String path,boolean mkdirs) throws Exception {
        Uri cur=root;
        if(path.equals(".")) return cur;
        if(path.length()>4096) throw new IOException();
        for(String part:path.split("/",-1)){
            Uri next=child(cur,part);
            if(next==null && mkdirs) next=create(cur,part,true);
            if(next==null) throw new FileNotFoundException();
            cur=next;
        }
        return cur;
    }
    private JSONObject info(Cursor c) throws Exception {
        return new JSONObject().put("name",c.getString(1)).put("directory",Document.MIME_TYPE_DIR.equals(c.getString(2)))
            .put("size",c.getLong(3)).put("modified",c.getLong(4));
    }
    private Uri renameExact(Uri dir,Uri doc,String name) throws Exception {
        Uri result=DocumentsContract.renameDocument(resolver,doc,name);
        if(result==null || !result.equals(child(dir,name))) throw new IOException();
        return result;
    }
    private void recoverReplacement() throws Exception {
        String record=recovery.getString(root.toString(),null);
        if(record==null) return;
        JSONObject state=new JSONObject(record);
        String dest=state.getString("destination");
        Uri dir=resolve(parent(dest),false);
        Uri backup=child(dir,state.getString("backup"));
        Uri current=child(dir,base(dest));
        if(backup!=null) {
            if(current==null) renameExact(dir,backup,base(dest));
            else if(child(dir,state.getString("source"))==null) {
                if(!DocumentsContract.deleteDocument(resolver,backup)) throw new IOException();
            } else throw new IOException("replacement conflict");
        }
        if(!recovery.edit().remove(root.toString()).commit()) throw new IOException();
    }
    private void replace(String from,String dest) throws Exception {
        if(!parent(from).equals(parent(dest))) throw new IOException();
        Uri dir=resolve(parent(dest),false);
        Uri src=resolve(from,false);
        Uri old=resolve(dest,false);
        String backup=".apparat-recovery-"+UUID.randomUUID();
        JSONObject state=new JSONObject().put("destination",dest).put("source",base(from)).put("backup",backup);
        if(!recovery.edit().putString(root.toString(),state.toString()).commit()) throw new IOException();
        try {
            renameExact(dir,old,backup);
            renameExact(dir,src,base(dest));
        } finally {
            recoverReplacement();
        }
    }
    String access(String op,String path,String arg) {
        try {
            recoverReplacement();
            if(op.equals("replace")) {
                replace(path,arg);
                return "{}";
            }
            if(op.equals("mkdir")) {
                resolve(path,true);
                return "{}";
            }
            if(op.equals("create") || op.equals("newdir")) {
                create(resolve(parent(path),false),base(path),op.equals("newdir"));
                return "{}";
            }
            Uri src=resolve(path,false);
            if(op.equals("stat") || op.equals("list")) {
                boolean list=op.equals("list");
                try(Cursor c=resolver.query(list ? children(src) : src,COLUMNS,null,null,null)) {
                    if(c==null) throw new IOException();
                    JSONArray entries=new JSONArray();
                    while(c.moveToNext()) {
                        if(!list) return info(c).toString();
                        if(entries.length()>=8192) throw new IOException();
                        entries.put(info(c));
                    }
                    if(!list) throw new FileNotFoundException();
                    return entries.toString();
                }
            }
            if(op.equals("open")) {
                ParcelFileDescriptor fd=resolver.openFileDescriptor(src,arg);
                if(fd==null) throw new IOException();
                return Integer.toString(fd.detachFd());
            }
            if(path.equals(".")) throw new IOException();
            if(op.equals("remove")){
                if(!DocumentsContract.deleteDocument(resolver,src)) throw new IOException();
                return "{}";
            }
            if(op.equals("rename")){
                Uri a=resolve(parent(path),false);
                Uri b=resolve(parent(arg),false);
                String name=base(arg);
                validate(name);
                if(child(b,name)!=null) throw new IOException("exists");
                if(!a.equals(b)) {
                    if(child(b,base(path))!=null) throw new IOException("exists");
                    src=DocumentsContract.moveDocument(resolver,src,a,b);
                    if(src==null) throw new IOException();
                }
                if(!base(path).equals(name)) {
                    Uri moved=DocumentsContract.renameDocument(resolver,src,name);
                    if(moved==null || !moved.equals(child(b,name))) throw new IOException();
                }
                return "{}";
            }
        } catch(FileNotFoundException missing) {
            return "!missing";
        } catch(Exception failure) {
            if("exists".equals(failure.getMessage())) return "!exists";
        }
        return "!provider";
    }
}
