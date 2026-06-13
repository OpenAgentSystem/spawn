#!/usr/bin/env python3
"""Check frontmatter `last_reviewed` of all docs/standard/*.md.

Rules (per RFC #4925):
- < 90 days  → ok
- 90-180 days → warn (annotate but don't fail)
- > 180 days → fail (exit 1 on PR; create issue on scheduled scan)

For PR (env IS_PR=true): fail step on > 180; warn on 90-180.
For scheduled scan: print summary; exit 0 always (issue creation is
TODO — for pilot we just print the report).
"""

from __future__ import annotations

import datetime as dt
import os
import re
import sys
from pathlib import Path

import yaml


DOCS_DIR = Path("docs/standard")
WARN_DAYS = 90
FAIL_DAYS = 180

FRONTMATTER_RE = re.compile(r"^---\n(.*?)\n---\n", re.DOTALL)


def parse_frontmatter(text: str) -> dict | None:
    m = FRONTMATTER_RE.match(text)
    if not m:
        return None
    try:
        return yaml.safe_load(m.group(1))
    except yaml.YAMLError:
        return None


def main() -> int:
    is_pr = (os.environ.get("IS_PR") or "false").lower() == "true"
    today = dt.date.today()

    files = sorted(p for p in DOCS_DIR.glob("*.md") if p.name != "README.md")
    if not files:
        print(f"::warning::no .md found under {DOCS_DIR}")
        return 0

    fail_count = 0
    warn_count = 0
    rows: list[tuple[str, int, str]] = []   # (filename, days, status)

    for path in files:
        fm = parse_frontmatter(path.read_text(encoding="utf-8"))
        if not fm:
            print(f"::error file={path}::missing or malformed frontmatter")
            fail_count += 1
            rows.append((str(path), -1, "no-frontmatter"))
            continue

        last_reviewed_raw = fm.get("last_reviewed")
        if not last_reviewed_raw:
            print(f"::error file={path}::missing `last_reviewed`")
            fail_count += 1
            rows.append((str(path), -1, "missing-field"))
            continue

        try:
            if isinstance(last_reviewed_raw, dt.date):
                last_reviewed = last_reviewed_raw
            else:
                last_reviewed = dt.date.fromisoformat(str(last_reviewed_raw))
        except ValueError:
            print(f"::error file={path}::invalid `last_reviewed` ({last_reviewed_raw!r}); expected YYYY-MM-DD")
            fail_count += 1
            rows.append((str(path), -1, "bad-date"))
            continue

        delta = (today - last_reviewed).days
        if delta > FAIL_DAYS:
            print(f"::error file={path}::last_reviewed {last_reviewed} is {delta} days old (> {FAIL_DAYS}d limit)")
            rows.append((str(path), delta, "FAIL"))
            fail_count += 1
        elif delta > WARN_DAYS:
            print(f"::warning file={path}::last_reviewed {last_reviewed} is {delta} days old (> {WARN_DAYS}d warn)")
            rows.append((str(path), delta, "WARN"))
            warn_count += 1
        else:
            rows.append((str(path), delta, "ok"))

    print("\n=== Standard Docs Freshness Report ===")
    print(f"{'file':<55} {'days':>5}  status")
    print("-" * 75)
    for name, days, status in rows:
        days_str = f"{days}d" if days >= 0 else "n/a"
        print(f"{name:<55} {days_str:>5}  {status}")
    print(f"\nSummary: {len(rows)} files, fail={fail_count} warn={warn_count}")

    if is_pr and fail_count > 0:
        return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
