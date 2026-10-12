#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
# Added by DRYAS maintainers for focused downstream CI, not the full upstream suite.
"""Run native downstream regressions, rejecting missing tests and skipped cases."""

import argparse
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parent.parent
# Each pattern includes an entire behavior group, including its subtests.
TARGETS = {
    "./internal/check/passes": "Test(PassesGolden|ImportedNameClash|Resolved(Typing|Expression)Survives|CyclesNeedAnActualBasePath|PartUsagePartDefinition|ConstraintDirectSpecializationCycle|ConstraintTransitiveSpecializationCycle|ConstraintNestedMemberSpecializationCycle)",
    "./internal/frontend/lsp": "Test(CompletionVisibleImportedNames|CompletionRefreshesAfterImportEdits|CompletionThenTypingKeepsCrossFileNavigation|CompletionSuppressesCommentBodies|CompletionResumesOutsideCommentBodies|CompletionResolve|HoverIncludesOwnedDocumentation|HoverOwnedDocumentationPreservesFullTextAndOwnership|HoverReferencedOwnedDocumentationAcrossFiles|HoverIncludesCompleteBundledVoltageDocumentation)",
    "./internal/frontend/grpc": "TestDryas(DerivedNativeQueries|ImplicitNativeProvenance)",
    "./internal/frontend/stdiorpc": "TestServe",
    "./internal/semantic/resolve": "Test(RootImportClashPreservesGlobalDeclaration|HiddenReexportPreservesIndependentImport|GlobalRootImportClash)",
    "./internal/semantic/semantics": "Test(MutualSpecializationDoesNotEraseFeatureTypes|CrossReferenceRetainsUntypedPartBase)",
    "./internal/workspace/libs": "Test(EmbeddedSnapshotIsCurrent|SnapshotIndexMatchesFreshLoad)$",
    "./internal/workspace/model": "Test(ConnectorEndNamesResolve|ExprTypeCheckNoStdlibFalsePositives|ExprTypeCheckPublishedStdlibDefects|ExprTypeCheckNoExampleFalsePositives|StdlibMagneticDipoleMomentNameClash)$",
    "./internal/exec/runtime": "Test(MeasurementRefValues|MeasurementRefReport|QuantityCalculations|QuantityCalculationsReport|PointArithmetic|AdoptRebindsAnExtentWhenItsTypeNameIsShadowed)$",
    "./internal/frontend/repl": "TestPromptDoesNotSeeALoadedFilesRootImports$",
    "./tests/model": "Test(FilteredImport.*|IncrementalEqualsFresh)$",
}


def run(*args: str) -> str:
    return subprocess.check_output(args, cwd=ROOT, text=True, encoding="utf-8")


def test_package(package: str, pattern: str) -> int:
    pattern = "^" + pattern
    listing = run("go", "test", "-list", pattern, package)
    expected = {line for line in listing.splitlines() if line.startswith("Test")}
    if not expected:
        raise RuntimeError(f"No regression tests found in {package}")
    command = ["go", "test", "-json", "-count=1", "-timeout=5m", "-run", pattern, package]
    result = subprocess.run(command, cwd=ROOT, capture_output=True, text=True, encoding="utf-8")
    passed = set()
    skipped = []
    for line in result.stdout.splitlines():
        event = json.loads(line)
        name = event.get("Test")
        if event.get("Action") == "pass" and name and "/" not in name:
            passed.add(name)
        if event.get("Action") == "skip":
            skipped.append(name or package)
    if result.returncode or skipped or not expected.issubset(passed):
        print(result.stdout, end="")
        print(result.stderr, end="", file=sys.stderr)
        raise RuntimeError(f"Incomplete or failed regression run in {package}; skipped={skipped}; missing={sorted(expected - passed)}")
    print(f"{package}: {len(passed)} tests passed (including their subtests), no skips", flush=True)
    return len(passed)


def build() -> None:
    revision = run("git", "rev-parse", "--short=12", "HEAD").strip()
    version = "dev-dryas-ci-" + revision
    output = ROOT / "bin" / "dryas-ci"
    output.mkdir(parents=True, exist_ok=True)
    suffix = ".exe" if os.name == "nt" else ""
    for name in ("sysml", "sysml-lsp", "sysml-grpc"):
        executable = output / (name + suffix)
        run("go", "build", "-trimpath", "-ldflags=-X main.Version=" + version,
            "-o", str(executable), "./cmd/" + name)
        reported = run(str(executable), "-version")
        if version not in reported:
            raise RuntimeError(f"{name} did not report the CI build version")
        print(f"{name}: built and version check passed", flush=True)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--build", action="store_true", help="build and launch the three CLI executables")
    args = parser.parse_args()
    if args.build:
        build()
    else:
        total = sum(test_package(package, pattern) for package, pattern in TARGETS.items())
        print(f"Downstream regression gate: {total} tests passed; this is not the full upstream suite.")
