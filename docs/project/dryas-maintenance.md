# DRYAS maintenance branch

This branch starts at upstream v0.9.2 and carries separately committed semantic
fixes plus experimental read-only exports. It does not replace upstream develop.
The upstream module path, licenses and notices are retained.

## Branches

- `main` is the DRYAS default and integration branch. It starts from the existing
  `codex/dryas-0.9.2` maintenance history, including the reviewed import fixes,
  native query exports and imported-name completion. New DRYAS work branches from
  `main` and targets `main`.
- `develop` follows the upstream development line. Upstream updates are reviewed
  and integrated into `main` deliberately.
- `codex/fix-import-clashes` remains dedicated to upstream PR #981.
- `codex/dryas-0.9.2` is retained as the previous maintenance branch. Ongoing
  product development moves to `main`.

Changing the default branch does not upgrade the installed Tracemgr engine or
publish a binary release. Product engine versions remain selected separately.

The stdio-only methods DryasDescribeInherited, DryasFindBySpecialization and
DryasDescribeProvenance reuse GetSymbolRequest and QueryResponse envelopes.
They expose existing native effective members, feature types, transitive
specialization, implicit relations and annotation-site metadata. They do not
implement a second semantic resolver. Missing, ambiguous, unnamed or provisional
target identities are rejected where
the experimental projection cannot represent them. Inherited-feature result IDs
still need collision validation on malformed duplicate declarations before this
interface can claim a stable identity contract.

These methods are a private integration experiment, not a stable upstream API.
They are not registered as gRPC/Connect methods and have no generated SDK contract.
An upstream API proposal must define public schemas, client bindings and compatibility
tests before this protocol can be advertised as a general-purpose public service.

DRYAS adapters, evaluation cases, engineering rules and UI stay in the DRYAS repository.
`dryas-labs/OpenSysML` is maintained as the product engine in its own right.
The existing upstream submission is [Open-MBEE PR #981](https://github.com/Open-MBEE/OpenSysML/pull/981).
Other fixes remain in this fork; new upstream PRs are outside the current scope.
Upstream work is integrated selectively and tested with the downstream changes.
A successful development benchmark does not establish full standard conformance
or release qualification.

## Build

Use the Go version required by go.mod. For a reproducible development binary:

```sh
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.3" -o bin/sysml ./cmd/sysml
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.3" -o bin/sysml-grpc ./cmd/sysml-grpc
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.3" -o bin/sysml-lsp ./cmd/sysml-lsp
```

This is a development version label, not an upstream release or a published tag.
See CONTRIBUTING.md for the full build, vet, formatting, test, corpus and CI gates.
Windows test failures and absent corpora must be reported rather than hidden.

## Unqualified completion

The development build also routes ordinary completion through native visible-name
enumeration and name lookup. Imported definitions, aliases and short names are
offered without exposing private members or hidden clashing imports. Completion
candidates remain a general visible-name list; this change does not claim exhaustive
metatype filtering for every grammar position.

## Import-review integration

The import review changes from commit `1f7f7c9935ca5a1982cbc9af773f811bd9b760a3`
are backported from PR #981. They prevent root imports from re-entering through
global lookup, keep hidden re-exports from poisoning independent valid imports,
and consider inherited imports together while retaining callable-overload handling.
Existing part-typing, cycle, cascade-diagnostic, query-export and completion fixes
remain in this branch.

The pre-integration commit `4ede7ba080fbc7dd6f1e11dd4b3304abc2bb7612` and the
integrated implementation produce identical diagnostic messages over the pinned
2026-08 pilot corpora. The fork-specific expectation record now reflects those
already reviewed downstream fixes; see [the corpus note](imported-name-clashes.md).

This source prepares the next development engine, `v0.9.2-dryas.3`. Existing
Tracemgr 0.1.2 packages remain tied to `v0.9.2-dryas.2` until a separate product
engine update is tested and packaged. Updating this branch does not silently
replace an installed engine or alter historical benchmark results.
