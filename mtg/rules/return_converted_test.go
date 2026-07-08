package rules

import (
	"testing"

	cardh "github.com/natefinch/council4/mtg/cards/h"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TestReturnConvertedDiesTriggerEntersBackFace drives the real dies→trigger→
// stack→resolve path to prove a "return it to the battlefield converted" dies
// trigger (lowered to game.PutOnBattlefield{EntryTransformed: true}) returns the
// double-faced card to the battlefield as a new permanent on its back face.
// Optimus Prime's front face uses this rider.
func TestReturnConvertedDiesTriggerEntersBackFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, returnConvertedBotDef())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceFront,
	}
	g.Battlefield = append(g.Battlefield, permanent)

	if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
		t.Fatal("destroyPermanent() = false, want the front-face permanent to die")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want the dies trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, cardID)
	if returned == nil {
		t.Fatal("no permanent returned to the battlefield")
	}
	if returned.Face != game.FaceBack || !returned.Transformed {
		t.Fatalf("returned face/transformed = %v/%v, want back/true", returned.Face, returned.Transformed)
	}
	if got := effectivePower(g, returned); got != 5 {
		t.Fatalf("effective power = %d, want 5 from converted back face", got)
	}
}

// TestHarvestHandDiesReturnsTransformedBackFace drives the real curated card
// Harvest Hand // Scrounged Scythe end-to-end: its "return it to the battlefield
// transformed" dies trigger must return the card on its back face (the Scrounged
// Scythe Equipment), which the parser fix for "transformed" now enables.
func TestHarvestHandDiesReturnsTransformedBackFace(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardInstance(g, game.Player1, cardh.HarvestHand())
	permanent := &game.Permanent{
		ObjectID:       g.IDGen.Next(),
		CardInstanceID: cardID,
		Owner:          game.Player1,
		Controller:     game.Player1,
		Face:           game.FaceFront,
	}
	g.Battlefield = append(g.Battlefield, permanent)

	if _, ok := destroyPermanent(g, permanent.ObjectID); !ok {
		t.Fatal("destroyPermanent() = false, want the front-face creature to die")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("putTriggeredAbilitiesOnStack() = false, want the dies trigger")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	returned := permanentForCard(g, cardID)
	if returned == nil {
		t.Fatal("Harvest Hand was not returned to the battlefield")
	}
	if returned.Face != game.FaceBack || !returned.Transformed {
		t.Fatalf("returned face/transformed = %v/%v, want back/true (Scrounged Scythe)", returned.Face, returned.Transformed)
	}
}

func returnConvertedBotDef() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Return Bot",
			Types:     []types.Card{types.Artifact, types.Creature},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentDied,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{{
							Primitive: game.PutOnBattlefield{
								Source:           game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
								EntryTransformed: true,
							},
						}},
					}.Ability(),
				},
			},
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Return Bot Vehicle",
			Types:     []types.Card{types.Artifact, types.Creature},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
		}),
	}
}
