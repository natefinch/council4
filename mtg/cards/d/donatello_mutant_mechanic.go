package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/opt"
)

// Donatello, Mutant Mechanic
//
// Type: Legendary Creature — Mutant Ninja Turtle
// Cost: {3}{U}
//
// Oracle text:
//
//	{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.
//	Whenever an artifact you control is put into a graveyard from the battlefield, if it had counters on it, put those counters on up to one target artifact or creature you control.
var DonatelloMutantMechanic = &game.CardDef{
	Name: "Donatello, Mutant Mechanic",
	ManaCost: opt.Val(mana.Cost{
		mana.GenericMana(3),
		mana.ColoredMana(mana.Blue),
	}),
	ManaValue:     4,
	Colors:        []mana.Color{mana.Blue},
	ColorIdentity: mana.NewColorIdentity(mana.Blue),
	Supertypes:    []game.Supertype{game.Legendary},
	Types:         []game.CardType{game.TypeCreature},
	Subtypes:      []string{game.CreatureSubtypeMutant, game.CreatureSubtypeNinja, game.CreatureSubtypeTurtle},
	Power:         opt.Val(game.PT{Value: 3}),
	Toughness:     opt.Val(game.PT{Value: 5}),
	OracleText:    "{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.\nWhenever an artifact you control is put into a graveyard from the battlefield, if it had counters on it, put those counters on up to one target artifact or creature you control.",
	Abilities: []game.AbilityDef{
		{
			Kind:   game.ActivatedAbility,
			Text:   "{T}: Put three +1/+1 counters on target artifact you control. If it isn't a creature, it becomes a 0/0 Robot creature in addition to its other types. Activate only as a sorcery.",
			Timing: game.SorceryOnly,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Targets: []game.TargetSpec{
				{MinTargets: 1, MaxTargets: 1, Constraint: "artifact you control"},
			},
			Effects: []game.Effect{
				{Type: game.EffectAddCounter, Amount: 3, TargetIndex: 0, CounterKind: counter.PlusOnePlusOne},
				{
					Type:        game.EffectApplyContinuous,
					TargetIndex: 0,
					Condition: opt.Val(game.EffectCondition{
						Text:               "it isn't a creature",
						TargetIndex:        0,
						MatchPermanentType: true,
						PermanentType:      game.TypeCreature,
						Negate:             true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:       game.LayerType,
							AddTypes:    []game.CardType{game.TypeCreature},
							AddSubtypes: []string{game.CreatureSubtypeRobot},
						},
						{
							Layer:        game.LayerPowerToughnessSet,
							SetPower:     opt.Val(game.PT{Value: 0}),
							SetToughness: opt.Val(game.PT{Value: 0}),
						},
					},
				},
			},
		},
		{
			Kind: game.TriggeredAbility,
			Text: "Whenever an artifact you control is put into a graveyard from the battlefield, if it had counters on it, put those counters on up to one target artifact or creature you control.",
			Trigger: opt.Val(game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:                 game.EventZoneChanged,
					Controller:            game.TriggerControllerYou,
					RequirePermanentTypes: []game.CardType{game.TypeArtifact},
					MatchFromZone:         true,
					FromZone:              game.ZoneBattlefield,
					MatchToZone:           true,
					ToZone:                game.ZoneGraveyard,
				},
				InterveningIf:                          "it had counters on it",
				InterveningIfEventPermanentHadCounters: true,
			}),
			Targets: []game.TargetSpec{
				{MinTargets: 0, MaxTargets: 1, Constraint: "artifact or creature you control"},
			},
			Effects: []game.Effect{
				{
					Type:        game.EffectMoveCounters,
					TargetIndex: 0,
					CounterSource: game.CounterSourceSpec{
						Kind: game.CounterSourceEventPermanent,
					},
					Description: "move all counters from the triggering artifact to target",
				},
			},
		},
	},
}
