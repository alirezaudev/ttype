package storage

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

// Recordings live in sidecar files rather than history.json: the stored-result
// schema is an additive-only contract and a keystroke log has no place in it.
//
//	"TTRP" · version byte · target length + bytes · event count ·
//	per event: kind byte, delta-ms since the previous event, rune (runes only)
const (
	replayMagic    = "TTRP"
	replayVersion  = 1
	maxReplayFiles = 50
)

var ErrNoReplay = errors.New("no replay recorded for this result")

func (s *JSONStore) replaysDir() string {
	return filepath.Join(s.dirs.Data, "replays")
}

// The id becomes a file name, so it may not wander out of the replays dir.
func replayFileName(id string) (string, error) {
	if id == "" || strings.ContainsAny(id, `/\.`) {
		return "", fmt.Errorf("invalid replay id %q", id)
	}
	return id + ".bin", nil
}

func (s *JSONStore) SaveReplay(id string, replay domain.Replay) error {
	name, err := replayFileName(id)
	if err != nil {
		return err
	}
	if len(replay.Events) == 0 {
		return fmt.Errorf("replay has no events")
	}
	if err := ensureDir(s.replaysDir()); err != nil {
		return fmt.Errorf("replays dir: %w", err)
	}
	if err := writeFile(filepath.Join(s.replaysDir(), name), encodeReplay(replay)); err != nil {
		return err
	}
	s.pruneReplays()
	return nil
}

func (s *JSONStore) LoadReplay(id string) (domain.Replay, error) {
	name, err := replayFileName(id)
	if err != nil {
		return domain.Replay{}, err
	}

	data, err := os.ReadFile(filepath.Join(s.replaysDir(), name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.Replay{}, ErrNoReplay
		}
		return domain.Replay{}, err
	}
	return decodeReplay(data)
}

// Housekeeping after a save, so failures are not worth reporting.
func (s *JSONStore) pruneReplays() {
	entries, err := os.ReadDir(s.replaysDir())
	if err != nil {
		return
	}

	type aged struct {
		name string
		mod  time.Time
	}
	var files []aged
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".bin") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, aged{e.Name(), info.ModTime()})
	}
	if len(files) <= maxReplayFiles {
		return
	}

	sort.Slice(files, func(i, j int) bool { return files[i].mod.Before(files[j].mod) })
	for _, f := range files[:len(files)-maxReplayFiles] {
		_ = os.Remove(filepath.Join(s.replaysDir(), f.name))
	}
}

func encodeReplay(replay domain.Replay) []byte {
	var buf bytes.Buffer
	buf.WriteString(replayMagic)
	buf.WriteByte(replayVersion)
	putUvarint(&buf, uint64(len(replay.Target)))
	buf.WriteString(replay.Target)
	putUvarint(&buf, uint64(len(replay.Events)))

	prev := int64(0)
	for _, ev := range replay.Events {
		buf.WriteByte(byte(ev.Kind))
		ms := ev.Offset.Milliseconds()
		delta := ms - prev
		if delta < 0 {
			delta = 0
		}
		prev = ms
		putUvarint(&buf, uint64(delta))
		if ev.Kind == domain.ReplayRune {
			putUvarint(&buf, uint64(ev.Rune))
		}
	}
	return buf.Bytes()
}

func decodeReplay(data []byte) (domain.Replay, error) {
	r := bytes.NewReader(data)

	magic := make([]byte, len(replayMagic))
	if _, err := r.Read(magic); err != nil || string(magic) != replayMagic {
		return domain.Replay{}, fmt.Errorf("not a replay file")
	}
	version, err := r.ReadByte()
	if err != nil {
		return domain.Replay{}, fmt.Errorf("truncated replay")
	}
	if version != replayVersion {
		return domain.Replay{}, fmt.Errorf("unsupported replay version %d", version)
	}

	targetLen, err := binary.ReadUvarint(r)
	if err != nil || targetLen > uint64(r.Len()) {
		return domain.Replay{}, fmt.Errorf("corrupt replay: target length")
	}
	target := make([]byte, targetLen)
	if _, err := r.Read(target); err != nil {
		return domain.Replay{}, fmt.Errorf("corrupt replay: target")
	}

	count, err := binary.ReadUvarint(r)
	if err != nil || count > uint64(r.Len())*2 {
		return domain.Replay{}, fmt.Errorf("corrupt replay: event count")
	}

	events := make([]domain.ReplayEvent, 0, count)
	ms := int64(0)
	for i := uint64(0); i < count; i++ {
		kind, err := r.ReadByte()
		if err != nil {
			return domain.Replay{}, fmt.Errorf("corrupt replay: event %d", i)
		}
		delta, err := binary.ReadUvarint(r)
		if err != nil {
			return domain.Replay{}, fmt.Errorf("corrupt replay: event %d delta", i)
		}
		ms += int64(delta)

		ev := domain.ReplayEvent{
			Kind:   domain.ReplayEventKind(kind),
			Offset: time.Duration(ms) * time.Millisecond,
		}
		switch ev.Kind {
		case domain.ReplayRune:
			value, err := binary.ReadUvarint(r)
			if err != nil {
				return domain.Replay{}, fmt.Errorf("corrupt replay: event %d rune", i)
			}
			ev.Rune = rune(value)
		case domain.ReplayBackspace, domain.ReplayDeleteWord:
		default:
			return domain.Replay{}, fmt.Errorf("corrupt replay: unknown event kind %d", kind)
		}
		events = append(events, ev)
	}

	return domain.Replay{Target: string(target), Events: events}, nil
}

func putUvarint(buf *bytes.Buffer, v uint64) {
	var tmp [binary.MaxVarintLen64]byte
	buf.Write(tmp[:binary.PutUvarint(tmp[:], v)])
}
