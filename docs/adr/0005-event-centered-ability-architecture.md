# Event-centered ability architecture

Phase 9 builds triggered abilities, replacement/prevention effects, and reporting around typed game events emitted by the rules engine as state changes happen. Event data types live with the pure data model in `game/`, while event emission, trigger detection, and replacement/prevention behavior live in `rules/`; this preserves the existing `game` data / `rules` behavior split while giving card definitions a structured event vocabulary.

## Considered Options

- **Logs as events**: Rejected because report logs omit rules details and would force gameplay behavior to depend on analytics shape.
- **Direct callbacks inside each mutation**: Rejected because triggers and replacement effects would become scattered across unrelated rule helpers.
- **Typed game event data in `game/`, behavior in `rules/` (chosen)**: Keeps event emission close to state mutation while providing one shared vocabulary for trigger specs, replacement/prevention effects, and derived logs.
