<div align="center">

# ttype

**Terminal typing practice with code modes, replays, and a wpm chart.**
Inspired by [Monkeytype](https://monkeytype.com), but it never leaves your terminal.

[![CI](https://github.com/alirezaudev/ttype/actions/workflows/ci.yml/badge.svg)](https://github.com/alirezaudev/ttype/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/alirezaudev/ttype?color=blue)](https://github.com/alirezaudev/ttype/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/alirezaudev/ttype)](go.mod)
[![Platforms](https://img.shields.io/badge/platforms-linux%20%7C%20macos%20%7C%20windows-lightgrey)](#install)

<img src="https://github.com/alirezaudev/ttype/releases/download/v1.0.0/demo.gif" alt="A twelve word test, one typo, and the results screen">

</div>

## Why

- **For fun.** I wanted to stop tabbing out to a browser to warm up, so I wrote the thing I kept opening.
- Eight text modes, including **Go, SQL, Python, shell and regex** snippets, where spaces and brackets are part of the practice.
- Every run is **recorded keystroke by keystroke** and can be watched back at the speed it was typed.
- A per-second **wpm chart**, a consistency score, and a heatmap of the characters you actually miss.
- **Blind mode**, **zen mode**, and a **min-wpm threshold** that ends the run when you slow down.
- Downloadable word lists, three themes, shell completions, and a man page.
- One static binary. No browser, no runtime, no config file to write by hand.

## Install

### Linux and macOS

```sh
curl -fsSL https://raw.githubusercontent.com/alirezaudev/ttype/main/install.sh | sh
```

Or, if you would rather use wget:

```sh
wget -qO- https://raw.githubusercontent.com/alirezaudev/ttype/main/install.sh | sh
```

The script picks the right build for your machine, checks it against
`checksums.txt`, and installs the binary and the man page under `/usr/local` —
or `~/.local`, if that is where it can write. After that, `ttype update` does
the same thing in place. It leaves copies installed by a package manager or
`go install` alone and tells you the command to use instead.

### Windows

Download `ttype_<version>_windows_amd64.tar.gz` (or `_arm64`) from the
[latest release](https://github.com/alirezaudev/ttype/releases/latest), unpack
it, and put `ttype.exe` somewhere on your `PATH`.

### With Go

```sh
go install github.com/alirezaudev/ttype/cmd/ttype@latest
```

The binary lands in `$(go env GOPATH)/bin`, so make sure that is on your `PATH`.

### From source

Needs Go 1.26:

```sh
git clone https://github.com/alirezaudev/ttype.git
cd ttype
make build
./bin/ttype
```

### Uninstall

`ttype uninstall` removes the binary and the man page, and `--purge` takes your
history with it. If the binary is already gone:

```sh
curl -fsSL https://raw.githubusercontent.com/alirezaudev/ttype/main/uninstall.sh | sh
```

## Typing

```sh
ttype                       # 60 second timed test
ttype --time 30             # a shorter one
ttype --words 25            # race a word count instead of the clock
ttype --mode go             # Go snippets instead of prose
ttype --language spanish    # downloaded once, cached after that
ttype --punctuation --numbers
ttype --blind               # no feedback until the run ends
ttype --zen                 # words only, nothing else on screen
ttype --min-wpm 60          # fail the run if you drop below 60
ttype --seed 42             # same text, every time
```

Eight modes: `words` `sentences` `sql` `go` `backend` `python` `shell` `regex`.

In the two word modes, space commits the current word and skips whatever is
left of it — the same as the reference app. The code modes keep spaces literal,
because indentation is part of what you are practising.

Whatever you set in the settings panel (`ctrl+s`, or `S` on the result screen)
becomes the default for next time, so the flags are for one-offs.

## More of it

<details>
<summary><b>Code modes</b> — Go, SQL and shell, then the mode picker</summary>

![Typing Go, SQL and shell snippets, then the mode picker](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/modes.gif)

</details>

<details>
<summary><b>Replays</b> — watch a past run play itself back</summary>

![Browsing history and watching a run replay at 1x and 2x](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/replay.gif)

</details>

<details>
<summary><b>Blind, zen and min-wpm</b></summary>

![Blind mode, zen mode, and a run that fails the min-wpm threshold](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/gameplay.gif)

</details>

<details>
<summary><b>Themes</b> — default, monokai, dracula</summary>

![The same test in the default, monokai and dracula themes](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/themes.gif)

</details>

<details>
<summary><b>Stats</b> — averages, a trend sparkline, CSV out</summary>

![The stats screen, the trend sparkline and a CSV export](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/stats.gif)

</details>

<details>
<summary><b>Help and settings</b></summary>

![The help overlay and the settings panel](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/help.gif)

</details>

<details>
<summary><b>The CLI</b> — the command list, doctor, and saved defaults</summary>

![ttype --help, ttype doctor and ttype config](https://github.com/alirezaudev/ttype/releases/download/v1.0.0/cli.gif)

</details>

## Keys

**While typing**

| Key | Action |
| --- | --- |
| `backspace` | Delete a character |
| `ctrl+backspace` `ctrl+w` `alt+backspace` | Delete the word and the space after it |
| `space` | Next word, skipping the rest of this one |
| `ctrl+o` | Toggle the live stats line |
| `ctrl+s` | Settings, without leaving the run |
| `tab` `enter` | Restart with fresh text |
| `?` | Help — before the first keystroke; after that it is just a character |
| `esc` `ctrl+c` | Quit |

**On the result screen**

| Key | Action |
| --- | --- |
| `tab` `enter` `r` | Restart |
| `C` | Copy the result line |
| `S` | Settings |
| `M` | Mode picker |
| `L` | Language picker |
| `u` | Releases page, when there is a newer version |
| `esc` `q` | Quit |

**Watching a replay** (`enter` on a row in `ttype history`)

| Key | Action |
| --- | --- |
| `space` | Pause |
| `1` `2` `4` | Speed |
| `r` | Start over |
| `esc` `q` | Back |

## Commands

| Command | What it does |
| --- | --- |
| `ttype history` | Past runs in a table; `enter` plays one back keystroke by keystroke |
| `ttype stats` | Averages and bests — `--trend` draws a sparkline, `--export csv` dumps the lot |
| `ttype config` | Show or change the saved defaults |
| `ttype languages` | List cached word lists; `languages download` fetches every one |
| `ttype clear` | Delete history, languages, or both |
| `ttype doctor` | Check the terminal, locale, clipboard and data directory |
| `ttype update` | Update to the latest release |
| `ttype uninstall` | Remove the binary and man page; `--purge` takes your data too |
| `ttype completion` | Completion script for bash, zsh, fish or powershell |

Shell completions are also generated into `contrib/completions/`, and
`contrib/zsh/` has a plugin that sets them up along with a `tt` alias.

## What the numbers mean

- **wpm** — correct characters ÷ 5, over the time taken.
- **raw** — every character typed, right or wrong.
- **accuracy** — correct keystrokes as you typed them. Backspacing over a
  mistake fixes the text, not the accuracy.
- **consistency** — how even your per-second speed was. Steady beats spiky,
  even at the same average.
- **chars** — correct / incorrect / extra / skipped.

Personal bests are tracked per configuration, so a 60 second English run is
never measured against a 15 second Spanish one with punctuation on.

## Where things live

| Path | Contents |
| --- | --- |
| `~/.config/ttype/config.json` | Saved defaults |
| `~/.local/share/ttype/history.json` | Every run, newest 1000 kept |
| `~/.local/share/ttype/bests.json` | Personal bests |
| `~/.local/share/ttype/replays/` | Recorded runs, newest 50 kept |
| `~/.local/share/ttype/languages/` | Downloaded word lists |

`XDG_CONFIG_HOME` and `XDG_DATA_HOME` are honoured; `NO_COLOR` turns the colour
off. `man ttype` has the long version of all of this.
