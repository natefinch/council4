package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestEntersAsCopyUntilEndOfTurnGrantsHaste(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bearPT := game.PT{Value: 2}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(bearPT),
		Toughness: opt.Val(bearPT),
	}})
	mirror := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Cursed Mirror",
		Types: []types.Card{types.Artifact},
	}})

	replacement := game.EntersAsCopyReplacement(
		"As this artifact enters, you may have it become a copy of any creature on the battlefield until end of turn, except it has haste.",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}},
		false, false, nil, true, []game.Keyword{game.Haste}, nil,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, mirror, &replacement.Replacement)

	if got := permanentEffectiveName(g, mirror); got != "Grizzly Bears" {
		t.Fatalf("effective name = %q, want Grizzly Bears", got)
	}
	if !hasKeyword(g, mirror, game.Haste) {
		t.Fatal("until-end-of-turn copy did not grant the haste rider keyword")
	}
	var found bool
	for i := range g.ContinuousEffects {
		effect := &g.ContinuousEffects[i]
		if effect.Layer == game.LayerCopy && effect.AffectedObjectID == mirror.ObjectID {
			found = true
			if effect.Duration != game.DurationUntilEndOfTurn {
				t.Fatalf("copy effect duration = %v, want DurationUntilEndOfTurn", effect.Duration)
			}
		}
	}
	if !found {
		t.Fatal("no LayerCopy continuous effect registered for the temporary copy")
	}
}

func TestEntersAsCopyOverlaysChosenPermanentValues(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dragonPT := game.PT{Value: 4}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:            "Shivan Dragon",
		Types:           []types.Card{types.Creature},
		Subtypes:        []types.Sub{types.Dragon},
		Power:           opt.Val(dragonPT),
		Toughness:       opt.Val(dragonPT),
		StaticAbilities: []game.StaticAbility{{Text: "Flying", KeywordAbilities: game.SimpleKeywords(game.Flying)}},
	}})
	clone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Clone",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	}})

	replacement := game.EntersAsCopyReplacement(
		"You may have Clone enter the battlefield as a copy of any creature on the battlefield.",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}},
		true, false, nil, false, nil, nil,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, clone, &replacement.Replacement)

	if got := permanentEffectiveName(g, clone); got != "Shivan Dragon" {
		t.Fatalf("effective name = %q, want Shivan Dragon", got)
	}
	if got := effectivePower(g, clone); got != 4 {
		t.Fatalf("effective power = %d, want copied 4", got)
	}
	if !hasKeyword(g, clone, game.Flying) {
		t.Fatal("copy did not grant copied Flying keyword")
	}
}

func TestEntersAsCopyAddSubtypeRiderAddsSubtype(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	dragonPT := game.PT{Value: 4}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Shivan Dragon",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     opt.Val(dragonPT),
		Toughness: opt.Val(dragonPT),
	}})
	mock := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Mockingbird",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	}})

	replacement := game.EntersAsCopyReplacement(
		"You may have this creature enter as a copy of any creature on the battlefield, except it's a Bird in addition to its other types and it has flying.",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}},
		true, false, nil, false, []game.Keyword{game.Flying}, []types.Sub{types.Bird},
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, mock, &replacement.Replacement)

	if got := permanentEffectiveName(g, mock); got != "Shivan Dragon" {
		t.Fatalf("effective name = %q, want Shivan Dragon", got)
	}
	values := effectivePermanentValues(g, mock)
	if !slices.Contains(values.subtypes, types.Bird) {
		t.Errorf("subtypes = %v, want to include Bird", values.subtypes)
	}
	if !slices.Contains(values.subtypes, types.Dragon) {
		t.Errorf("subtypes = %v, want to retain copied Dragon", values.subtypes)
	}
	if !hasKeyword(g, mock, game.Flying) {
		t.Fatal("copy did not grant the flying rider keyword")
	}
}

func TestEntersAsCopyNotLegendaryRiderDropsLegendary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 3}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:       "Legendary Bear",
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Power:      opt.Val(pt),
		Toughness:  opt.Val(pt),
	}})
	clone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Spark Double",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	}})

	replacement := game.EntersAsCopyReplacement(
		"copy text",
		&game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou},
		false, true, nil, false, nil, nil,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, clone, &replacement.Replacement)

	values := effectivePermanentValues(g, clone)
	for _, super := range values.supertypes {
		if super == types.Legendary {
			t.Fatal("not-legendary rider failed to drop the legendary supertype")
		}
	}
	if got := effectivePower(g, clone); got != 3 {
		t.Fatalf("effective power = %d, want copied 3", got)
	}
}

func TestEntersAsCopyConditionalCounterMatchesCopiedType(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	pt := game.PT{Value: 3}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}})
	clone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Spark Double",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 0}),
	}})

	replacement := game.EntersAsCopyReplacement(
		"copy text",
		&game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerYou},
		false, true, []game.ConditionalCounterPlacement{
			{Kind: counter.PlusOnePlusOne, Amount: 1, IfType: types.Creature},
			{Kind: counter.Loyalty, Amount: 1, IfType: types.Planeswalker},
		}, false, nil, nil,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, clone, &replacement.Replacement)

	// The copied permanent is a creature, so only the creature-gated +1/+1
	// counter is placed; the planeswalker-gated loyalty counter is not.
	if got := clone.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
	if got := clone.Counters.Get(counter.Loyalty); got != 0 {
		t.Fatalf("loyalty counters = %d, want 0", got)
	}
}

// TestEntersTappedAsCopyEntersTapped verifies the "enter tapped as a copy" form
// (Vesuva) overlays the chosen permanent's copiable values and also taps the
// entering permanent once the copy choice is confirmed.
func TestEntersTappedAsCopyEntersTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ancient Tomb",
		Types: []types.Card{types.Land},
	}})
	vesuva := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Vesuva",
		Types: []types.Card{types.Land},
	}})

	replacement := game.EntersTappedAsCopy(game.EntersAsCopyReplacement(
		"You may have this land enter tapped as a copy of any land on the battlefield.",
		&game.Selection{RequiredTypes: []types.Card{types.Land}},
		true, false, nil, false, nil, nil,
	))
	applyEntersAsCopy(enterBattlefieldContext{}, g, vesuva, &replacement.Replacement)

	if got := permanentEffectiveName(g, vesuva); got != "Ancient Tomb" {
		t.Fatalf("effective name = %q, want Ancient Tomb", got)
	}
	if !vesuva.Tapped {
		t.Fatal("enter-tapped-as-copy did not tap the entering permanent")
	}
}

// TestEntersAsCopyWithoutTappedStaysUntapped verifies the plain enters-as-copy
// replacement (Clone) leaves the entering permanent untapped, so the tapped
// rider is scoped to the EntersTappedAsCopy form only.
func TestEntersAsCopyWithoutTappedStaysUntapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Ancient Tomb",
		Types: []types.Card{types.Land},
	}})
	clone := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Clone Land",
		Types: []types.Card{types.Land},
	}})

	replacement := game.EntersAsCopyReplacement(
		"You may have this land enter as a copy of any land on the battlefield.",
		&game.Selection{RequiredTypes: []types.Card{types.Land}},
		true, false, nil, false, nil, nil,
	)
	applyEntersAsCopy(enterBattlefieldContext{}, g, clone, &replacement.Replacement)

	if got := permanentEffectiveName(g, clone); got != "Ancient Tomb" {
		t.Fatalf("effective name = %q, want Ancient Tomb", got)
	}
	if clone.Tapped {
		t.Fatal("plain enters-as-copy must not tap the entering permanent")
	}
}
