package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	loyalFireSageOracleID        = "0c052420-2d8b-4bbf-8197-b8a36e7208c9"
	loyalFireSageFirebendingText = "Firebending 1 (Whenever this creature attacks, add {R}. " +
		"This mana lasts until end of combat.)"
	loyalFireSageAllyText = "{5}: Create a 1/1 white Ally creature token."
)

var LoyalFireSage = cardset.Entry{
	OracleID: loyalFireSageOracleID,
	Name:     "Loyal Fire Sage",
	New:      newLoyalFireSage,
	Verification: "mtg/rules/loyal_fire_sage_test.go: " +
		"TestLoyalFireSageAttackSpendsFirebendingManaForAlly",
}

func newLoyalFireSage() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Loyal Fire Sage",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Cleric, types.Sub("Ally")},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: loyalFireSageFirebendingText + "\n" + loyalFireSageAllyText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: loyalFireSageFirebendingText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventAttackerDeclared,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddMana{
						Amount:                  game.Fixed(1),
						ManaColor:               mana.R,
						PersistUntilEndOfCombat: true,
					},
				}}}.Ability(),
			}},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           loyalFireSageAllyText,
				ZoneOfFunction: zone.Battlefield,
				ManaCost:       opt.Val(cost.Mana{cost.O(5)}),
				Content: game.Mode{
					Text: loyalFireSageAllyText,
					Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(loyalFireSageAllyToken()),
						},
					}},
				}.Ability(),
			}},
		},
	}
}

func loyalFireSageAllyToken() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Ally",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Ally")},
		Colors:    []color.Color{color.White},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}}
}
