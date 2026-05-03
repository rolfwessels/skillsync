# Registry stores bundles in Claude-canonical format only

The registry always stores bundles as Claude-format artifacts. When syncing to a project, skillsync transforms bundles into the project's declared format(s) on the fly. We chose this over a multi-format registry (storing one copy per format) because the AI tooling landscape is in flux — new formats will emerge, and maintaining N copies in the registry would require authors to update every format on every change. A single canonical source with format-adapting transforms keeps the registry simple and lets new formats be added without touching existing bundles.

## Considered Options

- **Multi-format registry**: store `claude/`, `cursor/`, etc. variants side by side. Rejected: authoring burden, drift between copies.
- **Format-neutral registry**: invent a new neutral format. Rejected: unnecessary abstraction; Claude format is already well-understood by the primary author.
