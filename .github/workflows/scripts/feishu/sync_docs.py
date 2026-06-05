#!/usr/bin/env python3
"""Sync docs/standard/ Markdown files to Lark Wiki nodes.

Reads `docs/standard/.feishu-sync.yml` for mapping. Strips frontmatter,
prepends a banner identifying GitHub as source, then writes content to
each configured Lark Wiki node via the Lark Open API.

Used by `.github/workflows/sync-docs-to-feishu.yml`.

Required env:
  LARK_APP_ID
  LARK_APP_SECRET
  COMMIT_SHA (full SHA, from github.sha)
  REPO_URL (https://github.com/<owner>/<repo>)

Optional env:
  DRY_RUN (true|false; default false)
"""

from __future__ import annotations

import datetime as dt
import os
import re
import sys
from pathlib import Path
from typing import Any

import requests
import yaml


DOCS_DIR = Path("docs/standard")
SYNC_CONFIG = DOCS_DIR / ".feishu-sync.yml"

LARK_API_BASE = "https://open.feishu.cn"
FRONTMATTER_RE = re.compile(r"^---\n.*?\n---\n", re.DOTALL)


def load_config() -> dict[str, Any]:
    if not SYNC_CONFIG.exists():
        sys.exit(f"::error::{SYNC_CONFIG} not found")
    with open(SYNC_CONFIG, encoding="utf-8") as f:
        return yaml.safe_load(f) or {}


def strip_frontmatter(text: str) -> str:
    return FRONTMATTER_RE.sub("", text, count=1)


def format_banner(template: str, commit_sha: str, repo_url: str, source_path: str) -> str:
    short_sha = commit_sha[:8]
    source_url = f"{repo_url}/blob/{commit_sha}/{source_path}"
    sync_at = dt.datetime.now(dt.UTC).strftime("%Y-%m-%d %H:%M UTC")
    return (
        template.replace("{{source_url}}", source_url)
        .replace("{{sync_at}}", sync_at)
        .replace("{{commit_sha_short}}", short_sha)
        + "\n---\n\n"
    )


def get_tenant_access_token(app_id: str, app_secret: str) -> str:
    resp = requests.post(
        f"{LARK_API_BASE}/open-apis/auth/v3/tenant_access_token/internal",
        json={"app_id": app_id, "app_secret": app_secret},
        timeout=10,
    )
    resp.raise_for_status()
    data = resp.json()
    if data.get("code", 0) != 0:
        sys.exit(f"::error::Lark token request failed: {data}")
    return data["tenant_access_token"]


def update_wiki_node_content(
    token: str, node_id: str, title: str, body: str, dry_run: bool
) -> None:
    """Replace the document content at the given Wiki node.

    Lark Wiki node points to a Docx document. We update via Docx update_block
    or replace_blocks API. For first iteration we use the simpler "raw
    content replace" via batch_update on the document body block.

    For the pilot we do not implement full block-level diffing; the
    sync_docs script overwrites the document body each run. Direct
    editing in Lark will be lost on next sync (which is the intent per
    RFC #4925).
    """
    if dry_run:
        print(f"[dry-run] would push {len(body)} bytes to wiki node {node_id} ({title})")
        return

    # Get the document_id behind the wiki node
    resp = requests.get(
        f"{LARK_API_BASE}/open-apis/wiki/v2/spaces/get_node",
        params={"token": node_id, "obj_type": "wiki"},
        headers={"Authorization": f"Bearer {token}"},
        timeout=10,
    )
    if resp.status_code != 200:
        print(f"::warning::get_node failed for {node_id}: HTTP {resp.status_code}")
        return
    data = resp.json()
    if data.get("code", 0) != 0:
        print(f"::warning::get_node code != 0 for {node_id}: {data}")
        return
    document_id = (data.get("data") or {}).get("node", {}).get("obj_token")
    if not document_id:
        print(f"::warning::no document_id resolved for wiki node {node_id}")
        return

    # Lark Docx update via "raw_content" endpoint
    resp = requests.patch(
        f"{LARK_API_BASE}/open-apis/docx/v1/documents/{document_id}/raw_content",
        json={"content": body, "title": title},
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json; charset=utf-8",
        },
        timeout=15,
    )
    if resp.status_code != 200:
        print(f"::error::raw_content update failed for {document_id}: HTTP {resp.status_code} body={resp.text[:200]}")
        return
    data = resp.json()
    if data.get("code", 0) != 0:
        print(f"::error::raw_content code != 0 for {document_id}: {data}")
        return
    print(f"::notice::synced {title} ({len(body)}B) → wiki node {node_id}")


def main() -> int:
    config = load_config()
    space_id = config.get("space_id") or ""
    banner_template = config.get("banner") or ""
    mappings = config.get("mappings") or []

    if not mappings:
        print("::warning::no mappings configured in .feishu-sync.yml — nothing to sync")
        return 0
    if not space_id:
        print("::warning::space_id not configured in .feishu-sync.yml — only DRY-RUN behavior")

    dry_run = (os.environ.get("DRY_RUN") or "false").lower() == "true"
    commit_sha = os.environ.get("COMMIT_SHA") or "unknown"
    repo_url = os.environ.get("REPO_URL") or ""

    app_id = os.environ.get("LARK_APP_ID") or ""
    app_secret = os.environ.get("LARK_APP_SECRET") or ""

    if not dry_run and (not app_id or not app_secret):
        sys.exit("::error::LARK_APP_ID / LARK_APP_SECRET missing (workflow should have caught this earlier)")

    token = "" if dry_run else get_tenant_access_token(app_id, app_secret)

    pushed = 0
    skipped = 0
    for entry in mappings:
        fname = entry.get("file") or ""
        title = entry.get("title") or fname
        node_id = entry.get("node_id") or ""

        path = DOCS_DIR / fname
        if not path.exists():
            print(f"::warning::{path} not found, skipping")
            skipped += 1
            continue
        if not node_id:
            print(f"::warning::{fname}: node_id empty, skipping (configure in .feishu-sync.yml to enable)")
            skipped += 1
            continue

        raw = path.read_text(encoding="utf-8")
        body = strip_frontmatter(raw)
        banner = format_banner(banner_template, commit_sha, repo_url, str(path).replace("\\", "/")) if banner_template else ""
        full = banner + body

        update_wiki_node_content(token, node_id, title, full, dry_run)
        pushed += 1

    print(f"\nSummary: pushed={pushed} skipped={skipped} dry_run={dry_run}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
