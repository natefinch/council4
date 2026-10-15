package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GandalfTheWhite is the card definition for Gandalf the White.
//
// Type: Legendary Creature — Avatar Wizard
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Flash
//	You may cast legendary spells and artifact spells as though they had flash.
//	If a legendary permanent or an artifact entering or leaving the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
var GandalfTheWhite = newGandalfTheWhite

func newGandalfTheWhite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Gandalf the White",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Avatar, types.Wizard},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					Text: "You may cast legendary spells and artifact spells as though they had flash.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellCharacteristicFilters: []game.CharacteristicFilter{
								game.CharacteristicFilter{
									Supertypes: []types.Super{types.Legendary},
								},
								game.CharacteristicFilter{
									Types: []types.Card{types.Artifact},
								},
							},
						},
					},
				},
				game.StaticAbility{
					Text: "If a legendary permanent or an artifact entering or leaving the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectAdditionalTriggerForControlledPermanent,
							Duration: game.DurationPermanent,
							TriggerCausePermanentFilters: []game.CharacteristicFilter{
								game.CharacteristicFilter{
									Supertypes: []types.Super{types.Legendary},
								},
								game.CharacteristicFilter{
									Types: []types.Card{types.Artifact},
								},
							},
							TriggerCausePermanentEnters: true,
							TriggerCausePermanentLeaves: true,
						},
					},
				},
			},
			OracleText: `
			Flash
			You may cast legendary spells and artifact spells as though they had flash.
			If a legendary permanent or an artifact entering or leaving the battlefield causes a triggered ability of a permanent you control to trigger, that ability triggers an additional time.
		`,
		},
	}
}
