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
