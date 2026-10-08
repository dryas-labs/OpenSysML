# DRYAS maintenance branch

This branch starts at upstream v0.9.2 and carries separately committed semantic
fixes plus experimental read-only exports. It does not replace upstream develop.
The upstream module path, licenses and notices are retained. See
[license and distribution requirements](../../UPSTREAM.md#license-and-attribution-maintenance)
and the [downstream change record](../../CHANGES.md).

## Branches

- `main` is the DRYAS default and integration branch. It starts from the existing
  `codex/dryas-0.9.2` maintenance history, including the reviewed import fixes,
  native query exports and imported-name completion. New DRYAS work branches from
  `main` and targets `main`.
- `develop` follows the upstream development line. Upstream updates are reviewed
  and integrated into `main` deliberately.
- `codex/fix-import-clashes` remains dedicated to upstream PR #981.

Changing the default branch does not upgrade the installed Systrace engine or
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
Suitable fixes and general improvements will continue to be proposed upstream.
Acceptance and release timing remain upstream decisions; see [contribution status](../../UPSTREAM.md).
Upstream work is integrated selectively and tested with the downstream changes.
A successful development benchmark does not establish full standard conformance
or release qualification.

## Build

Use the Go version required by go.mod. For a reproducible development binary:

```sh
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.4" -o bin/sysml ./cmd/sysml
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.4" -o bin/sysml-grpc ./cmd/sysml-grpc
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.4" -o bin/sysml-lsp ./cmd/sysml-lsp
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

This source includes the development engine changes used by Systrace with
`v0.9.2-dryas.4`. Systrace selects and tests an exact source revision independently.
Updating this branch does not replace an installed engine.

## Documentation editing

The language server uses native lexer token spans to suppress code completion
inside regular comments, documentation and notes. An unfinished block remains
documentation at EOF; completion resumes after a closing delimiter or line-note
terminator. Delimiters inside strings and unrestricted names do not open comments.

Ordinary hover combines existing leading notes with the complete documentation
directly owned by the native declaration. Named, anonymous and multiple doc
members are included; documentation belonging to nested or sibling declarations
is not collected. Library records are matched to their parsed source declaration
by their native source span. Definition locations and editor Ctrl-hover previews
are unchanged.

## Selected completion details

Clients advertising documentation resolve support receive a compact list and a
session-local handle. `completionItem/resolve` reads the selected declaration's
native source and supplies its signature, visible spelling, actual declaration
name, source filename, and complete leading/owned documentation. Aliases and short
names retain their offered insertion spelling. Standard-library documentation is
loaded on demand; a closed workspace record is parsed as an immutable snapshot,
without replacing its indexed symbols. Hover uses the same documentation helper.

Handles are retained for eight lists, contain no host paths, and expire on restart
or model-context changes. Continued typing of the same identifier is permitted;
an edited declaration, import, other document, or an expired list requires fresh
completion. Resolve changes only `documentation`; it never adds imports, changes
insertion text or infers missing physical semantics. Clients without documentation
resolve retain eager documentation for already parsed workspace declarations.
