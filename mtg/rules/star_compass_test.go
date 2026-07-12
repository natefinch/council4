package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLandsProduceManaSelectionFiltersNonbasicLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	island := addBasicLandPermanent(g, game.Player1, types.Island)
	islandCard, _ := g.GetCardInstance(island.CardInstanceID)
	islandCard.Def.Supertypes = []types.Super{types.Basic}
	islandCard.Def.ManaAbilities = []game.ManaAbility{game.TapManaAbility(mana.U)}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.R)},
	}})
	choice := &game.ResolutionChoice{
		ColorSource:    game.ResolutionChoiceColorSourceLandsProduce,
		PlayerRelation: game.PlayerYou,
		Selection:      &game.Selection{Supertypes: []types.Super{types.Basic}},
	}
	if got := landsProduceMana(g, game.Player1, choice); !slices.Equal(got, []mana.Color{mana.U}) {
		t.Fatalf("mana choices = %v, want only blue from basic Island", got)
	}
}
