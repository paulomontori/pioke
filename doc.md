# PioKe — Interactive Real-Time Karaoke Engine

> **A lightweight, cross-platform, interactive karaoke & chord synth engine built in Go.** Designed to run efficiently on single-board computers (Raspberry Pi), low-spec PCs, Smart TVs, and mobile devices.

---

## 📋 1. Project Overview

The goal of this project is to develop an open-source, ultra-lightweight **interactive karaoke engine** written in **Go**. 

The engine reads a structured text file (`.json` or `.yaml`) containing lyrics, timestamps, chords, and musical metadata. It parses the timeline and synchronizes two main outputs in real time:
1. **Visual Lyrics & Chords Rendering:** Highlights current words/syllables and displays chord diagrams or symbols.
2. **Real-Time Audio Synthesis:** Synthesizes audio tones for chords using lightweight wave generation or MIDI/SoundFont integration without requiring heavy pre-recorded audio tracks.

### Key Highlights
* **Language:** Go (Golang) — selected for zero-dependency binary compilation, cross-platform support, low CPU/RAM footprint, and simple concurrency models (`goroutines` & `channels`).
* **License:** GNU General Public License v3.0 (GPLv3) — ensures the project remains open-source forever and prevents proprietary monetization without source disclosure.
* **Target Platforms:** Raspberry Pi (Raspberry Pi OS / Linux Arm), Linux x86_64, Windows, macOS, Android / Smart TVs.

---

## 🛠️ 2. System Architecture

The application is split into four decoupled modules:

```
                  ┌──────────────────────┐
                  │   Song File (.json)  │
                  └──────────┬───────────┘
                             │
                             ▼
                  ┌──────────────────────┐
                  │    pkg/parser        │
                  └──────────┬───────────┘
                             │ Struct Data
                             ▼
                  ┌──────────────────────┐
                  │    pkg/engine        │◄─── Clock / Timer Ticker
                  └─────┬──────────┬─────┘
                        │          │
         Sync Events    │          │ Sync Events
                        ▼          ▼
      ┌───────────────────┐      ┌───────────────────┐
      │     pkg/audio     │      │      pkg/ui       │
      │  (Audio / Synth)  │      │   (CLI / TUI)     │
      └───────────────────┘      └───────────────────┘
```

### Module Breakdown

1. **`pkg/parser`**: Parses input files (`.json`), validates structure, handles time conversion, and loads data into memory (`Song` struct).
2. **`pkg/engine`**: Controls the playback lifecycle (play, pause, stop, seek). Uses `time.Ticker` to track playback position in milliseconds and emits events via Go channels.
3. **`pkg/audio`**: Real-time audio synthesizer. Plays frequency-based wave tones (sine, square, triangle) or outputs MIDI events for corresponding chords.
4. **`pkg/ui`**: Manages visual display. Initially built as a terminal UI (TUI) or simple 2D window (e.g., via `ebiten` or `termbox-go`).

---

## 📄 3. Song File Format Specification (`.json`)

Each song is defined by a simple, human-readable JSON schema:

```json
{
  "title": "Evidências",
  "artist": "Chitãozinho & Xororó",
  "bpm": 100,
  "key": "E",
  "time_signature": "4/4",
  "timeline": [
    {
      "time_ms": 12500,
      "duration_ms": 1200,
      "chord": "E",
      "lyric": "Quando eu digo "
    },
    {
      "time_ms": 13700,
      "duration_ms": 2000,
      "chord": "G#m",
      "lyric": "que não quero mais você..."
    }
  ]
}
```

---

## 🚀 4. MVP Functional Requirements

For the initial prototype (Minimum Viable Product):

- [x] Parse JSON song configuration into Go structs.
- [x] Maintain a main playback loop with millisecond accuracy using `time.Ticker`.
- [x] Print real-time synchronized lyrics and chords to `stdout` / terminal.
- [x] Provide clean, modular package architecture (`pkg/parser`, `pkg/engine`, `pkg/audio`, `pkg/ui`).

---

## 📜 5. Licensing

This project is licensed under the **GNU General Public License v3.0 (GPLv3)**.

* Anyone is free to use, run, modify, and distribute this software.
* Any modified versions or derivative works that are distributed **must remain open-source** under GPLv3.
