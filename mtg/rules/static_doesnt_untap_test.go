package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestDoesntUntapStaticBodyKeepsSourceTapped verifies that a permanent whose own
// static ability prevents untapping stays tapped through its controller's untap
// step while still shedding summoning sickness, and that other permanents untap
// normally.
func TestDoesntUntapStaticBodyKeepsSourceTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	frozen := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Frozen Bear",
		Types:           []types.Card{types.Creature},
		Power:           opt.Val(game.PT{Value: 2}),
		Toughness:       opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{game.DoesntUntapStaticBody},
	}})
	control := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Normal Bear",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
	frozen.Tapped = true
	frozen.SummoningSick = true
	control.Tapped = true
	control.SummoningSick = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !frozen.Tapped {
		t.Fatal("doesn't-untap source untapped during its untap step")
	}
	if frozen.SummoningSick {
		t.Fatal("doesn't-untap source did not shed summoning sickness")
	}
	if control.Tapped {
		t.Fatal("doesn't-untap static ability prevented an unrelated permanent from untapping")
	}
}

// TestDoesntUntapAttachedKeepsEnchantedCreatureTapped verifies that an Aura or
// Equipment whose static ability freezes the creature it is attached to keeps
// that creature tapped through the untap step, leaving other creatures alone.
func TestDoesntUntapAttachedKeepsEnchantedCreatureTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchanted := makeCreaturePermanent(g, game.Player1, "Frozen Solid Target")
	control := makeCreaturePermanent(g, game.Player1, "Free Creature")
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Frozen Solid",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			{
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
					Allow:     game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}}},
			},
			{
				RuleEffects: []game.RuleEffect{{
					Kind:             game.RuleEffectDoesntUntap,
					AffectedAttached: true,
				}},
			},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent(aura, enchanted) = false")
	}
	enchanted.Tapped = true
	control.Tapped = true
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if !enchanted.Tapped {
		t.Fatal("enchanted creature untapped despite attached doesn't-untap rule")
	}
	if control.Tapped {
		t.Fatal("attached doesn't-untap rule prevented an unrelated creature from untapping")
	}
}

// TestDoesntUntapMonarchGuardUntapsWhenControllerIsMonarch verifies that the
// "doesn't untap during its controller's untap step unless that player is the
// monarch" guard (Fall from Favor) freezes the enchanted creature while its
// controller is not the monarch but lets it untap normally once its controller
// holds the monarch designation.
func TestDoesntUntapMonarchGuardUntapsWhenControllerIsMonarch(t *testing.T) {
	newGame := func() (*Engine, *game.Game, *game.Permanent) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		enchanted := makeCreaturePermanent(g, game.Player1, "Fallen Favorite")
		aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name:     "Fall from Favor",
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				{
					KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
						Allow:     game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					}}},
				},
				{
					RuleEffects: []game.RuleEffect{{
						Kind:                           game.RuleEffectDoesntUntap,
						AffectedAttached:               true,
						UntapUnlessControllerIsMonarch: true,
					}},
				},
			},
		}})
		if !attachPermanent(g, aura, enchanted) {
			t.Fatal("attachPermanent(aura, enchanted) = false")
		}
		enchanted.Tapped = true
		g.Turn.ActivePlayer = game.Player1
		g.Turn.PriorityPlayer = game.Player1
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})
		return engine, g, enchanted
	}

	engine, g, enchanted := newGame()
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if !enchanted.Tapped {
		t.Fatal("enchanted creature untapped while its controller was not the monarch")
	}

	engine, g, enchanted = newGame()
	g.Players[game.Player1].IsMonarch = true
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if enchanted.Tapped {
		t.Fatal("enchanted creature stayed tapped while its controller was the monarch")
	}
}

func TestUntapPhasesInAllPermanentsBeforeApplyingRestrictions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	enchanted := makeCreaturePermanent(g, game.Player1, "Earlier Frozen Target")
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Later Freezing Aura",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			{
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
					Allow:     game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}}},
			},
			{
				RuleEffects: []game.RuleEffect{{
					Kind:             game.RuleEffectDoesntUntap,
					AffectedAttached: true,
				}},
			},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent(aura, enchanted) = false")
	}
	enchanted.Tapped = true
	if !phaseOutPermanentTree(g, enchanted, game.Player1, make(map[game.ObjectID]bool)) {
		t.Fatal("phaseOutPermanentTree() = false")
	}
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player1
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Draw Step Card"}})

	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if enchanted.PhasedOut || aura.PhasedOut {
		t.Fatal("attached permanents did not phase in together")
	}
	if !enchanted.Tapped {
		t.Fatal("later phased-in Aura did not prevent earlier permanent from untapping")
	}
}

// TestCantAttackOrBlockAttachedProhibitsEnchantedCreature verifies that a
// Pacifism-style Aura that maps to AffectedAttached can't-attack and can't-block
// rule effects prevents the enchanted creature from attacking and blocking while
// leaving other creatures unaffected.
func TestCantAttackOrBlockAttachedProhibitsEnchantedCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	enchanted := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	aura := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:     "Pacifism",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			{
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
					Allow:     game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}}},
			},
			{
				RuleEffects: []game.RuleEffect{
					{Kind: game.RuleEffectCantAttack, AffectedAttached: true},
					{Kind: game.RuleEffectCantBlock, AffectedAttached: true},
				},
			},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent(aura, enchanted) = false")
	}
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareAttackers
	g.Turn.ActivePlayer = game.Player2

	if canAttackWith(g, enchanted, game.Player2) {
		t.Fatal("enchanted creature could attack through attached can't-attack rule")
	}
	if canBlockWith(g, enchanted, game.Player2) {
		t.Fatal("enchanted creature could block through attached can't-block rule")
	}
	if !canAttackWith(g, other, game.Player2) {
		t.Fatal("attached can't-attack rule affected an unrelated creature")
	}
	if !canBlockWith(g, other, game.Player2) {
		t.Fatal("attached can't-block rule affected an unrelated creature")
	}
}

// TestCantBeBlockedAttachedProhibitsBlockingEnchantedCreature verifies that an
// Aura mapping to an AffectedAttached can't-be-blocked rule effect prevents the
// enchanted creature from being blocked while leaving other attackers blockable.
func TestCantBeBlockedAttachedProhibitsBlockingEnchantedCreature(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	enchanted := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	other := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	aura := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Aether Tunnel",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Aura},
		StaticAbilities: []game.StaticAbility{
			{
				KeywordAbilities: []game.KeywordAbility{game.EnchantKeyword{Target: game.TargetSpec{
					Allow:     game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
				}}},
			},
			{
				RuleEffects: []game.RuleEffect{
					{Kind: game.RuleEffectCantBeBlocked, AffectedAttached: true},
				},
			},
		},
	}})
	if !attachPermanent(g, aura, enchanted) {
		t.Fatal("attachPermanent(aura, enchanted) = false")
	}
	g.Combat = &game.CombatState{}
	g.Turn.Phase = game.PhaseCombat
	g.Turn.Step = game.StepDeclareBlockers

	if canBlockAttacker(g, blocker, enchanted) {
		t.Fatal("blocker could block creature affected by attached can't-be-blocked rule")
	}
	if !canBlockAttacker(g, blocker, other) {
		t.Fatal("attached can't-be-blocked rule affected an unrelated attacker")
	}
}
