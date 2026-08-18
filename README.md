# ttype

Terminal-first typing practice inspired by [Monkeytype](https://monkeytype.com). A typing test that runs entirely in terminal.

## Build

```sh
make build
./bin/ttype
```

## Usage

```sh
ttype                       # 60 second timed test
ttype --time 30             # 30 second test
ttype --words 25            # finish 25 words
ttype --mode go             # type Go snippets instead of words
ttype --punctuation --numbers
ttype --blind               # no feedback while typing
ttype --zen                 # words only, nothing else on screen
ttype --seed 42             # same text every time
```

Results are saved locally, so `ttype history` and `ttype stats` show where you
are going.
