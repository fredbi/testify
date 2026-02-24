#!/usr/bin/env python3
"""Godoc spell-check agent.

Drives the go-fred-mcp godoc tool programmatically, using Claude haiku
for judgment calls (real issue vs false positive, suggesting fixes).

Usage:
    python .claude/agents/godoc_check.py [OPTIONS]

Options:
    --package PATTERN    Package pattern (default: ./...)
    --fix                Apply fixes (default: report only / dry-run)
    --update-wordlist    Add false positives to wordlist without code fixes
    --report-only        Skip AI classification, just print raw findings
    --model MODEL        Model for judgment calls (default: claude-haiku-4-5-20251001)
    --workspace DIR      Workspace root (default: cwd)
    --wordlist PATH      Path to wordlist file (default: .github/wordlist.txt)
    --verbose            Print detailed progress

Requires:
    pip install mcp anthropic
    ANTHROPIC_API_KEY env var (not needed with --report-only)
    go-fred-mcp binary on PATH
"""

from __future__ import annotations

import argparse
import asyncio
import json
import os
import re
import sys
from contextlib import AsyncExitStack
from pathlib import Path

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client


# ---------------------------------------------------------------------------
# MCP helpers
# ---------------------------------------------------------------------------

async def connect_mcp(
    workspace: str,
    exit_stack: AsyncExitStack,
) -> ClientSession:
    """Connect to go-fred-mcp via stdio and return an initialised session."""
    server_params = StdioServerParameters(
        command="go-fred-mcp",
        args=["--workspace", workspace],
        env={"MCP_GOPLS_LOG_LEVEL": "info"},
    )
    transport = await exit_stack.enter_async_context(
        stdio_client(server_params),
    )
    read_stream, write_stream = transport
    session = await exit_stack.enter_async_context(
        ClientSession(read_stream, write_stream),
    )
    await session.initialize()
    return session


async def call_godoc(
    session: ClientSession,
    *,
    package_pattern: str = "./...",
    analyzers: str = "hunspell",
    filters: str = "godoc-filter",
    issues_only: bool = True,
    custom_words: str = "",
    symbol_filter: str | None = None,
) -> dict:
    """Call the godoc MCP tool and return parsed JSON result."""
    args: dict = {
        "package_pattern": package_pattern,
        "analyzers": analyzers,
        "filters": filters,
        "issues_only": issues_only,
        "custom_words": custom_words,
    }
    if symbol_filter is not None:
        args["symbol_filter"] = symbol_filter

    result = await session.call_tool("godoc", arguments=args)
    text = result.content[0].text if result.content else "{}"
    return json.loads(text)


async def call_update_godoc(
    session: ClientSession,
    updates: dict,
) -> str:
    """Call update_godoc and return the summary text."""
    result = await session.call_tool(
        "update_godoc",
        arguments={"updates": json.dumps(updates)},
    )
    return result.content[0].text if result.content else ""


# ---------------------------------------------------------------------------
# Wordlist I/O
# ---------------------------------------------------------------------------

def read_wordlist(path: Path) -> str:
    """Read the wordlist file and return its contents."""
    if not path.exists():
        return ""
    return path.read_text()


def update_wordlist(path: Path, new_words: list[str]) -> None:
    """Add new words to the wordlist, preserving sort convention.

    Convention: uppercase words first (sorted), then lowercase (sorted).
    """
    existing = set()
    if path.exists():
        existing = {w for w in path.read_text().splitlines() if w.strip()}

    combined = existing | set(new_words)

    upper = sorted(w for w in combined if w and w[0].isupper())
    lower = sorted(w for w in combined if w and not w[0].isupper())

    path.write_text("\n".join(upper + lower) + "\n")


# ---------------------------------------------------------------------------
# Issue extraction
# ---------------------------------------------------------------------------


class Issue:
    """A single spell-check issue for a symbol."""

    def __init__(self, symbol: str, word: str, suggestion: str, raw_message: str):
        self.symbol = symbol
        self.word = word
        self.suggestion = suggestion
        self.raw_message = raw_message
        self.classification: str | None = None  # "REAL" or "FALSE_POSITIVE"
        self.reason: str = ""

    def __repr__(self) -> str:
        return f"Issue({self.symbol}: {self.word!r} -> {self.suggestion!r})"


def extract_issues(godoc_result: dict) -> list[Issue]:
    """Parse godoc triage output into Issue objects.

    The triage format (issues_only=true) is:
        {symbol: {analyzer: [{kind, message, position, suggestion}, ...]}}

    The full format (issues_only=false) is:
        {symbol: {comment, kind, file, line, ...}}
    """
    issues: list[Issue] = []
    for symbol, data in godoc_result.items():
        if not isinstance(data, dict):
            continue

        # Iterate over analyzer results (e.g., "hunspell": [...])
        for _analyzer, entries in data.items():
            if not isinstance(entries, list):
                continue
            for entry in entries:
                if not isinstance(entry, dict):
                    continue
                msg = entry.get("message", "")
                suggestion = entry.get("suggestion", "")

                # Extract flagged word from message like '"word" is misspelled'
                word = ""
                m = re.search(r'"([^"]+)"', msg)
                if m:
                    word = m.group(1)

                if word:
                    issues.append(Issue(symbol, word, suggestion, msg))
    return issues


# ---------------------------------------------------------------------------
# AI classification and fixing
# ---------------------------------------------------------------------------

CLASSIFY_PROMPT = """\
You are reviewing a hunspell spell-check finding in a Go doc comment.

Symbol: {symbol}
Flagged word: "{word}"
Hunspell suggestion: "{suggestion}"
Full message: {message}

Is this a real spelling error, or a false positive (e.g., a Go identifier,
parameter name, technical term, or valid but uncommon English word)?

Reply with exactly one line:
REAL: <brief reason>
or
FALSE_POSITIVE: <brief reason>"""

FIX_PROMPT = """\
Fix the spelling error in this Go doc comment. Change "{word}" appropriately.
The hunspell suggestion is "{suggestion}".
Keep the fix minimal -- only change what's needed. Preserve godoc formatting.
The first line must remain "// SymbolName verb..." per godoc conventions.

Current comment:
{comment}

Return ONLY the corrected comment text, nothing else."""


def classify_response(text: str) -> tuple[str, str]:
    """Parse a classification response into (label, reason)."""
    text = text.strip()
    if text.upper().startswith("REAL"):
        reason = text.split(":", 1)[1].strip() if ":" in text else ""
        return "REAL", reason
    if text.upper().startswith("FALSE_POSITIVE"):
        reason = text.split(":", 1)[1].strip() if ":" in text else ""
        return "FALSE_POSITIVE", reason
    # Fallback: if the response contains "false positive" anywhere, treat as FP
    if "false positive" in text.lower():
        return "FALSE_POSITIVE", text
    return "REAL", text


def _read_claude_code_token() -> str | None:
    """Read the OAuth access token from Claude Code's local credentials.

    Claude Code stores its OAuth token in ~/.claude/.credentials.json:
        {"claudeAiOauth": {"accessToken": "sk-ant-oat01-...", ...}}
    """
    creds_path = Path.home() / ".claude" / ".credentials.json"
    if not creds_path.exists():
        return None
    try:
        data = json.loads(creds_path.read_text())
        token = data.get("claudeAiOauth", {}).get("accessToken", "")
        return token if token else None
    except (json.JSONDecodeError, KeyError, OSError):
        return None


def _get_ai_client(*, verbose: bool = False):
    """Create an Anthropic async client.

    Authentication fallback chain:
        1. ANTHROPIC_API_KEY env var (standard API key)
        2. ANTHROPIC_AUTH_TOKEN env var (OAuth bearer token)
        3. Claude Code local credentials (~/.claude/.credentials.json)
        4. Error with helpful message
    """
    import anthropic  # noqa: PLC0415 — deferred import

    # 1. Standard API key
    api_key = os.environ.get("ANTHROPIC_API_KEY", "")
    if api_key:
        if verbose:
            print("  [verbose] Using ANTHROPIC_API_KEY", file=sys.stderr)
        return anthropic.AsyncAnthropic(api_key=api_key)

    # OAuth tokens (from env or Claude Code credentials) need a beta header.
    _oauth_headers = {"anthropic-beta": "oauth-2025-04-20"}

    # 2. Explicit auth token env var
    auth_token = os.environ.get("ANTHROPIC_AUTH_TOKEN", "")
    if auth_token:
        if verbose:
            print("  [verbose] Using ANTHROPIC_AUTH_TOKEN", file=sys.stderr)
        return anthropic.AsyncAnthropic(
            auth_token=auth_token,
            default_headers=_oauth_headers,
        )

    # 3. Claude Code local credentials
    cc_token = _read_claude_code_token()
    if cc_token:
        if verbose:
            print(
                "  [verbose] Using Claude Code OAuth token"
                " from ~/.claude/.credentials.json",
                file=sys.stderr,
            )
        return anthropic.AsyncAnthropic(
            auth_token=cc_token,
            default_headers=_oauth_headers,
        )

    # 4. No auth found
    print(
        "Error: No Anthropic authentication found.\n"
        "Options:\n"
        "  - Set ANTHROPIC_API_KEY env var (API key from console.anthropic.com)\n"
        "  - Set ANTHROPIC_AUTH_TOKEN env var (OAuth bearer token)\n"
        "  - Log in to Claude Code (claude login) to use its OAuth token\n"
        "  - Use --report-only to skip AI classification",
        file=sys.stderr,
    )
    sys.exit(2)


async def classify_issue(
    client,  # anthropic.AsyncAnthropic
    issue: Issue,
    model: str,
) -> None:
    """Classify a single issue using the AI model."""
    prompt = CLASSIFY_PROMPT.format(
        symbol=issue.symbol,
        word=issue.word,
        suggestion=issue.suggestion,
        message=issue.raw_message,
    )
    resp = await client.messages.create(
        model=model,
        max_tokens=100,
        messages=[{"role": "user", "content": prompt}],
    )
    text = resp.content[0].text if resp.content else ""
    issue.classification, issue.reason = classify_response(text)


async def get_fix(
    client,  # anthropic.AsyncAnthropic
    issue: Issue,
    comment: str,
    model: str,
) -> str:
    """Ask the AI model for a corrected comment."""
    prompt = FIX_PROMPT.format(
        word=issue.word,
        suggestion=issue.suggestion,
        comment=comment,
    )
    resp = await client.messages.create(
        model=model,
        max_tokens=2000,
        messages=[{"role": "user", "content": prompt}],
    )
    return resp.content[0].text.strip() if resp.content else comment


# ---------------------------------------------------------------------------
# Report-only mode (no AI needed)
# ---------------------------------------------------------------------------

def print_raw_report(package: str, issues: list[Issue]) -> int:
    """Print raw findings without AI classification."""
    print()
    print("=== Godoc Spell-Check Report (raw) ===")
    print(f"Package: {package}")
    print(f"Issues found: {len(issues)}")

    # Group by symbol
    by_symbol: dict[str, list[Issue]] = {}
    for iss in issues:
        by_symbol.setdefault(iss.symbol, []).append(iss)

    print()
    for symbol, sym_issues in sorted(by_symbol.items()):
        print(f"  {symbol}:")
        for iss in sym_issues:
            sug = f" (suggestion: {iss.suggestion})" if iss.suggestion else ""
            print(f'    "{iss.word}"{sug}')

    print()
    print("Use without --report-only to classify with AI,")
    print("or use --fix to apply corrections.")
    return 1


# ---------------------------------------------------------------------------
# Main workflow
# ---------------------------------------------------------------------------

async def run(args: argparse.Namespace) -> int:
    """Execute the godoc spell-check workflow."""
    workspace = str(Path(args.workspace).resolve())
    wordlist_path = Path(workspace) / args.wordlist
    verbose = args.verbose

    def log(msg: str) -> None:
        if verbose:
            print(f"  [verbose] {msg}", file=sys.stderr)

    # --- Setup ---
    exit_stack = AsyncExitStack()
    ai = None
    if not args.report_only:
        ai = _get_ai_client(verbose=verbose)

    try:
        print(f"Connecting to go-fred-mcp (workspace: {workspace})...")
        session = await connect_mcp(workspace, exit_stack)
        log("MCP session initialised")

        # --- Step 1: Read wordlist ---
        custom_words = read_wordlist(wordlist_path)
        word_count = len([w for w in custom_words.splitlines() if w.strip()])
        log(f"Loaded {word_count} custom words from {wordlist_path}")

        # --- Step 2: Triage ---
        print(f"Running spell-check triage on {args.package}...")
        triage = await call_godoc(
            session,
            package_pattern=args.package,
            issues_only=True,
            custom_words=custom_words,
        )

        if not triage:
            print("\n=== Godoc Spell-Check Report ===")
            print(f"Package: {args.package}")
            print("No issues found. Documentation is clean!")
            return 0

        issues = extract_issues(triage)
        if not issues:
            print("\n=== Godoc Spell-Check Report ===")
            print(f"Package: {args.package}")
            print("No issues found. Documentation is clean!")
            return 0

        print(f"Found {len(issues)} issue(s) across {len(triage)} symbol(s)")

        # --- Report-only mode ---
        if args.report_only:
            return print_raw_report(args.package, issues)

        # --- Step 3: Classify each issue ---
        assert ai is not None
        print("Classifying issues...")
        for i, issue in enumerate(issues, 1):
            log(f"Classifying {i}/{len(issues)}: {issue.symbol} -> {issue.word!r}")
            await classify_issue(ai, issue, args.model)
            log(f"  -> {issue.classification}: {issue.reason}")

        real_issues = [iss for iss in issues if iss.classification == "REAL"]
        false_positives = [iss for iss in issues if iss.classification == "FALSE_POSITIVE"]

        # --- Step 4: Report ---
        print()
        print("=== Godoc Spell-Check Report ===")
        print(f"Package: {args.package}")
        print(f"Issues found: {len(issues)}")
        print(f"  Real errors: {len(real_issues)}")
        print(f"  False positives: {len(false_positives)}")

        if real_issues:
            print()
            print("Real errors:")
            for iss in real_issues:
                sug = f' -> suggest "{iss.suggestion}"' if iss.suggestion else ""
                print(f'  {iss.symbol}: "{iss.word}"{sug}')
                if verbose:
                    print(f"    reason: {iss.reason}")

        if false_positives:
            print()
            fp_words = sorted({iss.word for iss in false_positives})
            print(f"False positives ({len(fp_words)} unique words):")
            for w in fp_words:
                print(f"  {w}")

        # --- Step 5: Update wordlist only (no code fixes) ---
        if args.update_wordlist and not args.fix:
            if false_positives:
                fp_words = sorted({iss.word for iss in false_positives})
                print(f"\nAdding {len(fp_words)} word(s) to wordlist: {wordlist_path}")
                update_wordlist(wordlist_path, fp_words)
            if real_issues:
                print()
                print("Run with --fix to also apply code corrections.")
            return 1 if real_issues else 0

        # --- Dry-run: report only ---
        if not args.fix:
            if real_issues:
                print()
                print("Run with --fix to apply corrections.")
            return 1 if real_issues else 0

        # Apply fixes for real issues
        if real_issues:
            print()
            print("Applying fixes...")

            # Group issues by symbol to avoid duplicate fetches
            symbols_to_fix: dict[str, list[Issue]] = {}
            for iss in real_issues:
                symbols_to_fix.setdefault(iss.symbol, []).append(iss)

            fixed_count = 0
            for symbol, sym_issues in symbols_to_fix.items():
                log(f"Fetching full comment for {symbol}")

                # Extract just the symbol name (after last dot) for the filter
                sym_name = symbol.rsplit(".", 1)[-1] if "." in symbol else symbol
                full_data = await call_godoc(
                    session,
                    package_pattern=args.package,
                    symbol_filter=f"^{re.escape(sym_name)}$",
                    issues_only=False,
                    custom_words=custom_words,
                )

                if symbol not in full_data:
                    print(f"  Warning: could not fetch comment for {symbol}, skipping")
                    continue

                comment = full_data[symbol].get("comment", "")
                if not comment:
                    print(f"  Warning: empty comment for {symbol}, skipping")
                    continue

                # Apply fixes for each issue in this symbol
                current_comment = comment
                for iss in sym_issues:
                    log(f"  Fixing {iss.word!r} in {symbol}")
                    current_comment = await get_fix(
                        ai, iss, current_comment, args.model,
                    )

                # Apply the fix via update_godoc
                result = await call_update_godoc(
                    session,
                    {symbol: {"comment": current_comment}},
                )
                log(f"  update_godoc result: {result}")
                fixed_count += 1
                print(f"  Fixed: {symbol}")

            print(f"\n{fixed_count} symbol(s) fixed.")

        # Update wordlist with false positives
        if false_positives:
            fp_words = sorted({iss.word for iss in false_positives})
            print(f"\nAdding {len(fp_words)} word(s) to wordlist: {wordlist_path}")
            update_wordlist(wordlist_path, fp_words)
            # Re-read for verification pass
            custom_words = read_wordlist(wordlist_path)

        # --- Step 6: Verification pass ---
        if real_issues:
            print("\nVerifying fixes...")
            verify = await call_godoc(
                session,
                package_pattern=args.package,
                issues_only=True,
                custom_words=custom_words,
            )
            remaining = extract_issues(verify) if verify else []
            print(f"Before: {len(issues)} issues")
            print(f"After:  {len(remaining)} issues")
            if remaining:
                print("\nRemaining issues:")
                for iss in remaining:
                    print(f'  {iss.symbol}: "{iss.word}"')
                return 1

        print("\nAll clean!")
        return 0

    finally:
        await exit_stack.aclose()


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="Godoc spell-check agent using go-fred-mcp + Claude AI",
    )
    parser.add_argument(
        "--package",
        default="./...",
        help="Package pattern (default: ./...)",
    )
    parser.add_argument(
        "--fix",
        action="store_true",
        help="Apply fixes (default: report only)",
    )
    parser.add_argument(
        "--update-wordlist",
        action="store_true",
        help="Add false positives to wordlist without applying code fixes",
    )
    parser.add_argument(
        "--report-only",
        action="store_true",
        help="Skip AI classification, just print raw findings",
    )
    parser.add_argument(
        "--model",
        default="claude-haiku-4-5-20251001",
        help="Model for judgment calls (default: claude-haiku-4-5-20251001)",
    )
    parser.add_argument(
        "--workspace",
        default=".",
        help="Workspace root (default: cwd)",
    )
    parser.add_argument(
        "--wordlist",
        default=".github/wordlist.txt",
        help="Path to wordlist file relative to workspace (default: .github/wordlist.txt)",
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Print detailed progress",
    )
    args = parser.parse_args()

    if args.report_only and (args.fix or args.update_wordlist):
        parser.error("--report-only is mutually exclusive with --fix and --update-wordlist")

    sys.exit(asyncio.run(run(args)))


if __name__ == "__main__":
    main()
