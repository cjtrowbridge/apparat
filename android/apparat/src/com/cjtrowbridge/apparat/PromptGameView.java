/*
 * Routes printable key events in the global canvas editor through the game's
 * hardware keyboard path. Software keyboards compose through the engine IME.
 */
package com.cjtrowbridge.apparat;

import android.content.Context;
import android.os.Build;
import android.view.KeyEvent;
import android.view.View;
import android.view.WindowInsets;
import android.view.WindowManager;
import android.widget.EditText;
import android.widget.FrameLayout;
import com.cjtrowbridge.apparat.apparatmobile.Apparatmobile;
import com.cjtrowbridge.apparat.apparatmobile.EbitenView;

final class PromptGameView extends EbitenView {
    PromptGameView(Context context) {
        super(context);
    }

    @Override public boolean dispatchKeyEventPreIme(KeyEvent event) {
        int character = event.getUnicodeChar();
        if (findFocus() instanceof EditText && character >= ' ' &&
                Character.isValidCodePoint(character) && !event.isCtrlPressed() &&
                !event.isAltPressed() && !event.isMetaPressed() &&
                "global-agentic-session".equals(Apparatmobile.activeTab())) {
            // EbitenView forwards these events to the canvas. Delivering the
            // same characters to EditText would also generate IME commits.
            if (event.getAction() == KeyEvent.ACTION_DOWN && event.getRepeatCount() == 0) {
                return onKeyDown(event.getKeyCode(), event);
            }
            if (event.getAction() == KeyEvent.ACTION_UP) {
                return onKeyUp(event.getKeyCode(), event);
            }
        }
        return super.dispatchKeyEventPreIme(event);
    }

    static void fitKeyboard(MainActivity activity, FrameLayout content) {
        if (Build.VERSION.SDK_INT < 30) return;
        activity.getWindow().setDecorFitsSystemWindows(false);
        activity.getWindow().setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_ADJUST_NOTHING);
        content.setOnApplyWindowInsetsListener(new View.OnApplyWindowInsetsListener() {
            @Override public WindowInsets onApplyWindowInsets(View target, WindowInsets insets) {
                int types = WindowInsets.Type.ime() | WindowInsets.Type.systemBars() | WindowInsets.Type.displayCutout();
                android.graphics.Insets safe = insets.getInsets(types);
                target.setPadding(safe.left, safe.top, safe.right, safe.bottom);
                return insets;
            }
        });
    }
}
