package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// thespiansStageCopyAbility builds the activated become-a-copy ability of a
// Thespian's Stage-like land, copying the targeted land and retaining the copy
// ability itself ("except it has this ability").
func thespiansStageCopyAbility() game.ActivatedAbility {
	return game.ActivatedAbility{
		Text: "{2}, {T}: This land becomes a copy of target land, except it has this ability.",
		Content: game.Mode{
			Targets: []game.TargetSpec{{}},
			Sequence: []game.Instruction{{
				Primitive: game.BecomeCopy{
					Object:             game.TargetPermanentReference(0),
					RetainsThisAbility: true,
				},
			}},
		}.Ability(),
	}
}

func TestBecomeCopyRetainsThisAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Dark Depths",
		Types: []types.Card{types.Land},
	}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:               "Thespian's Stage",
		Types:              []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{thespiansStageCopyAbility()},
	}})

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		SourceID:   source.ObjectID,
		Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleBecomeCopy(r, game.BecomeCopy{
		Object:             game.TargetPermanentReference(0),
		RetainsThisAbility: true,
	})
	if !resolved.succeeded {
		t.Fatal("handleBecomeCopy did not succeed")
	}
	if got := permanentEffectiveName(g, source); got != "Dark Depths" {
		t.Fatalf("effective name = %q, want copied Dark Depths", got)
	}
	if !permanentHasBecomeCopyAbility(g, source) {
		t.Fatal("copy did not retain the become-a-copy ability")
	}
}

func TestBecomeCopyUntilEndOfTurnDuration(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Grizzly Bears",
		Types: []types.Card{types.Creature},
		Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
	}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Mirage Mirror",
		Types: []types.Card{types.Artifact},
	}})

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		SourceID:   source.ObjectID,
		Targets:    []game.Target{{Kind: game.TargetPermanent, PermanentID: target.ObjectID}},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	handleBecomeCopy(r, game.BecomeCopy{Object: game.TargetPermanentReference(0), UntilEndOfTurn: true})

	if got := permanentEffectiveName(g, source); got != "Grizzly Bears" {
		t.Fatalf("effective name = %q, want copied Grizzly Bears", got)
	}
	last := g.ContinuousEffects[len(g.ContinuousEffects)-1]
	if last.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want DurationUntilEndOfTurn", last.Duration)
	}
}

func permanentHasBecomeCopyAbility(g *game.Game, permanent *game.Permanent) bool {
	for _, ability := range permanentEffectiveAbilities(g, permanent) {
		activated, ok := ability.(*game.ActivatedAbility)
		if !ok {
			continue
		}
		if activatedAbilityHasBecomeCopy(activated) {
			return true
		}
	}
	return false
}

// TestBecomeCopyOfGraveyardCard covers a become-a-copy effect whose copy target
// is a permanent card in the controller's graveyard (Shifting Woodland): the
// source land becomes a copy of the targeted graveyard card until end of turn.
func TestBecomeCopyOfGraveyardCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	graveyardCard := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:      "Tarmogoyf",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 5}),
	}})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Shifting Woodland",
		Types: []types.Card{types.Land},
	}})

	obj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackActivatedAbility,
		Controller: game.Player1,
		SourceID:   source.ObjectID,
		Targets:    []game.Target{currentCardTarget(t, g, graveyardCard)},
	}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, log: &TurnLog{}}

	resolved := handleBecomeCopy(r, game.BecomeCopy{
		Card:           game.CardReference{Kind: game.CardReferenceTarget},
		UntilEndOfTurn: true,
	})
	if !resolved.succeeded {
		t.Fatal("handleBecomeCopy did not succeed copying a graveyard card")
	}
	if got := permanentEffectiveName(g, source); got != "Tarmogoyf" {
		t.Fatalf("effective name = %q, want copied Tarmogoyf", got)
	}
	last := g.ContinuousEffects[len(g.ContinuousEffects)-1]
	if last.Layer != game.LayerCopy || !last.CopyValues.Exists {
		t.Fatalf("expected a LayerCopy continuous effect with copy values, got %+v", last)
	}
	if last.Duration != game.DurationUntilEndOfTurn {
		t.Fatalf("duration = %v, want DurationUntilEndOfTurn", last.Duration)
	}
}
