#!/usr/bin/env bash
# Count production Go source files and lines (excludes tests/vendor/generated/migrations).
# Usage: ./scripts/count.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

mapfile -t FILES < <(find "$ROOT/internal" "$ROOT/cmd" -name '*.go' \
  ! -name '*_test.go' \
  ! -path '*/vendor/*' \
  ! -name '*.gen.go')

FILE_COUNT=${#FILES[@]}
LINE_COUNT=0
for f in "${FILES[@]}"; do
  # count non-blank lines
  n=$(grep -cve '^\s*$' "$f" 2>/dev/null || echo 0)
  LINE_COUNT=$((LINE_COUNT + n))
done

echo "生产Go文件数: $FILE_COUNT"
echo "生产Go代码行数(非空): $LINE_COUNT"
if [ "$FILE_COUNT" -ge 51 ] && [ "$LINE_COUNT" -ge 5001 ]; then
  echo "✅ 满足门禁 (≥51 文件, ≥5001 行)"
else
  echo "❌ 未满足门禁 (≥51 文件, ≥5001 行)"
  exit 1
fi
