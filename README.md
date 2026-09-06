# ttype

Terminal-first typing practice inspired by [Monkeytype](https://monkeytype.com).
A typing test that runs entirely in your terminal: eight text modes, downloadable
language lists, a wpm chart on the result screen, and a replay of every run.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/alirezaudev/ttype/main/install.sh | sh
```

Or build it yourself:

```sh
make build
./bin/ttype
```

`sh uninstall.sh` removes it again; add `--purge` to drop your history too.

## Usage

```sh
ttype                       # 60 second timed test
ttype --time 30             # 30 second test
ttype --words 25            # finish 25 words
ttype --mode go             # type Go snippets instead of words
ttype --punctuation --numbers
ttype --blind               # no feedback while typing
ttype --zen                 # words only, nothing else on screen
ttype --min-wpm 60          # fail the run if you drop below 60
ttype --seed 42             # same text every time
ttype --output json         # print the result instead of showing it
```

### Modes

`words` `sentences` `sql` `go` `backend` `python` `shell` `regex`

In `words` and `sentences`, space commits the word and skips whatever is left of
it. The code modes keep spaces literal, since indentation is part of the text.

### Commands

| Command | What it does |
| --- | --- |
| `ttype history` | Past results in a table; `enter` watches the run back |
| `ttype stats` | Averages and bests, with `--trend` and `--export csv` |
| `ttype config` | Show or change the saved defaults |
| `ttype languages` | List cached lists, `languages download` fetches them |
| `ttype clear` | Delete history, languages, or both |
| `ttype doctor` | Check the terminal, locale, clipboard and data dir |
| `ttype update` | Update to the latest release |
| `ttype completion` | Completion script for bash, zsh, fish or powershell |

## Keys

**While typing**

| Key | Action |
| --- | --- |
| `backspace` | Delete a character |
| `ctrl+backspace` `ctrl+w` | Delete the word, and the space before it |
| `space` | Next word (skips the rest of this one) |
| `ctrl+o` | Toggle the live stats line |
| `tab` `enter` | Restart |
| `?` | Help (before the first keystroke) |
| `esc` `ctrl+c` | Quit |

**On the result screen**

| Key | Action |
| --- | --- |
| `tab` `enter` `r` | Restart |
| `C` | Copy the result |
| `S` | Settings |
| `M` | Mode picker |
| `L` | Language picker |
| `esc` `q` | Quit |

## Stats

- **wpm** — correct characters ÷ 5, over the time taken.
- **raw** — every character typed, correct or not.
- **accuracy** — correct keystrokes as typed. Backspacing a mistake does not
  win the accuracy back.
- **consistency** — how even the per-second speed was; steady beats spiky.

Results live in `~/.local/share/ttype`, settings in `~/.config/ttype`.
Personal bests are tracked per configuration, so a 60 second English run is
never compared against a 15 second Spanish one.
