"""Present one checkout document in a GTK multiline editor.
Exchange bounded text and responses with Apparat through private process pipes.
"""
import json
import sys

import gi
gi.require_version("Gtk", "4.0")
from gi.repository import Gio, GLib, Gtk

title, body, position = json.loads(sys.stdin.readline())
app = Gtk.Application(application_id="app.apparat.Editor", flags=Gio.ApplicationFlags.NON_UNIQUE)


def activate(application):
    win = Gtk.ApplicationWindow(application=application, title=title)
    win.set_default_size(760, 540)
    bar = Gtk.HeaderBar()
    save = Gtk.Button(label="Save")
    undo = Gtk.Button(label="Undo")
    review = Gtk.Button(label="Review conflict")
    bar.pack_end(save)
    bar.pack_start(undo)
    bar.pack_end(review)
    review.set_visible(False)
    win.set_titlebar(bar)
    box = Gtk.Box(orientation=Gtk.Orientation.VERTICAL)
    view = Gtk.TextView(monospace=True, wrap_mode=Gtk.WrapMode.WORD_CHAR)
    buf = view.get_buffer()
    buf.set_text(body)
    buf.set_enable_undo(True)
    scroll = Gtk.ScrolledWindow(vexpand=True)
    scroll.set_child(view)
    box.append(scroll)
    status = Gtk.Label(label="")
    status.set_visible(False)
    box.append(status)
    win.set_child(box)

    def request(action):
        text = buf.get_text(buf.get_start_iter(), buf.get_end_iter(), True)
        print(json.dumps([action, text]), flush=True)
        return json.loads(sys.stdin.readline())

    def draft(*_):
        reason = request("draft")
        win.set_title(title + " •")
        status.set_text(reason)
        status.set_visible(bool(reason))

    def apply(*_):
        reason = request("save")
        if reason:
            status.set_text(reason)
            status.set_visible(True)
            review.set_visible(True)
        else:
            win.close()

    def compare(*_):
        dialog = Gtk.Dialog(title="Current checkout", transient_for=win, modal=True)
        preview = Gtk.TextView(editable=False, monospace=True, wrap_mode=Gtk.WrapMode.WORD_CHAR)
        preview.get_buffer().set_text(request("compare"))
        pane = Gtk.ScrolledWindow(vexpand=True)
        pane.set_child(preview)
        dialog.set_default_size(540, 400)
        dialog.get_content_area().append(pane)
        dialog.add_button("Back", Gtk.ResponseType.CANCEL)
        dialog.add_button("Keep my draft against this version", Gtk.ResponseType.ACCEPT)

        def respond(_, response):
            if response == Gtk.ResponseType.ACCEPT:
                status.set_text(request("rebase") or "Draft retained. Review it before saving.")
            dialog.close()
        dialog.connect("response", respond)
        dialog.present()

    buf.connect("changed", draft)
    save.connect("clicked", apply)
    undo.connect("clicked", lambda *_: buf.undo() if buf.get_can_undo() else None)
    review.connect("clicked", compare)
    def remember(*_):
        caret = buf.get_iter_at_mark(buf.get_insert()).get_offset()
        anchor = buf.get_iter_at_mark(buf.get_selection_bound()).get_offset()
        top = round(scroll.get_vadjustment().get_value())
        print(json.dumps(["position", json.dumps([caret, anchor, top])]), flush=True)
        sys.stdin.readline()
        return False

    def restore():
        caret, anchor, top = position
        n = buf.get_char_count()
        buf.select_range(buf.get_iter_at_offset(min(caret, n)), buf.get_iter_at_offset(min(anchor, n)))
        scroll.get_vadjustment().set_value(top)
        return False

    action = Gio.SimpleAction.new("save", None)
    action.connect("activate", lambda *_: apply())
    win.add_action(action)
    application.set_accels_for_action("win.save", ["<Primary>s"])
    win.connect("close-request", remember)
    win.present()
    GLib.idle_add(restore)
    view.grab_focus()


app.connect("activate", activate)
app.run([])
