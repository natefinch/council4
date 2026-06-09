# Oracle Compiler Expansion

This checklist tracks the next major expansions of executable Oracle-text
compilation. Check off a step only after its implementation, full-corpus delta
review, Opus 4.8 review, repository validation, and separate commit are
complete.

**Current corpus support: 1,838 / 37,628 cards**

## Progress

- [x] **1. Lower parameterized Equip using `EquipActivatedAbility`**
  - Planning signal: approximately 430 blockers.
  - Completed in `2992ba8` (`Lower parameterized Equip abilities`).
- [x] **2. Add complete Enchant/Aura and Protection templates and lowering**
  - Planning signal: 1,083 blockers.
  - Completed in `a41f015` (`Lower Enchant and Protection keywords`).
- [x] **3. Build composable multi-effect sequence lowering**
  - Planning signal: foundation for thousands of cards.
  - Completed in `8df2720` (`Lower ordered effect sequences`).
- [x] **4. Expand supported spells using sequence lowering**
  - Planning signal: 4,625 blockers.
  - Completed with Surveil, Investigate, Proliferate, Regenerate, Fight, and
    reminder-aware exact lowering.
- [ ] **5. Lower ordinary activated abilities with mana/tap costs**
  - Planning signal: 5,947 ability blockers plus 690 cost issues.
- [ ] **6. Expand enter/dies triggers**
  - Include multiple effects, optional effects, and broader subjects.
  - Planning signal: 9,599 trigger blockers plus 632 related issues.
- [ ] **7. Lower conditional enters-tapped replacements**
  - Planning signal: 1,247 blockers.
- [ ] **8. Add common static effects**
  - Include buffs, restrictions, cost changes, and mixed keyword text.
  - Planning signal: 3,609 blockers.
- [ ] **9. Lower loyalty and modal abilities through shared effect lowering**
  - Planning signal: 695 loyalty blockers plus 335 modal blockers.
- [ ] **10. Add layouts, then frequency-driven parser mechanics**
  - Add Adventure, split, and prepare layouts first.
  - Planning signal: 344 playable layouts and 10,288 Oracle constructs.

## Completion Gate

Each numbered step requires:

1. Test-driven vertical slices with exact accepted and rejected wording.
2. A full-corpus generation run and inspection of every newly supported card.
3. An independent Opus 4.8 review and resolution of material findings.
4. Repository tests, vet, build, lint, and generated-package validation.
5. Updated support counts in this file and `cardgen/oracle/README.md`.
6. A separate commit before work begins on the next numbered step.
