# ttype

Terminal typing practice with code modes, replays, and a wpm chart.
Inspired by [Monkeytype](https://monkeytype.com), but it never leaves your terminal.

```sh
ttype
```

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/alirezaudev/ttype/main/install.sh | sh
```

The script picks the right build for your machine, verifies its checksum, and
drops the binary and man page under `/usr/local` (or `~/.local`, if that is
where it can write). Later on, `ttype update` does the same in place.

Building it yourself needs Go 1.26:

```sh
make build && ./bin/ttype
```

To remove it again: `sh uninstall.sh`, or `--purge` to take your history with it.

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
ttype --output json         # print the result instead of showing it
```

Eight modes: `words` `sentences` `sql` `go` `backend` `python` `shell` `regex`.

In the two word modes, space commits the current word and skips whatever is
left of it — the same as the reference app. The code modes keep spaces literal,
because indentation is part of what you are practising.

Whatever you set in the settings panel (`S` on the result screen) becomes the
default for next time, so the flags are for one-offs.

## Keys

**While typing**

| Key | Action |
| --- | --- |
| `backspace` | Delete a character |
| `ctrl+backspace` `ctrl+w` `alt+backspace` | Delete the word and the space after it |
| `space` | Next word, skipping the rest of this one |
| `ctrl+o` | Toggle the live stats line |
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
