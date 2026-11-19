package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Source review: MMQ #132, Oracle ID 414eb407-dd39-41d2-bf25-a7a77ff35cb4,
// printing revision 54f01925-7fd0-472d-91a4-3309e615f22f; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// Scryfall card and rulings APIs were retrieved 2026-11-19 from
// https://api.scryfall.com/cards/54f01925-7fd0-472d-91a4-3309e615f22f and
// https://api.scryfall.com/cards/54f01925-7fd0-472d-91a4-3309e615f22f/rulings;
// Oracle matched the pinned text and rulings returned []. Oracle: “If an
// opponent controls a Forest and you control a Swamp, you may cast this spell
// without paying its mana cost. {B}: This creature gets +1/+1 until end of turn.”
// The pinned CR edition
// 2026-09-25 was read from the official document
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf:
// §§ 205.3i, 305.6, 118.9(a-d), 601.2(b/f/h), 602.2(a-b), 613.4c, 611.2a,
// 514.2. The optional alternate cost replaces only the mana cost and is available
// with the specified controllers' land subtypes; {B} activates the self-pump,
// which modifies P/T this turn.
func TestDeepwoodLegateConditionalFreeCastAndPump(t *testing.T) {
	g, engine, legate := playLegate(t, "414eb407-dd39-41d2-bf25-a7a77ff35cb4", "Deepwood Legate", types.Swamp, types.Forest)
	if got := effectivePower(g, legate); got != 1 {
		t.Fatalf("Deepwood Legate power before activation = %d, want 1", got)
	}
	if got, ok := effectiveToughness(g, legate); !ok || got != 1 {
		t.Fatalf("Deepwood Legate toughness before activation = %d (%t), want 1", got, ok)
	}

	setMainPhasePriority(g, game.Player1)
	activate := action.ActivateAbility(legate.ObjectID, 0, nil, 0)
	if containsAction(engine.legalActions(g, game.Player1), activate) {
		t.Fatal("Deepwood Legate's {B} pump was available without black mana")
	}
	g.Players[game.Player1].ManaPool.Add(mana.B, 1)
	if !containsAction(engine.legalActions(g, game.Player1), activate) {
		t.Fatal("Deepwood Legate's {B} pump was unavailable with black mana")
	}
	if !engine.applyAction(g, game.Player1, activate) {
		t.Fatal("Deepwood Legate's {B} pump could not be activated")
	}
	if stack, ok := g.Stack.Peek(); !ok || stack.Kind != game.StackActivatedAbility {
		t.Fatalf("Deepwood Legate activation stack object = %+v, want activated ability", stack)
	}

	engine.resolveTopOfStack(g, &TurnLog{})
	if got := effectivePower(g, legate); got != 2 {
		t.Fatalf("Deepwood Legate power after activation = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, legate); !ok || got != 2 {
		t.Fatalf("Deepwood Legate toughness after activation = %d (%t), want 2", got, ok)
	}
}

// Source review: MMQ #197, Oracle ID 41b46aa5-6091-4bc4-8d17-d2c6b5fe1c5e,
// printing revision 6f0e9806-be8c-4b88-a4be-0111d1be81d9; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// Scryfall card and rulings APIs were retrieved 2026-11-19 from
// https://api.scryfall.com/cards/6f0e9806-be8c-4b88-a4be-0111d1be81d9 and
// https://api.scryfall.com/cards/6f0e9806-be8c-4b88-a4be-0111d1be81d9/rulings;
// Oracle matched the pinned text and rulings returned []. Oracle: “If an
// opponent controls a Plains and you control a Mountain, you may cast this spell
// without paying its mana cost. Haste.” The pinned CR edition 2026-09-25 was read from the official
// document https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf:
// §§ 205.3i, 305.6, 118.9(a-d), 601.2(b/f/h), 302.6, 702.10(a-b). The optional
// alternate cost doesn't change mana value; haste lets this newly cast creature
// attack this turn.
func TestKyrenLegateConditionalFreeCastAndHaste(t *testing.T) {
	g, engine, legate := playLegate(t, "41b46aa5-6091-4bc4-8d17-d2c6b5fe1c5e", "Kyren Legate", types.Mountain, types.Plains)
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{}

	attack := action.DeclareAttackers([]game.AttackDeclaration{{
		Attacker: legate.ObjectID,
		Target:   game.AttackTarget{Player: game.Player2},
	}})
	if !containsAction(legalDeclareAttackersActions(g, game.Player1), attack) {
		t.Fatal("newly cast Kyren Legate could not attack despite haste")
	}
	if !engine.applyDeclareAttackers(g, game.Player1, mustDeclareAttackersPayload(t, attack)) {
		t.Fatal("Engine rejected Kyren Legate's same-turn attack")
	}
	if len(g.Combat.Attackers) != 1 || g.Combat.Attackers[0].Attacker != legate.ObjectID {
		t.Fatalf("attackers = %+v, want Kyren Legate attacking Player2", g.Combat.Attackers)
	}
}

// Source review: MMQ #10, Oracle ID 780c9cb2-9d98-4789-87e0-cb81db477204,
// printing revision 1427a3a1-24e1-4697-b5eb-1c0a24f89e75; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// Scryfall card and rulings APIs were retrieved 2026-11-19 from
// https://api.scryfall.com/cards/1427a3a1-24e1-4697-b5eb-1c0a24f89e75 and
// https://api.scryfall.com/cards/1427a3a1-24e1-4697-b5eb-1c0a24f89e75/rulings;
// Oracle matched the pinned text and rulings returned []. Oracle: “If an
// opponent controls a Swamp and you control a Plains, you may cast this spell
// without paying its mana cost. Protection from black.” The pinned CR edition
// 2026-09-25 was read from
// the official document https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf:
// §§ 205.3i, 305.6, 118.9(a-d), 601.2(b/f/h), 702.16(a-f). The optional free
// cast uses those land-subtype conditions without changing mana value;
// protection from black makes a black spell unable to target this permanent.
func TestChoArrimLegateConditionalFreeCastAndProtectionFromBlack(t *testing.T) {
	g, engine, legate := playLegate(t, "780c9cb2-9d98-4789-87e0-cb81db477204", "Cho-Arrim Legate", types.Plains, types.Swamp)
	otherCreature := addCombatCreaturePermanent(g, game.Player3)
	blackRemovalID := addCardToHand(g, game.Player2, legateBlackRemoval())
	g.Players[game.Player2].ManaPool.Add(mana.B, 1)
	setMainPhasePriority(g, game.Player2)

	legal := engine.legalActions(g, game.Player2)
	otherTarget := action.CastSpell(blackRemovalID, []game.Target{game.PermanentTarget(otherCreature.ObjectID)}, 0, nil)
	protectedTarget := action.CastSpell(blackRemovalID, []game.Target{game.PermanentTarget(legate.ObjectID)}, 0, nil)
	if !containsAction(legal, otherTarget) {
		t.Fatal("black removal was not otherwise legal against an unprotected creature")
	}
	if containsAction(legal, protectedTarget) {
		t.Fatal("black removal could target Cho-Arrim Legate despite protection from black")
	}
}

// Source review: MMQ #100, Oracle ID d20b62fc-e248-4c84-87ac-7319b92531d4,
// printing revision db9adf84-ee7e-472b-bd96-9abf853afa83; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// Scryfall card and rulings APIs were retrieved 2026-11-19 from
// https://api.scryfall.com/cards/db9adf84-ee7e-472b-bd96-9abf853afa83 and
// https://api.scryfall.com/cards/db9adf84-ee7e-472b-bd96-9abf853afa83/rulings;
// Oracle matched the pinned text and rulings returned []. Oracle: “If an
// opponent controls a Mountain and you control an Island, you may cast this
// spell without paying its mana cost. Flying.” The pinned CR edition
// 2026-09-25 was read from the
// official document https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf:
// §§ 205.3i, 305.6, 118.9(a-d), 601.2(b/f/h), 702.9(a-b). The optional
// alternate cost doesn't change mana value and is enabled by those land
// subtypes; flying permits this creature to block a flying attacker.
func TestSaprazzanLegateConditionalFreeCastAndFlying(t *testing.T) {
	g, engine, legate := playLegate(t, "d20b62fc-e248-4c84-87ac-7319b92531d4", "Saprazzan Legate", types.Island, types.Mountain)
	attacker := addCombatCreaturePermanent(g, game.Player2, game.Flying)
	g.Turn.ActivePlayer = game.Player2
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{{
			Attacker: attacker.ObjectID,
			Target:   game.AttackTarget{Player: game.Player1},
		}},
	}

	block := action.DeclareBlockers([]game.BlockDeclaration{{
		Blocker:  legate.ObjectID,
		Blocking: attacker.ObjectID,
	}})
	if !containsAction(legalDeclareBlockersActions(g, game.Player1), block) {
		t.Fatal("Saprazzan Legate could not block a flying attacker")
	}
	if !engine.applyDeclareBlockers(g, game.Player1, mustDeclareBlockersPayload(t, block)) {
		t.Fatal("Engine rejected Saprazzan Legate blocking a flying attacker")
	}
	if len(g.Combat.Blockers) != 1 || g.Combat.Blockers[0].Blocker != legate.ObjectID {
		t.Fatalf("blockers = %+v, want Saprazzan Legate blocking", g.Combat.Blockers)
	}
}

// Source review: MMQ #266, Oracle ID f3623759-bf3a-4846-9ef7-1b732ad4355f,
// printing revision 827b9c99-87d7-493c-9dc3-0c6aa4a61b49; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// Scryfall card and rulings APIs were retrieved 2026-11-19 from
// https://api.scryfall.com/cards/827b9c99-87d7-493c-9dc3-0c6aa4a61b49 and
// https://api.scryfall.com/cards/827b9c99-87d7-493c-9dc3-0c6aa4a61b49/rulings;
// Oracle matched the pinned text and rulings returned []. Oracle: “If an
// opponent controls an Island and you control a Forest, you may cast this spell
// without paying its mana cost.” The pinned CR edition 2026-09-25 was read from the official
// document https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf:
// §§ 205.3i, 305.6, 118.9(a-d), 601.2(b/f/h). The optional alternate cost
// replaces only the printed mana cost, without changing mana value, while the
// caster controls a Forest and an opponent controls an Island.
func TestRushwoodLegateConditionalFreeCast(t *testing.T) {
	g, _, legate := playLegate(t, "f3623759-bf3a-4846-9ef7-1b732ad4355f", "Rushwood Legate", types.Forest, types.Island)
	if got := effectivePower(g, legate); got != 2 {
		t.Fatalf("Rushwood Legate power after its free cast = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, legate); !ok || got != 1 {
		t.Fatalf("Rushwood Legate toughness after its free cast = %d (%t), want 1", got, ok)
	}
}

func playLegate(t *testing.T, oracleID, name string, casterLand, opponentLand types.Sub) (*game.Game, *Engine, *game.Permanent) {
	t.Helper()
	def, err := cards.NewDefaultRegistry().LookupOracleID(oracleID)
	if err != nil {
		t.Fatalf("exact %s registry lookup failed: %v", name, err)
	}
	if def.Name != name {
		t.Fatalf("exact Oracle lookup returned %q, want %q", def.Name, name)
	}

	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, def)
	addBasicLandPermanent(g, game.Player1, casterLand).Tapped = true
	addBasicLandPermanent(g, game.Player2, opponentLand).Tapped = true
	setMainPhasePriority(g, game.Player1)

	cast := action.CastSpell(cardID, nil, 0, nil)
	if !containsAction(engine.legalActions(g, game.Player1), cast) {
		t.Fatalf("%s was not available to cast for free with its required land subtypes", name)
	}
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatalf("%s could not be cast using its conditional alternative cost", name)
	}
	if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 0 {
		t.Fatalf("%s did not use its free alternative cost: %+v", name, spell)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return g, engine, permanent
		}
	}
	t.Fatalf("resolved %s did not enter the battlefield", name)
	return nil, nil, nil
}

func legateBlackRemoval() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     "Legate Test Black Removal",
		ManaCost: opt.Val(cost.Mana{cost.B}),
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{color.Black},
		SpellAbility: opt.Val(game.Mode{
			Text: "Destroy target creature.",
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{{
				Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
			}},
		}.Ability()),
	}}
}
