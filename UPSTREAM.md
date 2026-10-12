# Upstream contributions

**English** | [简体中文](UPSTREAM.zh-Hans.md)

DRYAS will continue proposing suitable fixes and general improvements to
[Open-MBEE/OpenSysML](https://github.com/Open-MBEE/OpenSysML). Upstream maintainers
decide acceptance, revisions and release timing. Experimental or product-specific
interfaces may remain downstream. This policy does not promise that every change
will be submitted or accepted, and does not authorize automated posting of PRs.

## Verified contribution record

| Change | Upstream record | Verified status |
| --- | --- | --- |
| Imported-name visibility and clash handling | [PR #981](https://github.com/Open-MBEE/OpenSysML/pull/981) | Open, not merged at this review. |
| Part typing, base-reachable cycles and diagnostic cascade handling | No submission verified in this review | Maintained downstream. |
| Native model-query exports | No submission verified in this review | Experimental downstream stdio APIs; not an upstream SDK contract. |
| Imported-name completion and documentation-aware LSP features | No submission verified in this review | Maintained downstream. |

Live PR state may change. A submitted PR is not an accepted change; a merged PR
is not necessarily released; a released change is not automatically tested with
Systrace. Record the upstream issue/PR, upstream release if known, downstream
revision and test evidence when updating status. Do not invent links or infer
acceptance from a successful downstream test.

Keep contribution branches separate from downstream integration work. Send focused
patches with reproductions, tests and specification evidence where applicable.
Reconcile upstream changes deliberately; a merged upstream fix may still need
verification against downstream changes before replacing its implementation.

## License and attribution maintenance

Preserve LICENSE, applicable copyright and attribution notices, and applicable
upstream NOTICE content. Modified upstream files must carry prominent modification
notices. A Git commit or SPDX identifier alone is not a replacement. Keep a short
file-level description and maintain user-facing changes separately. Preserve all
original notices when adding DRYAS attribution; do not claim unchanged upstream code.

For formats without comments, choose a format-supported notice mechanism and
verify consumers still work. The constraints golden fixture carries a notice that
is included in its exact test comparison and regeneration; diagnostic expectations
are not relaxed or filtered to accommodate the notice.

Keep new maintenance documentation in English and Simplified Chinese (zh-Hans).
Legal texts retain their original wording. Upstream legacy documents are not
implicitly claimed to have a complete translation.

## Distribution checklist

- Include LICENSE and applicable NOTICE files in source and binary distributions.
- Record the exact source revision, downstream changes and supported build identifier.
- Inventory shipped components and dependencies; preserve their separate licenses.
- The OMG-derived bundled library is EPL-2.0; OpenSysML library extensions are
  Apache-2.0. Preserve the library NOTICE and license and satisfy applicable source
  availability requirements when distributing it. Do not describe an engine bundle
  as entirely Apache-2.0 without reviewing its contents.
- Check the actual release archive, installer or VSIX; repository files alone do
  not prove that notices reached recipients. Contributing a PR does not replace
  redistribution obligations.
- Do not imply upstream endorsement or treat the license as a general trademark grant.

This records maintenance requirements; it is not a claim that every existing
packaging target has completed a release license inventory.

## Fork CI

The DRYAS Engine CI workflow runs on pull requests targeting main, pushes to main,
and manual dispatch. Branch pushes are checked through their PR to avoid duplicate runs. Windows and Linux both build
all Go packages, check formatting and maintenance documents, run the native downstream
regressions without skipped cases, and build and launch the three engine executables.
Linux runs full-module vet; Windows vets the integration packages because unrelated
upstream FIFO tests do not compile there. The CI development version includes its
source revision and is not a supported Systrace release identifier.

Run `python scripts/dryas-ci.py` locally with Go on PATH; add `--build` to build and
smoke-test executables. The inherited PR workflow remains unchanged and covers the
broader upstream suite. A successful downstream gate does not mean that full suite
passed, and does not close known upstream-platform or self-model failures. This
workflow does not publish releases or distribute engine binaries.
