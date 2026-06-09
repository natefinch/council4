# Hybrid declarative card implementations with code escape hatch

Card implementations are declarative compositions of effect primitives (damage, destroy, draw, create token, etc.), not generated runtime mutation code. Card Generation parses Oracle text and compiles recognized semantics into validated declarative data. Cards too complex to express declaratively get hand-written Go implementations behind the same interface. We chose this because declarative data is cheaper to clone for future MCTS search, can be validated before source emission, and keeps priority, targeting, and state mutation inside the rules engine. The tradeoff is that cards with truly unique mechanics (Mindslaver, Hive Mind, etc.) need hand-written code, but the declarative system covers the common case.

## Considered Options

- **Pure declarative**: Every card expressed as data. Rejected because some MTG cards have mechanics that can't be reduced to parameter combinations without an impossibly large primitive set.
- **Generated runtime code**: Generate a Go function per card. Rejected because generated code touching game state mutation, priority, and targeting is error-prone and hard to validate — a subtle bug in one card corrupts the entire simulation.
- **Hybrid (chosen)**: Compile Oracle text to validated declarative data by default, with a hand-written code escape hatch for outliers. This gets the reliability and performance benefits of declarative data while retaining full expressiveness.
