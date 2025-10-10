# meta-cc Screenshots and Demos

This directory contains visual assets for documentation and marketplace listing.

## Assets

### Installation Demo
- **File**: `installation-demo.gif`
- **Description**: Demonstrates `/plugin install` workflow from start to finish
- **Duration**: ~5-10 seconds
- **Size**: <500KB
- **Status**: ⚠️ Placeholder - needs to be created manually

### Feature Demonstrations

#### meta-coach Subagent
- **File**: `meta-coach-analysis.png`
- **Description**: Shows @meta-coach providing workflow analysis and recommendations
- **Dimensions**: 1200x800 (approximate)
- **Size**: <400KB
- **Status**: ⚠️ Placeholder - needs to be created manually

#### meta-viz Dashboard
- **File**: `meta-viz-dashboard.png`
- **Description**: Shows ASCII charts and visual analytics from /meta-viz command
- **Dimensions**: 1200x800 (approximate)
- **Size**: <400KB
- **Status**: ⚠️ Placeholder - needs to be created manually

## Usage

Referenced in:
- `README.md` (installation section)
- `docs/marketplace-listing.md` (feature showcase)
- `.claude-plugin/marketplace.json` (screenshots array)

## Creating Screenshots

### Installation Demo (GIF)

**Requirements**:
- meta-cc plugin installed in Claude Code
- Terminal recording tool (asciinema + agg, or similar)

**Steps**:
```bash
# Option 1: Using asciinema + agg (recommended)
asciinema rec installation-demo.cast
# Run commands: /plugin marketplace add yaleh/meta-cc
#                /plugin install meta-cc
# Stop recording: Ctrl+D
agg installation-demo.cast installation-demo.gif

# Option 2: Using terminalizer
terminalizer record installation-demo
# Run commands as above
terminalizer render installation-demo -o installation-demo.gif

# Option 3: Using ttygif
ttyrec installation-demo.ttyrec
# Run commands as above
# Exit: exit
ttygif installation-demo.ttyrec
```

**Optimization**:
```bash
# Reduce file size if needed
gifsicle -O3 --colors 256 installation-demo.gif -o installation-demo-opt.gif
mv installation-demo-opt.gif installation-demo.gif
```

### Feature Screenshots (PNG)

**Requirements**:
- meta-cc plugin installed and working in Claude Code
- Screenshot capture tool (native OS tool or Claude Code export)

**Steps**:

#### Capturing meta-coach Analysis
1. Start Claude Code with meta-cc installed
2. Run command: `@meta-coach analyze my workflow`
3. Wait for complete analysis output
4. Capture screenshot (ensure terminal is 1200x800 or larger)
5. Save as `meta-coach-analysis.png`

#### Capturing meta-viz Dashboard
1. Ensure project has session history data
2. Run command: `/meta-viz`
3. Wait for complete dashboard rendering
4. Capture screenshot (ensure all charts are visible)
5. Save as `meta-viz-dashboard.png`

**Screenshot Tools**:
- macOS: `Cmd+Shift+4` (selection) or `Cmd+Shift+5` (with options)
- Linux: `gnome-screenshot -a` or `scrot -s` or `flameshot gui`
- Windows: `Snipping Tool` or `Win+Shift+S`

**Optimization**:
```bash
# Optimize PNG files (requires pngquant)
pngquant --quality=65-80 --ext .png --force meta-coach-analysis.png
pngquant --quality=65-80 --ext .png --force meta-viz-dashboard.png

# Or use imagemagick
convert meta-coach-analysis.png -quality 85 meta-coach-analysis-opt.png
mv meta-coach-analysis-opt.png meta-coach-analysis.png
```

## Quality Checklist

Before committing assets, verify:

- [ ] **Installation Demo (GIF)**:
  - [ ] Shows complete `/plugin install` workflow
  - [ ] Duration is 5-10 seconds (not too long)
  - [ ] File size is <500KB
  - [ ] Text is readable in motion
  - [ ] Terminal prompt/commands are visible
  - [ ] Success message is shown

- [ ] **meta-coach Screenshot (PNG)**:
  - [ ] Shows complete analysis output
  - [ ] Text is readable at normal zoom
  - [ ] Demonstrates key recommendations
  - [ ] File size is <400KB
  - [ ] No sensitive project information visible

- [ ] **meta-viz Screenshot (PNG)**:
  - [ ] Shows ASCII charts and visualizations
  - [ ] All dashboard sections visible
  - [ ] Colors/formatting preserved
  - [ ] File size is <400KB
  - [ ] Charts are meaningful (not empty data)

## Next Steps

After creating the visual assets:

1. Place files in this directory (`docs/screenshots/`)
2. Verify file sizes: `ls -lh docs/screenshots/*.{png,gif}`
3. Test in documentation:
   - Preview README.md with images
   - Check marketplace-listing.md rendering
4. Update this README.md (change status from "Placeholder" to "Complete")
5. Commit assets to repository

## Notes

- All assets are **manually created** (require running Claude Code with meta-cc)
- Assets should demonstrate **real functionality** (not mockups)
- Keep file sizes optimized for fast loading
- Ensure no sensitive information (API keys, personal data) is visible
- Screenshots should be professional quality (clear, well-formatted)
