package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GuardianProject is the card definition for Guardian Project.
//
// Type: Enchantment
// Cost: {3}{G}
//
// Oracle text:
//
//	Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.
var GuardianProject = &game.CardDef{
	Name: "Guardian Project",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(3),
		mana.ColoredMana(mana.Green),
	}),
	ManaValue:     4,
	Colors:        []mana.Color{mana.Green},
	ColorIdentity: mana.NewColorIdentity(mana.Green),
	Types:         []types.Card{types.Enchantment},
	OracleText:    "Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.",
	Abilities: []game.AbilityDef{
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever a nontoken creature you control enters, if it doesn't have the same name as another creature you control or a creature card in your graveyard, draw a card.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                 game.EventPermanentEnteredBattlefield,
					Controller:            game.TriggerControllerYou,
					RequirePermanentTypes: []types.Card{types.Creature},
					RequireNonToken:       true,
				},
				InterveningIf: "it doesn't have the same name as another creature you control or a creature card in your graveyard",
				InterveningCondition: opt.Val(game.Condition{
					Text: "it doesn't have the same name as another creature you control or a creature card in your graveyard",
					EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true,
				}),
			}),
			Effects: []game.Effect{
				{Type: game.EffectDraw, Amount: 1, TargetIndex: -1},
			},
		},
	},
}
