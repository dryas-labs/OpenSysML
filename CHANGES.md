# DRYAS changes

**English** | [简体中文](CHANGES.zh-Hans.md)

This record describes downstream changes relative to upstream source revision
`4226ad9a31bfda0a38b82b3e8e244fb2fc9eed61`, the common ancestor of this maintenance
line and the inspected upstream develop history. It is not a comparison with the
latest upstream release and is not a new release announcement.

| Area | Downstream changes |
| --- | --- |
| Resolution | Imported-name clashes, hidden re-exports, inherited imports and root-import visibility. |
| Semantic checks | Retain kind-implied part typing, preserve base-reachable cycles and mutual types, retain resolved typing errors after unrelated unresolved names. |
| Model API | Experimental native inherited-member, specialization and provenance queries over stdio. |
| Completion | Enumerate imported names, aliases and short names through native lookup; suppress code completion in documentation/comments. |
| Documentation | Native owned-doc hover and selected completion details, including library declarations and context-bound handles. |
| Maintenance | Explicit downstream identity, modification notices, contribution records and distribution requirements. |

Existing detailed change fragments remain in changes/unreleased/; upstream
CHANGELOG.md is preserved. [UPSTREAM.md](UPSTREAM.md) records submission status.
[DRYAS maintenance](docs/project/dryas-maintenance.md) documents build behavior.
The development label v0.9.2-dryas.4 is not an upstream release or an assertion that
all downstream work has been merged into this fork's main branch.

The current notice update changes attribution and documentation. The constraints
golden reader retains exact matching while including its file-level modification
notice; no engine semantic behavior or expected diagnostics are changed.
