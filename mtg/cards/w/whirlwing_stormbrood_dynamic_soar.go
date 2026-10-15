package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WhirlwingStormbrood is the card definition for Whirlwing Stormbrood // Dynamic Soar.
//
// Type: Creature — Dragon // Sorcery — Omen
// Cost: {4}{U} // {2}{G}
// Face: Dynamic Soar — Sorcery — Omen ({2}{G})
//
// Oracle text:
//
//	Flash
//	Flying
//	You may cast sorcery spells and Dragon spells as though they had flash.
var WhirlwingStormbrood = newWhirlwingStormbrood

func newWhirlwingStormbrood() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Whirlwing Stormbrood",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "You may cast sorcery spells and Dragon spells as though they had flash.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastSpellsAsThoughFlash,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							SpellCharacteristicFilters: []game.CharacteristicFilter{
								game.CharacteristicFilter{
									Types: []types.Card{types.Sorcery},
								},
								game.CharacteristicFilter{
									Subtypes: []types.Sub{types.Dragon},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Flash
			Flying
			You may cast sorcery spells and Dragon spells as though they had flash.
		`,
		},
		Layout: game.LayoutAdventure,
		Alternate: opt.Val(game.CardFace{
			Name: "Dynamic Soar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Sorcery},
			Subtypes: []types.Sub{types.Omen},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.AddCounter{
							Amount:      game.Fixed(3),
							Object:      game.TargetPermanentReference(0),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Put three +1/+1 counters on target creature you control. (Then shuffle this card into its owner's library.)
		`,
		}),
	}
}
