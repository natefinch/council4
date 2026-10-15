package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AncientGreenwarden is the card definition for Ancient Greenwarden.
//
// Type: Creature — Elemental
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	You may play lands from your graveyard.
//	If a land entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
var AncientGreenwarden = newAncientGreenwarden

func newAncientGreenwarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Ancient Greenwarden",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.PlayLandsFromGraveyardStaticBody,
				game.StaticAbility{
					Text: "If a land entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectAdditionalTriggerForEnteringPermanent,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Land},
						},
					},
				},
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			You may play lands from your graveyard.
			If a land entering causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
		`,
		},
	}
}
