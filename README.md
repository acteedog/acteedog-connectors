# Acteedog Plugins

Official and community-contributed connectors for [Acteedog](https://github.com/acteedog/acteedog-release).

## 📦 Connectors

### Features

- Activity: Retrieves activity information from the connected source.
- Context (Enrichment): Retrieves detailed information about contexts that are either directly associated with an activity or detected by the context detection mechanism.
- Context (Detection): When an activity or a context (including those obtained from other connectors) references a context supported by this connector, it is linked as a related context.
  - Example 1: Detected as a related context for another context (e.g., a GitHub pull request contains a link to a Jira issue, so the two are linked as related contexts).
  - Example 2: Detected as a related context for an activity (e.g., a Slack message contains a link to a GitHub pull request, so the activity and the context are linked).

### List

| Connector | Status      | Activity | Context (Enrichment) | Context (Detection) | Description                                                                |
| --------- | ----------- | -------- | -------------------- | ------------------- | -------------------------------------------------------------------------- |
| Github    | Released    | ✅       | ✅                   | ✅                  | Track commits, pull requests, issues, and reviews from GitHub repositories |
| Slack     | Released    | ✅       | ✅                   | ⬜                  | Monitor channels, direct messages, and threads from Slack workspaces       |
| Jira      | Development | N/A      | N/A                  | N/A                 | Track issues                                                               |

## 🏗️ Repository Structure

```
acteedog-connectors/
├── src/                          # Plugin source code
│   ├── github-connector/         # GitHub plugin (Go)
│   └── slack-connector/          # Slack plugin (Go)
├── catalog/                      # Plugin distribution catalog
│   ├── catalog.json              # Catalog metadata
│   ├── plugins/                  # Compiled WASM binaries
│   └── scripts/                  # Build and utility scripts
└── .github/workflows/            # CI/CD automation
```

## 🚀 Using Connectors

- Install connectors from the Acteedog connector settings dialog.
- Restart Acteedog.
- Configure the required settings and credentials on the Acteedog connector settings dialog.
  - For detailed instructions, please refer to the documentation: TBD
- Trigger an activity sync from the Acteedog activity screen.

## 🛠️ Developing Plugins

See [PLUGIN_DEVELOPMENT.md](./PLUGIN_DEVELOPMENT.md) for detailed instructions on creating new plugins.

### Quick Start

- Clone this repository and `cd acteedog-connectors/src`
- Initialize a new plugin: `xtp plugin init --schema-file acteedog-connector-schema.yaml --template Go --path your-connector`
- Implement the required functions according to the [plugin schema](./src/acteedog-connector-schema.yaml) and test code
- Build and test your plugin: `cd src/your-connector && xtp plugin build && xtp plugin test`
- Copy `dist/plugin.wasm` to `./catalog/connectors/your-connector/version/`
- Test the connector locally with Acteedog
  - Run `./catalog/scripts/generate-checksums.sh` and add the generated checksums as a new section in `./catalog/catalog.json`
  - Set the `ACTEEDOG_CONNECTOR_CATALOG_PATH` environment variable to the full path of `./catalog/catalog.json`
  - Start Acteedog and install the connector from the connector settings dialog

## 🤝 Contributing

We welcome contributions! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the Apache License, Version 2.0.
See the `LICENSE` file for the full license text.

### Notes

- This repository contains **WASM-based plugins** that are designed to run as **independent modules**.
- The main application (Acteedog) that loads and executes these plugins is **not open source** and is distributed under a separate proprietary license.
- Use of these plugins does **not** grant any rights to the source code, binaries, or services of Acteedog.

## 🔗 Links

- [Acteedog Main Repository](https://github.com/acteedog/acteedog-release)
- [Issue Tracker](https://github.com/acteedog/acteedog-plugins/issues)
