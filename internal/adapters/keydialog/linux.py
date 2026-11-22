"""Show local GTK identity prompts or access a user Secret Service collection.
Read one JSON request on stdin and return one JSON string or null on stdout.
"""
import base64
import json
import os
import sys
import uuid
import gi
gi.require_version("Gtk", "3.0")
from gi.repository import Gtk

def credentials(req):
    gi.require_version("Secret", "1")
    from gi.repository import Secret
    service = Secret.Service.get_sync(Secret.ServiceFlags.LOAD_COLLECTIONS, None)
    collection = Secret.Collection.for_alias_sync(service, "default", Secret.CollectionFlags.NONE, None)
    if collection is None:
        return None
    schema = Secret.Schema.new("org.apparat.Node", Secret.SchemaFlags.NONE, {"id": Secret.SchemaAttributeType.STRING})
    if req["Op"] == "available":
        return "Use system keyring"
    ref = req["Value"]
    if req["Op"] == "store":
        ref = str(uuid.uuid4())
        ok = Secret.password_store_sync(schema, {"id":ref}, "default", "Apparat node", req["Value"], None)
        return ref if ok else None
    if req["Op"] == "remove":
        Secret.password_clear_sync(schema, {"id":ref}, None)
        return ""
    return Secret.password_lookup_sync(schema, {"id":ref}, None)

def prompt(req):
    op = req["Op"]
    if op in ("available","store","load","remove"):
        return credentials(req)
    if op in ("file","save"):
        save = op == "save"
        action = Gtk.FileChooserAction.SAVE if save else Gtk.FileChooserAction.OPEN
        dialog = Gtk.FileChooserDialog(title=req["Title"], action=action)
        dialog.add_button("Cancel", Gtk.ResponseType.CANCEL)
        dialog.add_button("Save" if save else "Open", Gtk.ResponseType.OK)
        if save:
            dialog.set_current_name("apparat-recovery.json")
        if dialog.run() != Gtk.ResponseType.OK:
            return None
        if save:
            with open(dialog.get_filename(),"xb",opener=lambda p,f:os.open(p,f,0o600)) as out:
                out.write(base64.b64decode(req["Value"],validate=True))
            return ""
        with open(dialog.get_filename(),"rb") as src:
            data = src.read(16385)
        return base64.b64encode(data).decode() if len(data)<=16384 else None
    dialog = Gtk.Dialog(title=req["Title"], modal=True)
    box = dialog.get_content_area()
    box.set_border_width(16)
    box.set_spacing(10)
    label = Gtk.Label(label=req["Message"])
    label.set_line_wrap(True)
    label.set_max_width_chars(48)
    box.add(label)
    dialog.add_button("Close" if op=="notice" else "Cancel", Gtk.ResponseType.CANCEL)
    if op=="password":
        entry = Gtk.Entry()
        entry.set_visibility(False)
        entry.set_input_purpose(Gtk.InputPurpose.PASSWORD)
        entry.set_activates_default(True)
        box.add(entry)
        dialog.add_button("Continue",1)
        dialog.set_default_response(1)
    else:
        for i, item in enumerate(req["Options"] or []):
            dialog.add_button(item,i+1)
    dialog.show_all()
    result = dialog.run()
    if result<1:
        return None
    return entry.get_text() if op=="password" else req["Options"][result-1]

try:
    print(json.dumps(prompt(json.load(sys.stdin))))
except Exception:
    print("null")
