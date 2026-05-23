# Hybrid declarative card implementations with code escape hatch

Card implementations are declarative compositions of effect primitives (damage, destroy, draw, create token, etc.), not generated Go code. Cards too complex to express declaratively get hand-written Go implementations behind the same interface. We chose this over pure code generation because: (1) declarative data is cheaper to clone for future MCTS search — no function pointers or closures in game state; (2) LLM generation of structured data is far more reliable and validatable than generating arbitrary Go code that must handle priority, targeting, and state mutation correctly; (3) Forge and SabberStone both converged on declarative/scripted card definitions after years of development. The tradeoff is that ~5% of cards with truly unique mechanics (Mindslaver, Hive Mind, etc.) need hand-written code, but the declarative system covers the vast majority of Commander-playable cards.

## Considered Options

- **Pure declarative**: Every card expressed as data. Rejected because some MTG cards have mechanics that can't be reduced to parameter combinations without an impossibly large primitive set.
- **Pure code generation**: LLM generates a Go function per card. Rejected because generated code touching game state mutation, priority, and targeting is error-prone and hard to validate — a subtle bug in one card corrupts the entire simulation.
- **Hybrid (chosen)**: Declarative by default, code escape hatch for outliers. Gets the reliability and performance benefits of declarative for the common case while keeping full expressiveness available.
