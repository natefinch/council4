# Non-generated card implementations

This package owns authored Go definitions and versioned synchronous programs.
Add completed `cardset.Entry` values to `Cards`. The default registry includes
them explicitly; `genregistry` and the legacy compiler write only generated
letter packages and do not rewrite this directory.

Keep complete costs, targets, triggers and static/replacement declarations
inspectable. Use the [program context](../../cardprogram/README.md) for shared
Engine operations, not raw Game mutation. Preserve existing registrations; a
duplicate identity is an error unless a unique legacy replacement is explicitly
declared. Infrastructure and synthetic tests do not themselves promote a card.

Blood Moon is an authored, program-free static declaration: battlefield
`Selection` excludes the Basic supertype and `LayerType` applies
`SetBasicLandTypes`. This shared CR 305.7 SET removes old land types and
intrinsic/copiable abilities, unlike Urborg's retained ADD operation. It preserves
other type families, supertypes, animation and external grants. Type-layer
source-erasure dependencies are discovered from declarations, never card names.
CR 613.6 means layer-6 ability loss does **not** stop a Blood Moon effect already
begun in layer 4; overwriting its ability in layer 1 or removing the source does.
Acceptance uses the Oracle corpus cached October 13, 2026 and the official
September 25, 2026 Comprehensive Rules. Copied Aura restrictions (#3227) and the
held-out Humility/Opalescence composition challenges remain separate work.
