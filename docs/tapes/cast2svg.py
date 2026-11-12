#!/usr/bin/env python3
# This file written by AI :)
"""Turn an asciicast v2 recording into an animated SVG.

A GIF is a fixed grid of pixels; this stays sharp at any size and is a fraction
of the weight. Text is real text on a fixed cell grid; the characters that have
to line up exactly (braille charts, box lines, blocks) are drawn as shapes, so
they look the same in any font. Window and ANSI colours are Dracula.

Needs pyte (pip install pyte). `make demo` runs it.

--zoom-typing S enlarges the picture S times while a full-screen app is showing
only a few lines in the middle (a test being typed), and eases back to 1x when
the top of the screen fills (the results). It keeps a short test from sitting
small in an empty window.

usage: cast2svg.py in.cast out.svg [--hold 3] [--title ttype] [--zoom-typing 1.4]
"""
import copy, json, re, sys
from xml.sax.saxutils import escape
import pyte

CW, CH, FS, PAD, BAR = 10, 22, 16.6, 18, 34
BG, CHROME, RULE = "#282a36", "#21222c", "#44475a"  # Dracula
ANSI = {"default": "#f8f8f2", "black": "#21222c", "red": "#ff5555", "green": "#50fa7b", "brown": "#f1fa8c",
        "blue": "#bd93f9", "magenta": "#ff79c6", "cyan": "#8be9fd", "white": "#f8f8f2",
        "brightblack": "#6272a4", "brightred": "#ff6e6e", "brightgreen": "#69ff94", "brightbrown": "#ffffa5",
        "brightblue": "#d6acff", "brightmagenta": "#ff92df", "brightcyan": "#a4ffff", "brightwhite": "#ffffff"}
FONT = "ui-monospace, 'SF Mono', 'JetBrains Mono', Menlo, 'Cascadia Mono', Consolas, 'DejaVu Sans Mono', 'Liberation Mono', monospace"
BOX = {"─": "LR", "│": "UD", "┌": "RD", "┐": "LD", "└": "RU", "┘": "LU", "├": "URD", "┤": "ULD", "┬": "LRD",
       "┴": "LRU", "┼": "LRUD", "╶": "R", "╴": "L", "╵": "U", "╷": "D",
       "╭": "RD", "╮": "LD", "╰": "RU", "╯": "LU", "━": "LR", "┃": "UD"}
QUAD = {"▘": [(0, 0)], "▝": [(1, 0)], "▖": [(0, 1)], "▗": [(1, 1)], "▀": [(0, 0), (1, 0)], "▄": [(0, 1), (1, 1)],
        "█": [(0, 0), (1, 0), (0, 1), (1, 1)], "▌": [(0, 0), (0, 1)], "▐": [(1, 0), (1, 1)]}


def col(c, default):
    if c == "default":
        return default
    return ANSI.get(c) or ("#" + c)


def f(v):
    return f"{v:.2f}".rstrip("0").rstrip(".")


def drawn(ch, x):
    """SVG path data for characters we draw ourselves so they sit on the cell grid in any font."""
    o = ord(ch)
    if 0x2800 <= o <= 0x28FF:
        bits, d = o - 0x2800, []
        for bit, (dx, dy) in enumerate([(0, 0), (0, 1), (0, 2), (1, 0), (1, 1), (1, 2), (0, 3), (1, 3)]):
            if bits >> bit & 1:
                d.append(f"M{f(x + 1.2 + dx * 5)} {f(1.6 + dy * 5.2)}h3.2v3.2h-3.2z")
        return "fill", "".join(d)
    if ch in BOX:
        cx, cy, d = x + CW / 2, CH / 2, []
        for k in BOX[ch]:
            ex, ey = {"L": (x, cy), "R": (x + CW, cy), "U": (cx, 0), "D": (cx, CH)}[k]
            d.append(f"M{f(cx)} {f(cy)}L{f(ex)} {f(ey)}")
        return "stroke", "".join(d)
    if ch == "╱":
        return "stroke", f"M{f(x)} {CH}L{f(x + CW)} 0"
    if 0x2581 <= o <= 0x2588:  # lower one-eighth .. full block (sparklines, bars)
        h = (o - 0x2580) * CH / 8
        return "fill", f"M{f(x)} {f(CH - h)}h{CW}v{f(h)}h-{CW}z"
    if ch in QUAD:
        return "fill", "".join(f"M{f(x + qx * CW / 2)} {f(qy * CH / 2)}h{CW / 2}v{CH / 2}h-{CW / 2}z" for qx, qy in QUAD[ch])
    return None


def row_key(line, cols, cursor_x):
    return tuple((line[x].data, line[x].fg, line[x].bg, line[x].bold, line[x].underscore, line[x].reverse != (x == cursor_x))
                 for x in range(cols))


def render_row(key):
    out, x, n = [], 0, len(key)
    while x < n:  # backgrounds
        ch, fg, bg, bold, ul, rev = key[x]
        b = col(fg, ANSI["default"]) if rev else (None if bg == "default" else col(bg, BG))
        if b:
            x0 = x
            while x < n and (col(key[x][1], ANSI["default"]) if key[x][5] else (None if key[x][2] == "default" else col(key[x][2], BG))) == b:
                x += 1
            out.append(f'<rect x="{x0 * CW}" width="{(x - x0) * CW}" height="{CH}" fill="{b}"/>')
        else:
            x += 1
    x = 0
    fills, strokes = {}, {}
    while x < n:  # text runs with one style; drawn glyphs collected per colour
        ch, fg, bg, bold, ul, rev = key[x]
        c = (BG if bg == "default" else col(bg, BG)) if rev else col(fg, ANSI["default"])
        if ch == " ":
            x += 1
            continue
        d = drawn(ch, x * CW)
        if d:
            (fills if d[0] == "fill" else strokes).setdefault(c, []).append(d[1])
            x += 1
            continue
        x0, s = x, ""
        while x < n and key[x][0] != " " and not drawn(key[x][0], 0) and (key[x][1], key[x][2], key[x][3], key[x][4], key[x][5]) == (fg, bg, bold, ul, rev):
            s += key[x][0]
            x += 1
        # keep single spaces inside a run so words stay one text element
        while x + 1 < n and key[x][0] == " " and key[x + 1][0] != " " and not drawn(key[x + 1][0], 0) and (key[x + 1][1], key[x + 1][2], key[x + 1][3], key[x + 1][4], key[x + 1][5]) == (fg, bg, bold, ul, rev) and not ul:
            s += " "
            x += 1
            while x < n and key[x][0] != " " and not drawn(key[x][0], 0) and (key[x][1], key[x][2], key[x][3], key[x][4], key[x][5]) == (fg, bg, bold, ul, rev):
                s += key[x][0]
                x += 1
        a = f' fill="{c}"' + (' font-weight="700"' if bold else "") + (' text-decoration="underline"' if ul else "")
        tl = f' textLength="{len(s) * CW}" lengthAdjust="spacingAndGlyphs"' if len(s) > 1 else ""
        out.append(f'<text x="{x0 * CW}" y="16"{a}{tl} xml:space="preserve">{escape(s)}</text>')
    for c, ds in fills.items():
        out.append(f'<path fill="{c}" d="{"".join(ds)}"/>')
    for c, ds in strokes.items():
        out.append(f'<path fill="none" stroke="{c}" stroke-width="1.3" d="{"".join(ds)}"/>')
    return "".join(out)


def main():
    src, dst = sys.argv[1], sys.argv[2]
    hold = float(sys.argv[sys.argv.index("--hold") + 1]) if "--hold" in sys.argv else 3.0
    title = sys.argv[sys.argv.index("--title") + 1] if "--title" in sys.argv else "ttype"
    zoom = float(sys.argv[sys.argv.index("--zoom-typing") + 1]) if "--zoom-typing" in sys.argv else 0
    lines = open(src).read().splitlines()
    head = json.loads(lines[0])
    cols, rows = head["width"], head["height"]
    screen = pyte.Screen(cols, rows)
    stream = pyte.Stream(screen)
    frames = []  # (t, [row keys]); t excludes hidden stretches
    hidden, hide_at, skipped, saved = False, 0.0, 0.0, None
    app = []  # (t, in alternate screen) at each frame, for --zoom-typing
    for line in lines[1:]:
        t, kind, data = json.loads(line)
        if kind == "m":
            if data == "hide":
                hidden, hide_at = True, t
            else:
                hidden, skipped = False, skipped + (t - hide_at)
            continue
        # pyte has no alternate screen: keep the shell's screen ourselves while an app owns the display
        for part in re.split(r"(\x1b\[\?1049[hl])", data):
            if part == "\x1b[?1049h":
                saved = (copy.deepcopy(screen.buffer), screen.cursor.x, screen.cursor.y)
                screen.erase_in_display(2)
            elif part == "\x1b[?1049l":
                if saved:
                    screen.buffer, screen.cursor.x, screen.cursor.y = saved[0], saved[1], saved[2]
                    saved = None
            elif part:
                stream.feed(part)
        if hidden:
            continue
        t -= skipped
        cur = screen.cursor
        snap = [row_key(screen.buffer[y], cols, cur.x if (y == cur.y and not cur.hidden) else -1) for y in range(rows)]
        if frames and t - frames[-1][0] < 0.04:
            frames[-1] = (frames[-1][0], snap)
        elif not frames or frames[-1][1] != snap:
            frames.append((t, snap))
        app.append((t, saved is not None))
    t0 = frames[0][0]
    frames = [(t - t0, s) for t, s in frames]
    total = frames[-1][0] + hold
    blank = tuple((" ", "default", "default", False, False, False) for _ in range(cols))
    W, H = cols * CW + 2 * PAD, rows * CH + 2 * PAD + BAR
    o = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{W}" height="{H}" viewBox="0 0 {W} {H}" role="img" aria-label="{escape(title, {chr(34): "&quot;"})}" font-family="{FONT}" font-size="{FS}">',
         f'<rect x="0.5" y="0.5" width="{W - 1}" height="{H - 1}" rx="10" fill="{BG}" stroke="{RULE}"/>',
         f'<path d="M0.5 {BAR}V10.5a10 10 0 0 1 10-10H{W - 10.5}a10 10 0 0 1 10 10V{BAR}z" fill="{CHROME}"/>',
         f'<circle cx="20" cy="{BAR / 2}" r="5.5" fill="#ff5555"/><circle cx="38" cy="{BAR / 2}" r="5.5" fill="#f1fa8c"/><circle cx="56" cy="{BAR / 2}" r="5.5" fill="#50fa7b"/>',
         f'<text x="{W / 2}" y="{BAR / 2 + 4.5}" text-anchor="middle" font-size="12.5" fill="#6272a4">{escape(title)}</text>']
    n_states = 0
    # --zoom-typing: from the app taking the screen until its top rows fill
    z = None
    if zoom:
        start = next((t - t0 for t, on in app if on), None)
        if start is not None:
            top = lambda snap: any(k != blank for k in snap[:3])
            end = next((t for t, snap in frames if t > start + 0.5 and top(snap)), None)
            first = next(snap for t, snap in frames if t >= start + 0.3)
            used = [y for y in range(rows - 5) if first[y] != blank]
            if end and used:
                z = (start + 0.25, end - 0.3, (min(used) + max(used) + 1) / 2 * CH)
    if z:
        za, zb, zy = z
        k = [za / total, (za + 0.5) / total, zb / total, (zb + 0.3) / total]
        o.append(f'<clipPath id="screen"><rect x="1" y="{BAR}" width="{W - 2}" height="{H - BAR - 1}" rx="9"/></clipPath>')
        o.append(f'<g clip-path="url(#screen)"><g transform="translate({W / 2} {BAR + PAD + zy:.1f})"><g>'
                 f'<animateTransform attributeName="transform" type="scale" values="1;1;{zoom};{zoom};1;1" '
                 f'keyTimes="0;{k[0]:.4f};{k[1]:.4f};{k[2]:.4f};{k[3]:.4f};1" calcMode="spline" '
                 f'keySplines="0 0 1 1;0.4 0 0.2 1;0 0 1 1;0.4 0 0.2 1;0 0 1 1" dur="{total:.2f}s" repeatCount="indefinite"/>'
                 f'<g transform="translate({-W / 2} {-(BAR + PAD + zy):.1f})">')
    for y in range(rows):
        spans, start, key = [], 0.0, frames[0][1][y]
        for t, snap in frames[1:]:
            if snap[y] != key:
                spans.append((start, t, key))
                start, key = t, snap[y]
        spans.append((start, total, key))
        for a, b, key in spans:
            if key == blank:
                continue
            body = render_row(key)
            if not body:
                continue
            n_states += 1
            ka, kb = a / total, b / total
            if a == 0 and b == total:
                anim, vis = "", ""
            else:
                vals, kts = [], []
                if a > 0:
                    vals.append("hidden"); kts.append("0")
                vals.append("visible"); kts.append(f"{ka:.5f}" if a > 0 else "0")
                if b < total:
                    vals.append("hidden"); kts.append(f"{kb:.5f}")
                anim = f'<animate attributeName="visibility" calcMode="discrete" values="{";".join(vals)}" keyTimes="{";".join(kts)}" dur="{total:.2f}s" repeatCount="indefinite"/>'
                vis = ' visibility="hidden"' if a > 0 else ""
            o.append(f'<g transform="translate({PAD} {BAR + PAD + y * CH})"{vis}>{anim}{body}</g>')
    if z:
        o.append("</g></g></g></g>")
    o.append("</svg>\n")
    open(dst, "w").write("\n".join(o))
    print(f"{len(frames)} frames, {n_states} row states, {total:.1f}s loop")


main()
