package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// inklingTokenDef builds the 2/1 white and black Inkling creature token with
// flying that Inkshield creates.
func inklingTokenDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:            "Inkling",
			Colors:          []color.Color{color.White, color.Black},
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Inkling},
			Power:           opt.Val(game.PT{Value: 2}),
			Toughness:       opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
		},
	}
}

// TestInkshieldPreventsCombatDamageAndCreatesTokensForAmountPrevented is the
// end-to-end Inkshield check: resolving "Prevent all combat damage that would be
// dealt to you this turn" shields the controller, combat damage to them is
// prevented, the shield tallies how much it prevented, and the delayed
// end-step payoff creates one 2/1 Inkling token per 1 combat damage
// prevented this way.
func TestInkshieldPreventsCombatDamageAndCreatesTokensForAmountPrevented(t *testing.T) {
	const prevented = 6
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	// A single stack object stands in for the resolving Inkshield spell so the
	// shield and the delayed create-token trigger share one source card, the
	// key the dynamic "damage prevented this way" amount links on.
	cardInstanceID := g.IDGen.Next()
	spell := &game.StackObject{
		Kind:       game.StackSpell,
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   cardInstanceID,
	}
	log := &TurnLog{}

	resolveInstruction(engine, g, spell, game.PreventDamage{
		Player:     game.ControllerReference(),
		All:        true,
		CombatOnly: true,
	}, log)
	resolveInstruction(engine, g, spell, game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountDamagePreventedThisWay, Multiplier: 1}),
				Source: game.TokenDef(inklingTokenDef()),
			}}}}.Ability(),
		},
	}, log)

	attacker := addColoredSourceCard(g, game.Player2, color.Red)
	startLife := g.Players[game.Player1].Life
	if dealt := dealPlayerDamage(g, attacker, 0, game.Player2, game.Player1, prevented, true); dealt != 0 {
		t.Fatalf("combat damage dealt to shielded controller = %d, want 0", dealt)
	}
	if g.Players[game.Player1].Life != startLife {
		t.Fatalf("controller life = %d, want unchanged %d", g.Players[game.Player1].Life, startLife)
	}

	var tally int
	for i := range g.PreventionShields {
		if g.PreventionShields[i].SourceID == cardInstanceID {
			tally = g.PreventionShields[i].Prevented
		}
	}
	if tally != prevented {
		t.Fatalf("shield.Prevented = %d, want %d", tally, prevented)
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("delayed triggers after end step = %d, want 0", len(g.DelayedTriggers))
	}
	if got := countTokenPermanentsNamed(g, "Inkling"); got != prevented {
		t.Fatalf("Inkling tokens created = %d, want %d (one per 1 combat damage prevented this way)", got, prevented)
	}
}

// TestInkshieldCountsCombatDamagePreventedAcrossExtraCombats proves the payoff
// is a next-end-step delayed trigger rather than an end-of-combat one. The
// "prevent all combat damage this turn" shield keeps preventing across every
// combat phase, so its Prevented tally grows in each combat. The tokens are
// created once at the end step and count the whole turn's prevented total; an
// end-of-combat trigger would instead fire after the first combat and miss the
// damage prevented in a later (extra) combat.
func TestInkshieldCountsCombatDamagePreventedAcrossExtraCombats(t *testing.T) {
	const firstCombat = 4
	const secondCombat = 3
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)

	cardInstanceID := g.IDGen.Next()
	spell := &game.StackObject{
		Kind:       game.StackSpell,
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   cardInstanceID,
	}
	log := &TurnLog{}

	resolveInstruction(engine, g, spell, game.PreventDamage{
		Player:     game.ControllerReference(),
		All:        true,
		CombatOnly: true,
	}, log)
	resolveInstruction(engine, g, spell, game.CreateDelayedTrigger{
		Trigger: game.DelayedTriggerDef{
			Timing: game.DelayedAtBeginningOfNextEndStep,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.CreateToken{
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountDamagePreventedThisWay, Multiplier: 1}),
				Source: game.TokenDef(inklingTokenDef()),
			}}}}.Ability(),
		},
	}, log)

	attacker := addColoredSourceCard(g, game.Player2, color.Red)

	// First combat: damage prevented, tally starts.
	if dealt := dealPlayerDamage(g, attacker, 0, game.Player2, game.Player1, firstCombat, true); dealt != 0 {
		t.Fatalf("first-combat damage dealt to shielded controller = %d, want 0", dealt)
	}
	// Passing an end-of-combat boundary must NOT fire the end-step payoff.
	engine.runCombatPhase(g, allFirstLegalAgents(), &TurnLog{})
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("delayed triggers after first combat = %d, want 1 (payoff waits for the end step)", len(g.DelayedTriggers))
	}
	if got := countTokenPermanentsNamed(g, "Inkling"); got != 0 {
		t.Fatalf("Inkling tokens after first combat = %d, want 0 (payoff waits for the end step)", got)
	}

	// Extra combat this turn: the same still-active shield prevents more damage,
	// growing the tally the end-step payoff will read.
	if dealt := dealPlayerDamage(g, attacker, 0, game.Player2, game.Player1, secondCombat, true); dealt != 0 {
		t.Fatalf("second-combat damage dealt to shielded controller = %d, want 0", dealt)
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})
	if got := countTokenPermanentsNamed(g, "Inkling"); got != firstCombat+secondCombat {
		t.Fatalf("Inkling tokens = %d, want %d (all combat damage prevented this turn)", got, firstCombat+secondCombat)
	}
}

// TestInkshieldPreventsOnlyControllerCombatDamage confirms the "to you" shield
// is scoped to its controller and to combat damage: it does not prevent combat
// damage to another player, nor noncombat damage to the controller, and only the
// combat damage it actually prevents is tallied for the payoff.
func TestInkshieldPreventsOnlyControllerCombatDamage(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardInstanceID := g.IDGen.Next()
	spell := &game.StackObject{
		Kind:       game.StackSpell,
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   cardInstanceID,
	}

	resolveInstruction(engine, g, spell, game.PreventDamage{
		Player:     game.ControllerReference(),
		All:        true,
		CombatOnly: true,
	}, &TurnLog{})

	source := addColoredSourceCard(g, game.Player2, color.Red)
	// Combat damage to the controller is prevented and tallied.
	if dealt := dealPlayerDamage(g, source, 0, game.Player2, game.Player1, 3, true); dealt != 0 {
		t.Fatalf("combat damage to controller = %d, want 0", dealt)
	}
	// Combat damage to a different player is unaffected.
	if dealt := dealPlayerDamage(g, source, 0, game.Player1, game.Player2, 4, true); dealt != 4 {
		t.Fatalf("combat damage to non-controller = %d, want 4", dealt)
	}
	// Noncombat damage to the controller is unaffected.
	if dealt := dealPlayerDamage(g, source, 0, game.Player2, game.Player1, 2, false); dealt != 2 {
		t.Fatalf("noncombat damage to controller = %d, want 2", dealt)
	}

	var tally int
	for i := range g.PreventionShields {
		if g.PreventionShields[i].SourceID == cardInstanceID {
			tally = g.PreventionShields[i].Prevented
		}
	}
	if tally != 3 {
		t.Fatalf("shield.Prevented = %d, want 3 (only the controller's combat damage)", tally)
	}
}

// TestDamagePreventedThisWayAmountScopedToSourceCard confirms the dynamic amount
// sums only the shields the resolving card created, so a shield another card set
// up does not inflate one card's "damage prevented this way" payoff.
func TestDamagePreventedThisWayAmountScopedToSourceCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	ours := g.IDGen.Next()
	theirs := g.IDGen.Next()
	g.PreventionShields = append(g.PreventionShields,
		game.PreventionShield{ID: g.IDGen.Next(), Player: game.Player1, All: true, CombatOnly: true, SourceID: ours, Prevented: 5},
		game.PreventionShield{ID: g.IDGen.Next(), Player: game.Player1, All: true, CombatOnly: true, SourceID: theirs, Prevented: 9},
	)

	resolving := &game.StackObject{Kind: game.StackTriggeredAbility, SourceCardID: ours}
	if got := damagePreventedThisWayAmount(g, resolving); got != 5 {
		t.Fatalf("damage prevented this way for our card = %d, want 5", got)
	}
	if got := damagePreventedThisWayAmount(g, &game.StackObject{Kind: game.StackTriggeredAbility, SourceCardID: theirs}); got != 9 {
		t.Fatalf("damage prevented this way for other card = %d, want 9", got)
	}
}
