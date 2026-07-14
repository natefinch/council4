package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// cagedSunManaProducedSource builds a Caged Sun permanent using the authoritative
// mana-produced event model (#3031): it triggers whenever a land's ability adds
// mana of the entry-chosen color (tap, sacrifice, or pay-life), independent of
// whether the land tapped, and adds one additional mana of that color. It mirrors
// the committed mtg/cards/c/caged_sun.go trigger pattern.
func cagedSunManaProducedSource(g *game.Game, controller game.PlayerID, chosen mana.Color) *game.Permanent {
	source := addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Caged Sun",
		Types: []types.Card{types.Artifact},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                                   game.EventManaProduced,
					Controller:                              game.TriggerControllerYou,
					RequireManaProducedByLand:               true,
					RequireProducedManaColorFromEntryChoice: true,
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{Amount: game.Fixed(1), EntryChoiceFrom: game.EntryColorChoiceKey},
			}}}.Ability(),
		}},
	}})
	source.EntryChoices = map[game.ChoiceKey]game.ResolutionChoiceResult{
		game.EntryColorChoiceKey: {Kind: game.ResolutionChoiceMana, Color: chosen},
	}
	return source
}

// sacrificeForManaLand builds a land whose mana ability sacrifices the land
// (rather than tapping it) to add one mana of m, modelling a sacrifice-for-mana
// source (CR 106.11 / 605): the source is gone by the time the mana resolves, so
// the mana-produced event must still report it was a land.
func sacrificeForManaLand(g *game.Game, controller game.PlayerID, name string, subtype types.Sub, m mana.Color) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{subtype},
		ManaAbilities: []game.ManaAbility{{
			Text:            "Sacrifice this land: Add one mana.",
			AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this land", Amount: 1}},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: m},
			}}}.Ability(),
		}},
	}})
}

// TestCagedSunManaProducedOnTappedLand proves the authoritative event fires for
// an ordinary tapped-for-mana land: tapping a Mountain (chosen color red) yields
// the land's {R} plus Caged Sun's additional {R}.
func TestCagedSunManaProducedOnTappedLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunManaProducedSource(g, game.Player1, mana.R)
	mountain := basicColorLand(g, game.Player1, "Mountain", types.Mountain, mana.R)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mountain.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Mountain mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (1 Mountain + 1 Caged Sun)", got)
	}
}

// TestCagedSunManaProducedOnSacrificeLand proves Caged Sun triggers on a land
// that adds the chosen color by sacrificing itself (no tap). The land is gone by
// emission time, but the captured source provenance still reports it was a land.
func TestCagedSunManaProducedOnSacrificeLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunManaProducedSource(g, game.Player1, mana.R)
	crypt := sacrificeForManaLand(g, game.Player1, "Sac Land", types.Mountain, mana.R)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(crypt.ObjectID, 0, nil, 0)) {
		t.Fatal("activating sacrifice-for-mana land = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (1 sacrifice land + 1 Caged Sun, no tap required)", got)
	}
}

// TestCagedSunManaProducedIgnoresNonland proves the land filter: an artifact mana
// source adding the chosen color does not fire Caged Sun.
func TestCagedSunManaProducedIgnoresNonland(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunManaProducedSource(g, game.Player1, mana.R)
	rock := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:          "Mana Rock",
		Types:         []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.R)},
	}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(rock.ObjectID, 0, nil, 0)) {
		t.Fatal("activating artifact mana ability = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Caged Sun fired for a nonland mana source")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 1 {
		t.Fatalf("red mana = %d, want 1 (artifact only; land filter must exclude it)", got)
	}
}

// TestCagedSunManaProducedMixedOutputFiresOnce proves a single mana ability that
// adds several mixed units including the chosen color fires Caged Sun once and
// adds exactly one additional chosen-color mana.
func TestCagedSunManaProducedMixedOutputFiresOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunManaProducedSource(g, game.Player1, mana.R)
	dual := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Mixed Land",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Mountain, types.Forest},
		ManaAbilities: []game.ManaAbility{{
			Text:            "{T}: Add {R}{G}.",
			AdditionalCosts: cost.Tap,
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.R}},
				{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.G}},
			}}.Ability(),
		}},
	}})

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(dual.ObjectID, 0, nil, 0)) {
		t.Fatal("activating mixed-output land = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (Caged Sun triggers once per mana event)", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (1 land + 1 Caged Sun)", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.G); got != 1 {
		t.Fatalf("green mana = %d, want 1 (land only; not chosen color)", got)
	}
}

// TestCagedSunManaProducedNoRecursion proves Caged Sun's own additional mana does
// not emit a mana-produced event and so cannot retrigger itself: only one extra
// mana is added, not an unbounded loop.
func TestCagedSunManaProducedNoRecursion(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cagedSunManaProducedSource(g, game.Player1, mana.R)
	mountain := basicColorLand(g, game.Player1, "Mountain", types.Mountain, mana.R)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mountain.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Mountain mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})
	// Resolving Caged Sun must not have scheduled another trigger.
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size after resolution = %d, want 0 (Caged Sun's own mana must not retrigger)", got)
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Caged Sun retriggered on its own added mana (recursion)")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (no recursion)", got)
	}
}

// highTideDelayedTrigger schedules High Tide's until-end-of-turn repeating
// delayed trigger controlled by controller, mirroring mtg/cards/h/high_tide.go:
// whenever a player taps an Island for mana, that player adds an additional {U}.
func highTideDelayedTrigger(g *game.Game, controller game.PlayerID) {
	def := &game.DelayedTriggerDef{
		EventPattern: opt.Val(game.TriggerPattern{
			Event:                game.EventManaProduced,
			Controller:           game.TriggerControllerAny,
			RequireTappedForMana: true,
			SubjectSelection:     game.Selection{SubtypesAny: []types.Sub{types.Island}},
		}),
		Window: game.DelayedWindowThisTurn,
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.AddMana{
				Amount:    game.Fixed(1),
				ManaColor: mana.U,
				Player:    opt.Val(game.EventPlayerReference()),
			},
		}}}.Ability(),
	}
	if !scheduleDelayedTrigger(g, &game.StackObject{Controller: controller}, def) {
		panic("scheduleDelayedTrigger returned false")
	}
}

// TestHighTideAddsUToTappingPlayerAnyController proves High Tide's additional {U}
// goes to the player who taps the Island, regardless of who controls High Tide.
// Player1 casts High Tide; Player2 taps an Island and receives the extra {U}.
func TestHighTideAddsUToTappingPlayerAnyController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player2
	engine := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	island := basicColorLand(g, game.Player2, "Island", types.Island, mana.U)

	if !engine.applyAction(g, game.Player2, action.ActivateAbility(island.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Island mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].ManaPool.Amount(mana.U); got != 2 {
		t.Fatalf("Player2 blue mana = %d, want 2 (1 Island + 1 High Tide)", got)
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 0 {
		t.Fatalf("Player1 blue mana = %d, want 0 (mana goes to the tapping player)", got)
	}
}

// TestHighTideIslandOnly proves the subtype filter: tapping a non-Island land for
// mana does not fire High Tide.
func TestHighTideIslandOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	mountain := basicColorLand(g, game.Player1, "Mountain", types.Mountain, mana.R)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(mountain.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Mountain mana ability = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("High Tide fired for a non-Island land")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 0 {
		t.Fatalf("blue mana = %d, want 0 (non-Island produces no High Tide mana)", got)
	}
}

// TestHighTideTapOnly proves High Tide requires the Island to tap for mana: an
// Island that adds mana by sacrificing itself (no tap) does not fire it.
func TestHighTideTapOnly(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	island := sacrificeForManaLand(g, game.Player1, "Sac Island", types.Island, mana.U)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(island.ObjectID, 0, nil, 0)) {
		t.Fatal("activating sacrifice-for-mana Island = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("High Tide fired for a non-tap (sacrifice) Island mana ability")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 1 {
		t.Fatalf("blue mana = %d, want 1 (sacrifice Island only; tap required)", got)
	}
}

// TestMultipleHighTidesStack proves two resolved High Tides each add {U}: tapping
// one Island yields the Island's {U} plus two additional {U}.
func TestMultipleHighTidesStack(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	highTideDelayedTrigger(g, game.Player1)
	island := basicColorLand(g, game.Player1, "Island", types.Island, mana.U)

	if !engine.applyAction(g, game.Player1, action.ActivateAbility(island.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Island mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	if got := g.Stack.Size(); got != 2 {
		t.Fatalf("stack size = %d, want 2 (both High Tides trigger)", got)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 3 {
		t.Fatalf("blue mana = %d, want 3 (1 Island + 2 High Tides)", got)
	}
}

// TestHighTideFiresFromPaymentTap proves the authoritative event is emitted from
// the payment path too: tapping an Island for mana to pay a spell's cost fires
// High Tide, leaving the tapping player an additional {U} floating.
func TestHighTideFiresFromPaymentTap(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhasePrecombatMain
	g.Turn.Step = game.StepNone
	engine := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)
	basicColorLand(g, game.Player1, "Island", types.Island, mana.U)
	spellID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:         "Blue Spell",
		ManaCost:     opt.Val(cost.Mana{cost.U}),
		Types:        []types.Card{types.Instant},
		SpellAbility: opt.Val(game.AbilityContent{}),
	}})

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("casting spell paid by tapping an Island = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	// Resolve High Tide's trigger (it sits above the spell on the stack).
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 1 {
		t.Fatalf("blue mana = %d, want 1 (Island {U} spent on the spell, High Tide's {U} floating)", got)
	}
}

// TestHighTideExpiresAtCleanup proves the until-end-of-turn window: after cleanup
// removes the delayed trigger, tapping an Island next turn adds no extra {U}.
func TestHighTideExpiresAtCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	engine := NewEngine(nil)
	highTideDelayedTrigger(g, game.Player1)

	expireEventDelayedTriggers(g)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("this-turn delayed trigger survived cleanup: %d", len(g.DelayedTriggers))
	}

	island := basicColorLand(g, game.Player1, "Island", types.Island, mana.U)
	if !engine.applyAction(g, game.Player1, action.ActivateAbility(island.ObjectID, 0, nil, 0)) {
		t.Fatal("activating Island mana ability = false, want true")
	}
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("High Tide fired after its until-end-of-turn window ended")
	}
	if got := g.Players[game.Player1].ManaPool.Amount(mana.U); got != 1 {
		t.Fatalf("blue mana = %d, want 1 (High Tide expired; Island only)", got)
	}
}

// vorinclexUntapDenier builds a Vorinclex, Voice of Hunger-style permanent whose
// migrated trigger fires whenever an OPPONENT taps a land for mana (the
// authoritative EventManaProduced event, RequireTappedForMana), and marks that
// land ("that land", EventPermanentReference) to skip its next untap step. It
// proves the tap-for-mana family migration keeps its event-permanent binding.
func vorinclexUntapDenier(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Vorinclex",
		Types: []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                game.EventManaProduced,
					Controller:           game.TriggerControllerOpponent,
					RequireTappedForMana: true,
					SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Land}},
				},
			},
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.SkipNextUntap{Object: game.EventPermanentReference()},
			}}}.Ability(),
		}},
	}})
}

// TestManaProducedEventPermanentBindingMarksTappedLand proves the migrated
// tap-for-mana family (Vorinclex, Voice of Hunger) still resolves "that land" to
// the permanent that produced the mana: an opponent tapping a land for mana fires
// the EventManaProduced trigger and exerts that exact land.
func TestManaProducedEventPermanentBindingMarksTappedLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.PriorityPlayer = game.Player2
	engine := NewEngine(nil)
	vorinclexUntapDenier(g, game.Player1)
	forest := basicColorLand(g, game.Player2, "Forest", types.Forest, mana.G)

	if !engine.applyAction(g, game.Player2, action.ActivateAbility(forest.ObjectID, 0, nil, 0)) {
		t.Fatal("activating opponent's Forest mana ability = false, want true")
	}
	engine.putTriggeredAbilitiesOnStack(g)
	engine.resolveTopOfStack(g, &TurnLog{})

	if !forest.Exerted {
		t.Fatal("tapped-for-mana land was not marked to skip its next untap step")
	}
}
