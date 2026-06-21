package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// conditionalHexproofEquipment is an Equipment that grants its attached creature
// hexproof, but only while that creature is legendary. It mirrors the lowered
// shape of Champion's Helm's "As long as equipped creature is legendary, it has
// hexproof.": a StaticAbility whose Condition inspects the attached object's
// supertype and whose continuous effect grants Hexproof to the attached object.
func conditionalHexproofEquipment(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:     "Conditional Helm",
		Types:    []types.Card{types.Artifact},
		Subtypes: []types.Sub{types.Equipment},
		StaticAbilities: []game.StaticAbility{{
			Condition: opt.Val(game.Condition{
				Object:        opt.Val(game.SourceAttachedPermanentReference()),
				ObjectMatches: opt.Val(game.Selection{Supertypes: []types.Super{types.Legendary}}),
			}),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				Group:       game.AttachedObjectGroup(game.SourcePermanentReference()),
				AddKeywords: []game.Keyword{game.Hexproof},
			}},
		}},
	}})
}

// TestEquippedCreatureHexproofWhileLegendary confirms the conditional keyword
// grant applies the keyword to the attached creature only while that creature is
// legendary, matching Champion's Helm's "As long as equipped creature is
// legendary, it has hexproof."
func TestEquippedCreatureHexproofWhileLegendary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	equipment := conditionalHexproofEquipment(g, game.Player1)

	creature.Attachments = append(creature.Attachments, equipment.ObjectID)
	equipment.AttachedTo = opt.Val(creature.ObjectID)

	// A nonlegendary equipped creature does not gain hexproof.
	if hasKeyword(g, creature, game.Hexproof) {
		t.Fatal("nonlegendary equipped creature has hexproof, want none")
	}

	// Once the creature becomes legendary the condition is satisfied and the
	// grant applies.
	creature.CardInstanceID = g.IDGen.Next()
	g.CardInstances[creature.CardInstanceID] = &game.CardInstance{
		ID:    creature.CardInstanceID,
		Owner: game.Player1,
		Def: &game.CardDef{CardFace: game.CardFace{
			Name:       "Legendary Combat Creature",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
		}},
	}
	if !hasKeyword(g, creature, game.Hexproof) {
		t.Fatal("legendary equipped creature lacks hexproof, want hexproof")
	}

	// The grant stops applying once the Equipment leaves the battlefield.
	g.Battlefield = g.Battlefield[:len(g.Battlefield)-1]
	if hasKeyword(g, creature, game.Hexproof) {
		t.Fatal("creature retains hexproof after equipment leaves, want none")
	}
}
