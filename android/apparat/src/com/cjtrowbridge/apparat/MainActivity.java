package com.cjtrowbridge.apparat;

import android.app.Activity;
import android.os.Bundle;
import android.view.Window;
import android.view.WindowManager;

import com.cjtrowbridge.apparat.apparatmobile.Apparatmobile;
import com.cjtrowbridge.apparat.apparatmobile.EbitenView;

public class MainActivity extends Activity {
    private EbitenView view;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(WindowManager.LayoutParams.FLAG_FULLSCREEN, WindowManager.LayoutParams.FLAG_FULLSCREEN);
        Apparatmobile.ready();
        view = new EbitenView(this);
        view.setFocusable(true);
        view.setFocusableInTouchMode(true);
        setContentView(view);
    }
}
