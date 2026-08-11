Update the PiOke song format to support advanced features: syllables, pitch, velocity, and articulation.

Requirements:
1. In pkg/model/song.go, update the `TimelineEvent` struct to include:
   - `Velocity` (int): Volume intensity (0-127).
   - `Articulation` (string): e.g., "legato", "staccato".
   - `Syllables` ([]Syllable): A new list of syllable objects.

2. Create a new `Syllable` struct in pkg/model/song.go with:
   - `Text` (string): The syllable text (e.g., "Pa", "ra", "béns").
   - `OffsetMS` (int64): Start time of the syllable relative to the parent TimelineEvent's time_ms.
   - `DurationMS` (int64): How long the syllable lasts.
   - `Pitch` (string): Musical note for the melody (e.g., "G4", "C5").

3. In pkg/parser/json.go, ensure the JSON/YAML unmarshaling correctly maps these new fields. Add fallback logic so that if a song does not have `Syllables`, it still parses successfully using the basic `lyric` string.

4. In songs/evidencias.json, add an advanced timeline event example that includes the `velocity`, `articulation`, and a `syllables` array to validate the new structure.

5. Run `go test ./pkg/parser/...` and ensure the tests pass with the updated models.