package com.cjtrowbridge.apparat;

import android.content.ContentProvider;
import android.content.ContentValues;
import android.content.Context;
import android.database.Cursor;
import android.database.MatrixCursor;
import android.net.Uri;
import android.os.ParcelFileDescriptor;
import android.provider.OpenableColumns;

import java.io.File;
import java.io.FileNotFoundException;

public class UpdateApkProvider extends ContentProvider {
    @Override
    public boolean onCreate() {
        return true;
    }

    @Override
    public String getType(Uri uri) {
        return "application/vnd.android.package-archive";
    }

    @Override
    public ParcelFileDescriptor openFile(Uri uri, String mode) throws FileNotFoundException {
        return ParcelFileDescriptor.open(apkFor(uri), ParcelFileDescriptor.MODE_READ_ONLY);
    }

    @Override
    public Cursor query(Uri uri, String[] projection, String selection, String[] selectionArgs, String sortOrder) {
        File apk;
        try {
            apk = apkFor(uri);
        } catch (FileNotFoundException error) {
            return null;
        }
        MatrixCursor cursor = new MatrixCursor(new String[]{OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE});
        cursor.addRow(new Object[]{"latest.apk", apk.length()});
        return cursor;
    }

    @Override
    public Uri insert(Uri uri, ContentValues values) {
        return null;
    }

    @Override
    public int delete(Uri uri, String selection, String[] selectionArgs) {
        return 0;
    }

    @Override
    public int update(Uri uri, ContentValues values, String selection, String[] selectionArgs) {
        return 0;
    }

    private File apkFor(Uri uri) throws FileNotFoundException {
        if (!"latest.apk".equals(uri.getLastPathSegment())) {
            throw new FileNotFoundException("unknown update artifact");
        }
        Context context = getContext();
        if (context == null) {
            throw new FileNotFoundException("provider context is not ready");
        }
        File apk = new File(context.getCacheDir(), "latest.apk");
        if (!apk.isFile()) {
            throw new FileNotFoundException("downloaded update is missing");
        }
        return apk;
    }
}
