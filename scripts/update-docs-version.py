#!/usr/bin/env python3
"""
Updates the release version in docs/index.html.
Can be executed locally or as part of GitHub Actions CI/CD workflows.
"""

import argparse
import os
import re
import sys


def normalize_version(tag: str) -> str:
    tag = tag.strip()
    if not tag.startswith("v"):
        tag = f"v{tag}"
    return tag


def update_docs_version(file_path: str, tag: str) -> bool:
    if not os.path.exists(file_path):
        print(f"Error: {file_path} not found.", file=sys.stderr)
        return False

    version = normalize_version(tag)
    print(f"Updating {file_path} to version {version}...")

    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    original = content

    # 1. Update elements with class containing 'release-version'
    # e.g., <span class="release-version">v1.0.0</span>
    content = re.sub(
        r'(<[^>]*class=["\'][^"\']*\brelease-version\b[^"\']*["\'][^>]*>)[^<]*(</[^>]+>)',
        rf"\g<1>{version}\g<2>",
        content,
        flags=re.IGNORECASE
    )

    # 2. Update release links if specific tag was previously linked
    # e.g., href="https://github.com/smford/matrix-rain/releases/tag/v1.0.0"
    content = re.sub(
        r'(href=["\']https://github\.com/smford/matrix-rain/releases/tag/)v?[0-9]+\.[0-9]+\.[0-9]+[^"\']*(["\'])',
        rf"\g<1>{version}\g<2>",
        content,
        flags=re.IGNORECASE
    )

    # 3. Fallback: match legacy un-spanned badge `<div class="badge">v1.0.0`
    content = re.sub(
        r'(<div\s+class=["\']badge["\']>\s*)v[0-9]+\.[0-9]+\.[0-9]+',
        rf"\g<1>{version}",
        content,
        flags=re.IGNORECASE
    )

    if content != original:
        with open(file_path, "w", encoding="utf-8") as f:
            f.write(content)
        print(f"Successfully updated {file_path} with {version}.")
        return True
    else:
        print(f"No changes made; {file_path} is already at {version}.")
        return False


def main():
    parser = argparse.ArgumentParser(description="Update documentation with latest release version.")
    parser.add_argument("tag", help="Release version tag (e.g. v1.0.1 or 1.0.1)")
    parser.add_argument(
        "--file", "-f",
        default="docs/index.html",
        help="Path to HTML file to update (default: docs/index.html)"
    )
    args = parser.parse_args()

    update_docs_version(args.file, args.tag)


if __name__ == "__main__":
    main()
