"""Exchange text through GTK's negotiated clipboard targets on the requested display.
Copy acknowledges ownership on stdout and serves the selection until it is replaced.
"""
import os
import sys

try:
    import gi
    gi.require_version("Gtk", "3.0")
    gi.require_version("Gdk", "3.0")
    from gi.repository import Gdk, GLib, Gtk
except (ImportError, ValueError):
    sys.exit(7)

if not Gtk.init_check()[0]:
    sys.exit(4)

action = sys.argv[1]
selection = Gdk.SELECTION_CLIPBOARD
clip = Gtk.Clipboard.get(selection)


def transfer():
    if action == "copy":
        value = sys.stdin.buffer.read().decode("utf-8")
        clip.set_text(value, -1)
        Gdk.Display.get_default().sync()
        sys.stdout.buffer.write(b"1")
        sys.stdout.buffer.flush()
        if window is not None:
            window.hide()
        if os.environ["GDK_BACKEND"] == "x11":
            clip.connect("owner-change", owner_changed)
        return False
    available, targets = clip.wait_for_targets()
    metadata = {"TARGETS", "TIMESTAMP", "MULTIPLE", "SAVE_TARGETS"}
    if available and any(atom.name() not in metadata for atom in targets):
        if not clip.wait_is_text_available():
            sys.exit(3)
        value = clip.wait_for_text()
        if value is None:
            sys.exit(6)
        sys.stdout.buffer.write(value.encode("utf-8"))
    Gtk.main_quit()
    return False


def owner_changed(*args):
    if Gdk.selection_owner_get(selection) is None:
        Gtk.main_quit()


def focused(*args):
    if not window.is_active():
        return
    if not state["started"]:
        state["started"] = True
        GLib.timeout_add(50, transfer)


def focus_deadline():
    if not state["started"]:
        sys.exit(5)
    return False


window = None
state = {"started": False}
if os.environ["GDK_BACKEND"] == "wayland":
    window = Gtk.Window(title="Apparat clipboard")
    window.add(Gtk.Label(label="Select this window to complete the clipboard transfer."))
    window.connect("notify::is-active", focused)
    window.connect("delete-event", lambda *_: sys.exit(5))
    window.show_all()
    window.present()
    GLib.timeout_add_seconds(4, focus_deadline)
else:
    GLib.idle_add(transfer)
Gtk.main()
