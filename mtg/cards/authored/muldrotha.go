package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Muldrotha declares a shared Engine permission; its usage is game-owned.
var Muldrotha = cardset.Entry{
	OracleID: "e4625704-1d52-44e4-804f-2f45644d76ac", Name: "Muldrotha, the Gravetide", New: newMuldrotha,
	Verification: "mtg/rules/muldrotha_test.go: TestMuldrothaProductionCastingChoicesSimulatorAndAllTypes",
}

func newMuldrotha() *game.CardDef {
	const text = "During each of your turns, you may play a land and cast a permanent spell of each permanent type from your graveyard. (If a card has multiple permanent types, choose one as you play it.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green, color.Blue),
		CardFace: game.CardFace{
			Name: "Muldrotha, the Gravetide", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.G, cost.U}),
			Supertypes: []types.Super{types.Legendary}, Types: []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Elemental, types.Avatar},
			Colors:   []color.Color{color.Black, color.Green, color.Blue},
			Power:    opt.Val(game.PT{Value: 6}), Toughness: opt.Val(game.PT{Value: 6}), OracleText: text,
			StaticAbilities: []game.StaticAbility{{
				Text: text, ZoneOfFunction: zone.Battlefield,
				RuleEffects: []game.RuleEffect{{
					Kind:          game.RuleEffectPlayFromGraveyardPerPermanentType,
					PermissionKey: "permanent-types", CastFromZone: zone.Graveyard,
					AffectedPlayer: game.PlayerYou, RestrictedDuringControllerTurn: true,
				}},
			}},
		},
	}
}
