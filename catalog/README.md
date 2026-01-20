# Acteedog Plugin Catalog

This directory contains the plugin catalog for Acteedog, including plugin binaries (WASM), metadata, and icons.

## Directory Structure

```
plugin-catalog/
├── catalog.json                    # Main plugin catalog
├── plugins/
│   ├── github/
│   │   └── 1.0.0/
│   │       ├── plugin.wasm         # WASM binary
│   │       ├── manifest.json       # Plugin metadata
│   │       └── checksums.txt       # SHA256 checksum
│   ├── slack/
│   │   └── 1.0.0/
│   └── jira/                       # Coming soon
├── icons/
│   ├── github.svg
│   ├── slack.svg
│   └── jira.svg
└── scripts/
    └── generate-checksums.sh       # Generate checksums for plugins
```

## Usage

### For Development

Point Acteedog to this local catalog:

```bash
export ACTEEDOG_PLUGIN_CATALOG_PATH="file://$(pwd)/plugin-catalog/catalog.json"
npm run tauri dev
```

### Adding a New Plugin Version

1. Build the plugin:
   ```bash
   cd plugins/my-connector
   xtp plugin build
   ```

2. Copy to catalog:
   ```bash
   mkdir -p plugin-catalog/plugins/my-plugin/1.0.0
   cp plugins/my-connector/dist/plugin.wasm plugin-catalog/plugins/my-plugin/1.0.0/
   ```

3. Generate checksum:
   ```bash
   cd plugin-catalog
   ./scripts/generate-checksums.sh plugins/my-plugin/1.0.0
   ```

4. Create manifest.json:
   ```json
   {
     "id": "my-plugin",
     "name": "My Plugin",
     "version": "1.0.0",
     "description": "...",
     "checksum": "sha256:...",
     ...
   }
   ```

5. Update catalog.json:
   - Add plugin entry to `plugins` array
   - Use `file://` URLs for local development

## Future: Repository Split

When this catalog is moved to a separate repository:

1. Create `acteedog/plugin-catalog` repository
2. Copy this directory to the new repo
3. Update URLs in `catalog.json` from `file://` to `https://raw.githubusercontent.com/...`
4. Update hardcoded catalog URL in Acteedog's `PluginRegistry`

## Plugin Status

- `stable`: Production-ready, recommended
- `beta`: Feature-complete but may have bugs
- `experimental`: Early stage, expect breaking changes
- `deprecated`: No longer maintained
- `coming_soon`: Planned but not yet available

## Current Plugins

### Stable
- **GitHub Connector** (v1.0.0) - Track commits, PRs, and issues
- **Slack Connector** (v1.0.0) - Monitor channels and messages

### Coming Soon
- **Jira Connector** - Track issues and sprints
