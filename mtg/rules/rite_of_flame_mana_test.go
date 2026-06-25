package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// riteOfFlameSpellDef mirrors the generated Rite of Flame card definition: it
// adds {R}{R} and then one {R} for each card named Rite of Flame in every
// graveyard (DynamicAmountCardsNamedSourceInGraveyards).
func riteOfFlameSpellDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Rite of Flame",
		Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{
				{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.R}},
				{Primitive: game.AddMana{Amount: game.Fixed(1), ManaColor: mana.R}},
				{Primitive: game.AddMana{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:       game.DynamicAmountCardsNamedSourceInGraveyards,
						Multiplier: 1,
					}),
					ManaColor: mana.R,
				}},
			},
		}.Ability()),
	}}
}

// addNamedCardToGraveyard places a card with the given name into a player's
// graveyard and returns its instance ID.
func addNamedCardToGraveyard(g *game.Game, owner game.PlayerID, name string) game.ObjectID {
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{
		ID:    cardID,
		Def:   &game.CardDef{CardFace: game.CardFace{Name: name}},
		Owner: owner,
	}
	g.Players[owner].Graveyard.Add(cardID)
	return cardID
}

func resolveRiteOfFlame(t *testing.T, g *game.Game) {
	t.Helper()
	engine := NewEngine(nil)

	spellID := g.IDGen.Next()
	g.CardInstances[spellID] = &game.CardInstance{
		ID:    spellID,
		Def:   riteOfFlameSpellDef(),
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   spellID,
		Controller: game.Player1,
	})
	engine.resolveTopOfStack(g, &TurnLog{})
}

func TestRiteOfFlameAddsBaseTwoRedWithEmptyGraveyards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	resolveRiteOfFlame(t, g)
	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 2 {
		t.Fatalf("red mana = %d, want 2 (the base {R}{R})", got)
	}
}

func TestRiteOfFlameScalesWithCardsNamedSelfInAllGraveyards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Two copies in the caster's graveyard and one in an opponent's graveyard:
	// the count spans every graveyard.
	addNamedCardToGraveyard(g, game.Player1, "Rite of Flame")
	addNamedCardToGraveyard(g, game.Player1, "Rite of Flame")
	addNamedCardToGraveyard(g, game.Player2, "Rite of Flame")
	// A differently named card must not be counted.
	addNamedCardToGraveyard(g, game.Player1, "Lightning Bolt")

	resolveRiteOfFlame(t, g)

	if got := g.Players[game.Player1].ManaPool.Amount(mana.R); got != 5 {
		t.Fatalf("red mana = %d, want 5 (base {R}{R} plus one per Rite of Flame in all graveyards)", got)
	}
}
