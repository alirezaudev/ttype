#!/usr/bin/env python3
# This file written by AI :)
"""Play a VHS tape through a pty running bash and write an asciicast v2 file.

Tapes use a small subset of the VHS format: Type, Enter, Sleep, Ctrl+C, Down, Up,
Left, Right, Escape, Backspace, Tab, Space, Hide, Show, Set TypingSpeed. Output
and the other Set lines are ignored: the size comes from the command line.

Every run gets throwaway config and data dirs, so recording never shows or
touches real settings, history and bests, and the welcome screen stays closed.
Keys are typed at a fixed interval, and that interval is the wpm the recording
shows: wpm = (60000 / interval_ms) / 5, so the default 120ms is about 100 wpm.

Needs Linux and nothing outside the standard library. `make demo` runs it.

usage: tape2cast.py tape out.cast COLS ROWS BIN_DIR [THEME]
"""
import codecs, fcntl, json, os, pty, re, select, struct, sys, tempfile, termios, time

tape, out, cols, rows, bindir = sys.argv[1], sys.argv[2], int(sys.argv[3]), int(sys.argv[4]), sys.argv[5]
theme = sys.argv[6] if len(sys.argv) > 6 else "dracula"
KEYS = {"Enter": "\r", "Ctrl+C": "\x03", "Down": "\x1b[B", "Up": "\x1b[A", "Left": "\x1b[D", "Right": "\x1b[C",
        "Escape": "\x1b", "Backspace": "\x7f", "Tab": "\t", "Space": " "}


def dur(s):
    return float(s[:-2]) / 1000 if s.endswith("ms") else float(s.rstrip("s"))


steps, speed = [], 0.12  # ("key", delay, text) | ("sleep", s) | ("mark", "hide"/"show")
for raw in open(tape):
    line = raw.strip()
    if not line or line.startswith("#"):
        continue
    word = line.split()[0]
    if word == "Type":
        m = re.match(r'Type\s+(["\'`])(.*)\1\s*$', line)
        for ch in m.group(2):
            steps.append(("key", speed, ch))
    elif word in KEYS:
        n = int(line.split()[1]) if len(line.split()) > 1 else 1
        steps += [("key", 0.06, KEYS[word])] * n
    elif word == "Sleep":
        steps.append(("sleep", dur(line.split()[1])))
    elif word in ("Hide", "Show"):
        steps.append(("mark", word.lower()))
    elif word == "Set" and line.split()[1] == "TypingSpeed":
        speed = dur(line.split()[2])
    elif word in ("Set", "Output", "Source"):
        continue
    else:
        sys.exit(f"{tape}: unknown command: {line}")

cfg, data = tempfile.mkdtemp(), tempfile.mkdtemp()
os.makedirs(f"{cfg}/ttype")
json.dump({"onboarded": True, "theme": theme}, open(f"{cfg}/ttype/config.json", "w"))
env = {"PATH": f"{bindir}:/usr/bin:/bin", "HOME": os.environ["HOME"], "XDG_CONFIG_HOME": cfg, "XDG_DATA_HOME": data,
       "TTYPE_UPDATE": "off", "TERM": "xterm-256color", "COLORTERM": "truecolor", "LANG": "en_US.UTF-8", "PS1": "> "}

pid, fd = pty.fork()
if pid == 0:
    os.chdir(tempfile.mkdtemp())
    os.execve("/bin/bash", ["bash", "--norc", "--noprofile", "-i"], env)
fcntl.ioctl(fd, termios.TIOCSWINSZ, struct.pack("HHHH", rows, cols, 0, 0))
t0, ev = time.time(), []
decode = codecs.getincrementaldecoder("utf-8")("replace").decode  # a character can be split across reads


def pump(seconds):
    until = time.time() + seconds
    while True:
        left = until - time.time()
        if left <= 0:
            return
        if select.select([fd], [], [], left)[0]:
            try:
                data = os.read(fd, 65536)
            except OSError:
                return
            if not data:
                return
            text = decode(data)
            if text:
                ev.append([round(time.time() - t0, 4), "o", text])
            # answer the queries a real terminal would answer
            if b"\x1b]11;?" in data: os.write(fd, b"\x1b]11;rgb:2828/2a2a/3636\x1b\\")
            if b"\x1b]10;?" in data: os.write(fd, b"\x1b]10;rgb:f8f8/f8f8/f2f2\x1b\\")
            if b"\x1b[6n" in data: os.write(fd, b"\x1b[1;1R")
            if b"\x1b[c" in data or b"\x1b[0c" in data: os.write(fd, b"\x1b[?62;22c")


pump(0.6)
for step in steps:
    if step[0] == "key":
        pump(step[1])
        os.write(fd, step[2].encode())
    elif step[0] == "sleep":
        pump(step[1])
    else:
        pump(0.05)
        ev.append([round(time.time() - t0, 4), "m", step[1]])
pump(0.4)
with open(out, "w") as f:
    f.write(json.dumps({"version": 2, "width": cols, "height": rows}) + "\n")
    for e in ev:
        f.write(json.dumps(e, ensure_ascii=False) + "\n")
print(f"{os.path.basename(tape)}: {len(ev)} events, {ev[-1][0]:.1f}s")
