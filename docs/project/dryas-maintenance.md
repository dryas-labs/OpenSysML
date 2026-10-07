# DRYAS maintenance branch

This branch starts at upstream v0.9.2 and carries separately committed semantic
fixes plus experimental read-only exports. It does not replace upstream develop.
The upstream module path, licenses and notices are retained.

The stdio-only methods DryasDescribeInherited, DryasFindBySpecialization and
DryasDescribeProvenance reuse GetSymbolRequest and QueryResponse envelopes.
They expose existing native effective members, feature types, transitive
specialization, implicit relations and annotation-site metadata. They do not
implement a second semantic resolver. Missing, ambiguous, unnamed or provisional
identities fail explicitly where the experimental projection cannot represent them.

These methods are a private integration experiment, not a stable upstream API.
They are not registered as gRPC/Connect methods and have no generated SDK contract.
An upstream API proposal must define public schemas, client bindings and compatibility
tests before this protocol can be advertised as a general-purpose public service.

DRYAS adapters, evaluation cases, engineering rules and UI stay in the DRYAS repository.
General semantic fixes should be ported individually to current upstream develop,
reproduced and tested there, then proposed separately. A successful development
benchmark does not establish full standard conformance or release qualification.

## Build

Use the Go version required by go.mod. For a reproducible development binary:

```sh
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.1" -o bin/sysml ./cmd/sysml
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.1" -o bin/sysml-grpc ./cmd/sysml-grpc
go build -trimpath -ldflags="-X main.Version=v0.9.2-dryas.1" -o bin/sysml-lsp ./cmd/sysml-lsp
```

This is a development version label, not an upstream release or a published tag.
See CONTRIBUTING.md for the full build, vet, formatting, test, corpus and CI gates.
Windows test failures and absent corpora must be reported rather than hidden.
