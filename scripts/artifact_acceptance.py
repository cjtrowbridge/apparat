"""Exercise built Apparat artifacts with isolated runtime data.
Linux GUI validation runs inside a dedicated virtual display; Android uses
only the authorized Pixel 4a, preserves data and records the rendered screen.
"""
import argparse
import ctypes
import json
import os
from pathlib import Path
import re
import subprocess
import sys
import time


def run(*args, **kwargs):
    return subprocess.check_output(args, timeout=90, **kwargs)


def check_projects(path):
    from PIL import Image
    # The fresh registry may be empty. Check the hammer inside the selected
    # tab rectangle independently of registration count, device scale or rail.
    from PIL import ImageChops
    with Image.open(path) as frame:
        rgb = frame.convert('RGB')
    red, green, blue = rgb.split()

    def between(channel, low, high):
        return channel.point(lambda pixel: 255 if low <= pixel <= high else 0)

    selected = ImageChops.multiply(between(red, 58, 74), between(green, 92, 99))
    selected = ImageChops.multiply(selected, between(blue, 140, 154))
    bounds = selected.getbbox()
    if bounds:
        hammer = rgb.crop(bounds)
        orange = sum(r > 205 and 90 < g < 172 and b < 65 for r, g, b in hammer.getdata())
        if orange >= 10:
            # The add/clone control has a teal outline distinct from repo rows.
            teal = ImageChops.multiply(between(red, 35, 98), between(green, 151, 190))
            teal = ImageChops.multiply(teal, between(blue, 135, 175))
            if teal.histogram()[255] >= 120:
                return
            raise ValueError('Projects lacks the live add/clone control')
    raise ValueError('Projects hammer tab is not visibly selected')


def check_divider(x, xt, display, window, directory):
    from PIL import Image
    x.XResizeWindow(display, ctypes.c_ulong(int(window, 16)), 960, 820)
    x.XSync(display, 0)
    time.sleep(.7)
    start_x, y = 286, 250
    x.XDefaultRootWindow.restype = ctypes.c_ulong
    root = x.XDefaultRootWindow(display)
    root_x, root_y, child = ctypes.c_int(), ctypes.c_int(), ctypes.c_ulong()
    if not x.XTranslateCoordinates(display, ctypes.c_ulong(int(window, 16)), ctypes.c_ulong(root),
                                   start_x, y, ctypes.byref(root_x), ctypes.byref(root_y), ctypes.byref(child)):
        raise ValueError('Could not translate divider position to display coordinates')

    def capture(name):
        path = directory/f'divider-{name}.png'
        run('import', '-window', window, path)
        with Image.open(path) as screenshot:
            return screenshot.convert('RGB').crop((0, y, 500, y+1))

    before = capture('before')
    if before.getpixel((start_x, 0)) == before.getpixel((start_x-9, 0)):
        raise ValueError('Expected a visible Projects divider at its initial position')
    xt.XTestFakeMotionEvent(display, -1, root_x.value, root_y.value, 0)
    x.XSync(display, 0)
    time.sleep(.2)
    xt.XTestFakeButtonEvent(display, 1, 1, 0)
    x.XSync(display, 0)
    time.sleep(.2)
    try:
        for offset in (48, 100):
            xt.XTestFakeMotionEvent(display, -1, root_x.value+offset, root_y.value, 0)
            x.XSync(display, 0)
            time.sleep(.3)
        moved = capture('held')
    finally:
        xt.XTestFakeButtonEvent(display, 1, 0, 0)
        x.XSync(display, 0)
    time.sleep(.2)
    released = capture('released')
    for frame in (moved, released):
        if frame.getpixel((start_x, 0)) == before.getpixel((start_x, 0)):
            raise ValueError('Desktop divider did not leave its initial position')
        if frame.getpixel((start_x+100, 0)) == frame.getpixel((start_x+100-9, 0)):
            raise ValueError('Desktop divider did not track the second drag step')


def desktop_gui(binary, directory):
    x = ctypes.CDLL('libX11.so.6')
    xt = ctypes.CDLL('libXtst.so.6')
    x.XOpenDisplay.restype = ctypes.c_void_p
    display = ctypes.c_void_p(x.XOpenDisplay(None))
    if not display:
        raise RuntimeError('No test display')
    proc = subprocess.Popen([binary, '--runtime-dir', directory/'runtime'])
    try:
        for _ in range(80):
            windows = run('xwininfo', '-root', '-tree', text=True)
            found = re.search(r'(0x[0-9a-f]+) "Apparat"', windows)
            if found:
                break
            if proc.poll() is not None:
                raise ValueError('GUI exited before showing its window')
            time.sleep(.15)
        else:
            raise ValueError('Apparat window missing')
        window = int(found[1], 16)
        x.XSetInputFocus(display, ctypes.c_ulong(window), 1, 0)
        time.sleep(.7)
        for sym, down in ((0xffe9, 1), (0x32, 1), (0x32, 0), (0xffe9, 0)):
            key = x.XKeysymToKeycode(display, sym)
            xt.XTestFakeKeyEvent(display, key, down, 0)
            x.XSync(display, 0)
            time.sleep(.12)
        time.sleep(.4)
        check_divider(x, xt, display, found[1], directory)
        x.XResizeWindow(display, ctypes.c_ulong(window), 660, 820)
        x.XSync(display, 0)
        time.sleep(.7)
        xt.XTestFakeMotionEvent(display, -1, 150, 290, 0)
        for _ in range(10):
            xt.XTestFakeButtonEvent(display, 5, 1, 0)
            xt.XTestFakeButtonEvent(display, 5, 0, 0)
            x.XSync(display, 0)
            time.sleep(.05)
        time.sleep(.3)
        run('import', '-window', found[1], directory/'projects.png')
        check_projects(directory/'projects.png')
        def press(symbol):
            code = x.XKeysymToKeycode(display, symbol)
            for state in (1, 0):
                xt.XTestFakeKeyEvent(display, code, state, 0)
                x.XSync(display, 0)
                time.sleep(.15)
        alt = x.XKeysymToKeycode(display, 0xffe9)
        xt.XTestFakeKeyEvent(display, alt, 1, 0)
        press(0x35)
        xt.XTestFakeKeyEvent(display, alt, 0, 0)
        x.XSync(display, 0)
        time.sleep(.7)
        run('import','-window',found[1],directory/'stmpck.png')
        press(0xff0d)
        press(0xff0d)
        for direction in (0xff51, 0xff52, 0xff53, 0xff54):
            press(direction)
        run('import','-window',found[1],directory/'merge.png')
        press(0xff1b)
        save = json.loads((directory/'runtime/stmpck/games.json').read_text())
        if not save['Ready'][0] or not save['Merge']['HasUndo']:
            raise ValueError('STMPCK failed to save a playable Merge move')
    finally:
        proc.terminate()
        try:
            proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            proc.kill()
            proc.wait()


def android(binary, directory, digest):
    from shutil import which
    try:
        from scripts import android_release, build
    except ImportError:
        import android_release
        import build
    metadata = android_release.apk_metadata(build.default_sdk_root()/'build-tools'/build.ANDROID_BUILD_TOOLS,binary)
    if metadata['package'] != android_release.PACKAGE or metadata['signers'] != [android_release.expected_signer()]:
        raise ValueError('Unexpected APK package or signer; refusing installation')
    adb = which('adb')
    listing = run(adb, 'devices', '-l', text=True)
    devices = re.findall(r'^(\S+)\s+device\s+.*\bmodel:Pixel_4a\b', listing, re.M)
    if '08051JEC218861' not in devices:
        raise ValueError('Canonical authorized Pixel 4a required for device validation')
    def phone(*args, **kwargs):
        return run(adb, '-s', '08051JEC218861', *args, **kwargs)
    package = 'com.cjtrowbridge.apparat'
    phone('install', '-r', str(binary))
    remote = phone('shell', 'pm', 'path', package, text=True).strip().removeprefix('package:')
    if phone('shell', 'sha256sum', remote, text=True).split()[0] != digest:
        raise ValueError('Installed ordinary Apparat APK differs from candidate')
    phone('shell', 'am', 'start', '-W', '-n', package+'/.MainActivity')
    time.sleep(1)
    if not phone('shell','pidof',package,text=True).strip():
        raise ValueError('Apparat did not launch on the Pixel')
    phone('shell', 'input', 'keycombination', '-t', '180', 'KEYCODE_ALT_LEFT', 'KEYCODE_2')
    time.sleep(.7)
    path = directory/'projects.png'
    path.write_bytes(phone('exec-out', 'screencap', '-p'))
    check_projects(path)


def validate(plan, candidate, directory, digest):
    if plan.goos == 'android':
        android(candidate, directory, digest)
    elif plan.target == 'apparatd':
        env = dict(os.environ)
        env.pop('DISPLAY', None)
        run(candidate, '--runtime-dir', directory/'runtime', '--smoke-test', env=env)
    elif plan.goos == 'linux':
        run('xvfb-run', '-a', sys.executable, __file__, candidate, directory)
    else:
        raise ValueError(f'Native GUI qualification unavailable for {plan.goos}; not promoted')


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('binary', type=Path)
    parser.add_argument('directory', type=Path)
    args = parser.parse_args()
    desktop_gui(args.binary, args.directory)
