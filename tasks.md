# PioKe — Frontend Roadmap & UI Implementation Options

> **Document Scope:** Interface layer options, cross-platform rendering considerations, step-by-step UI roadmap, and integration patterns for the PioKe interactive karaoke engine.

---

## 🎨 1. Architecture & Rendering Options

To maintain low latency and low CPU/RAM usage on Raspberry Pi and modest hardware, three progressive UI options are defined:

| Option | Technology Stack | Best Used For | Memory Footprint | CPU Overhead |
| :--- | :--- | :--- | :--- | :--- |
| **Option A (MVP)** | Terminal UI (`bubbletea` / `termbox-go`) | SSH / Headless / Pi CLI | ~10 MB | Ultra Low (< 1%) |
| **Option B (Recommended)** | Lightweight 2D Engine (`Ebitengine`) | Desktop / Raspberry Pi OS | ~35 MB | Low (1-3%) |
| **Option C (Cross-Platform)** | Web / Mobile WebView (`Wails` / WebSockets) | Smart TVs / Android / Web | ~80 MB | Moderate |

---

## 🛠️ 2. Core Frontend Dependencies (Go)

### 2.1 Terminal UI (Option A)
* **`github.com/charmbracelet/bubbletea`**: Elm-architecture framework for terminal interfaces.
  * **Role:** High-performance, reactive terminal rendering for lyric synchronization, progress bars, and chord highlights.
* **`github.com/charmbracelet/lipgloss`**: Terminal styling engine.
  * **Role:** Text formatting, colors, borders, and layout grids.

### 2.2 Graphical 2D Engine (Option B)
* **`github.com/hajimehoshi/ebiten/v2`**: Ultra-lightweight 2D game library for Go.
  * **Role:** Cross-platform hardware-accelerated rendering (OpenGL/Metal/DirectX/WebGL), font rendering, screen scaling, and smooth lyric scrolling animations.

---

## 🚀 3. Step-by-Step Frontend Implementation Roadmap

---

### Step 7: UI Layer Abstraction & Renderer Interface
* **Objective:** Decouple the frontend display from the backend engine via Go interfaces.
* **Tasks:**
  1. Define `Renderer` interface in `pkg/ui`:
     ```go
     type Renderer interface {
         Init() error
         RenderTick(event model.PlaybackEvent) error
         Close() error
     }
     ```
  2. Implement event listener inside `pkg/ui` subscribing to `chan PlaybackEvent` from the engine.
  3. Ensure UI rendering runs on the main thread while engine timing runs in a dedicated goroutine.

---

### Step 8: Terminal UI Prototype (TUI - Option A)
* **Objective:** Implement a lightweight terminal interface for headless/CLI environments.
* **Tasks:**
  1. Create `pkg/ui/tui` using `bubbletea`.
  2. Build components:
     * **Header:** Song title, artist, BPM, key, elapsed time / total duration.
     * **Lyric Display:** Active sentence with highlighted active word/syllable.
     * **Chord Banner:** Current chord and upcoming chord preview.
     * **Controls Bar:** Play/Pause status indicator, Volume level indicator.
  3. Wire keyboard inputs (`Space` = Play/Pause, `Q` = Quit, `Left`/`Right` = Seek).

---

### Step 9: Graphical 2D Interface (Ebitengine - Option B)
* **Objective:** Build a clean graphical interface suitable for monitors, TVs, and Raspberry Pi desktop environments.
* **Tasks:**
  1. Create `pkg/ui/gui` using `ebiten/v2`.
  2. Load custom TrueType/OpenType fonts (`golang.org/x/image/font`).
  3. Implement smooth text highlight transitions (karaoke-style color fill across lyrics based on millisecond progress).
  4. Render visual chord boxes/diagrams alongside lyric timing.
  5. Implement automatic window resizing and full-screen toggling (`F11`).

---

### Step 10: Song Selector & Playlist Navigation
* **Objective:** Allow users to browse local files and select songs without restarting the application.
* **Tasks:**
  1. Implement song library scanner in `pkg/parser` to index `.json` / `.yaml` files in a `./songs` directory.
  2. Build a file browser / song menu UI component.
  3. Implement song loading and state resets between tracks.

---

### Step 11: End-to-End System Integration & Benchmarking
* **Objective:** Validate performance, memory usage, and sync accuracy across target devices.
* **Tasks:**
  1. Integrate backend engine, synthesizer, and graphical UI into unified binary (`cmd/pioke/main.go`).
  2. Run latency and timing benchmarks on Raspberry Pi 3/4/5 and low-end hardware.
  3. Profile memory allocations and CPU utilization during continuous audio playback.
