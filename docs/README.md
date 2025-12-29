# Documentation Assets

This directory contains documentation assets for the pkit project.

## Demo GIF

### Generating the Demo

1. **Install VHS**:
   ```bash
   go install github.com/charmbracelet/vhs@latest
   ```

2. **Ensure pkit is installed**:
   ```bash
   make install
   # or
   go install ./cmd/pkit
   ```

3. **Generate the demo**:
   ```bash
   vhs docs/demo.tape
   ```

   This will create `docs/demo.gif` (~2-3 MB).

4. **Preview the GIF**:
   - Open `docs/demo.gif` in your browser
   - Or use: `open docs/demo.gif` (macOS) / `xdg-open docs/demo.gif` (Linux)

### Customizing the Demo

Edit `demo.tape` to modify:
- **Timing**: Adjust `Sleep` durations
- **Commands**: Change what commands are shown
- **Appearance**: Modify `Theme`, `FontSize`, `Width`, `Height`
- **Speed**: Adjust `TypingSpeed` and `PlaybackSpeed`

Available themes: `Dracula`, `Monokai`, `Nord`, `Solarized`, `GitHub Dark`, etc.

### Tips

- Keep demos under 10 seconds for README viewing
- Use realistic commands that users would actually run
- Show the most impressive features first
- Add pauses after important actions
- Test the demo on different screen sizes

### File Size

- Current demo target: ~2-3 MB
- If too large, reduce:
  - Width/Height dimensions
  - PlaybackSpeed (increase value)
  - Duration (fewer commands)
  - Use `gifsicle` to optimize: `gifsicle -O3 demo.gif -o demo-opt.gif`
