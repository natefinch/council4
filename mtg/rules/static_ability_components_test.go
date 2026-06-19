package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestVisitPermanentStaticAbilityComponentsSkipsInvalidMergedFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	withStaticAbility := func(name string) *game.CardDef {
		return &game.CardDef{CardFace: game.CardFace{
			Name: name,
			StaticAbilities: []game.StaticAbility{{
				ContinuousEffects: []game.ContinuousEffect{{Layer: game.LayerType}},
			}},
		}}
	}
	permanent := addCombatPermanent(g, game.Player1, withStaticAbility("Top"))
	mergedID := addCardToHand(g, game.Player1, withStaticAbility("Merged"))
	permanent.MergedCards = []game.MergedCard{{
		CardInstanceID: mergedID,
		Face:           game.FaceIndex(99),
	}}

	count := 0
	visitPermanentStaticAbilityComponents(g, permanent, func(permanentAbilityComponent) {
		count++
	})
	if count != 1 {
		t.Fatalf("visited %d components, want only the valid top component", count)
	}
}

func TestVisitPermanentStaticAbilityComponentsSkipsCardsWithoutStaticAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCombatPermanent(g, game.Player1, &game.CardDef{})

	count := 0
	visitPermanentStaticAbilityComponents(g, permanent, func(permanentAbilityComponent) {
		count++
	})
	if count != 0 {
		t.Fatalf("visited %d components, want none", count)
	}
}
