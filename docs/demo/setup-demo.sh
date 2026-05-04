#!/bin/bash
set -e

rm -rf /tmp/ss-demo
mkdir -p /tmp/ss-demo/registry/skills/tdd
mkdir -p /tmp/ss-demo/linux-project/.skillsync
mkdir -p /tmp/ss-demo/windows-project/.skillsync

cat > /tmp/ss-demo/registry/skills/tdd/bundle.toml << 'EOF'
description = "Test-driven development red-green-refactor loop"
tags = ["testing", "tdd"]
EOF

cat > /tmp/ss-demo/registry/skills/tdd/SKILL.md << 'EOF'
# TDD Skill

Red-green-refactor loop for test-driven development.
EOF

cat > /tmp/ss-demo/linux-project/.skillsync/config.toml << 'EOF'
registry = "/tmp/ss-demo/registry"
formats = ["claude"]
bundles = ["skills/tdd"]
EOF

cd /tmp/ss-demo/linux-project && skillsync pull > /dev/null 2>&1

cat > /tmp/ss-demo/linux-project/.claude/skills/tdd/SKILL.md << 'EOF'
# TDD Skill

Red-green-refactor loop for test-driven development.

## Tip

Always write the failing test before writing any code.
EOF

cat > /tmp/ss-demo/windows-project/.skillsync/config.toml << 'EOF'
registry = "/tmp/ss-demo/registry"
formats = ["claude"]
bundles = []
EOF
