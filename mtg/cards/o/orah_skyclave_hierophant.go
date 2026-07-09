package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OrahSkyclaveHierophant is the card definition for Orah, Skyclave Hierophant.
//
// Type: Legendary Creature — Kor Cleric
// Cost: {2}{W}{B}
//
// Oracle text:
//
//	Lifelink
//	Whenever Orah or another Cleric you control dies, return target Cleric card with lesser mana value from your graveyard to the battlefield.
var OrahSkyclaveHierophant = newOrahSkyclaveHierophant

func newOrahSkyclaveHierophant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Orah, Skyclave Hierophant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.B,
			}),
			Colors:     []color.Color{color.Black, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kor, types.Cleric},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentDied,
							Controller:             game.TriggerControllerYou,
							SubjectSelectionOrSelf: true,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Cleric")}},
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Cleric card with lesser mana value from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Cleric")}, Controller: game.ControllerYou, ManaValueLessThanEventPermanent: true}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Lifelink
			Whenever Orah or another Cleric you control dies, return target Cleric card with lesser mana value from your graveyard to the battlefield.
		`,
		},
	}
}
