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

const (
	goblinGrapplerOracleID  = "48c1c84c-f690-42ce-9c5a-dd09b1e4197c"
	crestedCraghornOracleID = "597c390e-deeb-4452-aeda-7cf9312a8b01"
	brontotheriumOracleID   = "6e3301a5-a99b-4b17-aa40-34ccc5904977"
	deftbladeEliteOracleID  = "9a3e3a8f-7c19-4b52-a1b4-08eb09f6127a"
	krosanVorineOracleID    = "ddb95c3a-b383-4987-ba24-72e95f8e33bf"
	lowlandTrackerOracleID  = "de06b33f-0793-419a-bf65-1961634af8e7"

	provokeText = "Provoke (Whenever this creature attacks, you may have target creature defending player controls untap and block it if able.)"

	crestedCraghornText = "Haste\n" + provokeText
	brontotheriumText   = "Trample\n" + provokeText
	deftbladeEliteText  = provokeText + "\n{1}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn."
	krosanVorineText    = provokeText + "\nThis creature can't be blocked by more than one creature."
	lowlandTrackerText  = "First strike\n" + provokeText
)

var (
	GoblinGrappler = cardset.Entry{
		OracleID:     goblinGrapplerOracleID,
		Name:         "Goblin Grappler",
		New:          newGoblinGrappler,
		Verification: "mtg/rules/provoke_family_test.go: TestGoblinGrapplerProvokesDefendingCreature",
	}
	CrestedCraghorn = cardset.Entry{
		OracleID:     crestedCraghornOracleID,
		Name:         "Crested Craghorn",
		New:          newCrestedCraghorn,
		Verification: "mtg/rules/provoke_family_test.go: TestCrestedCraghornHasteAndProvoke",
	}
	Brontotherium = cardset.Entry{
		OracleID:     brontotheriumOracleID,
		Name:         "Brontotherium",
		New:          newBrontotherium,
		Verification: "mtg/rules/provoke_family_test.go: TestBrontotheriumTrampleAndProvoke",
	}
	DeftbladeElite = cardset.Entry{
		OracleID:     deftbladeEliteOracleID,
		Name:         "Deftblade Elite",
		New:          newDeftbladeElite,
		Verification: "mtg/rules/provoke_family_test.go: TestDeftbladeElitePreventsCombatDamage",
	}
	KrosanVorine = cardset.Entry{
		OracleID:     krosanVorineOracleID,
		Name:         "Krosan Vorine",
		New:          newKrosanVorine,
		Verification: "mtg/rules/provoke_family_test.go: TestKrosanVorineCannotBeBlockedByMoreThanOne",
	}
	LowlandTracker = cardset.Entry{
		OracleID:     lowlandTrackerOracleID,
		Name:         "Lowland Tracker",
		New:          newLowlandTracker,
		Verification: "mtg/rules/provoke_family_test.go: TestLowlandTrackerFirstStrikeAndProvoke",
	}
)

func provokeAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: provokeText,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:  game.EventAttackerDeclared,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature defending player controls",
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypes:               []types.Card{types.Creature},
					ControlledByDefendingPlayer: true,
				}),
			}},
			Sequence: []game.Instruction{
				{
					Primitive:               game.Untap{Object: game.TargetPermanentReference(0)},
					Optional:                true,
					PublishOptionalDecision: game.OptionalDecisionKey("provoke"),
				},
				{
					Primitive: game.ApplyRule{
						Object: opt.Val(game.TargetPermanentReference(0)),
						RuleEffects: []game.RuleEffect{{
							Kind:                     game.RuleEffectMustBlock,
							RequiredBlockAttackerRef: opt.Val(game.SourcePermanentReference()),
						}},
						Duration: game.DurationUntilEndOfCombat,
					},
					OptionalDecisionGate: game.OptionalDecisionKey("provoke"),
				},
			},
		}.Ability(),
	}
}

func provokeCreatureDef(
	name string,
	manaCost cost.Mana,
	cardColor color.Color,
	subtypes []types.Sub,
	power, toughness int,
	oracleText string,
	staticAbilities []game.StaticAbility,
	activatedAbilities []game.ActivatedAbility,
) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(cardColor),
		CardFace: game.CardFace{
			Name:               name,
			ManaCost:           opt.Val(manaCost),
			Types:              []types.Card{types.Creature},
			Subtypes:           subtypes,
			Colors:             []color.Color{cardColor},
			Power:              opt.Val(game.PT{Value: power}),
			Toughness:          opt.Val(game.PT{Value: toughness}),
			OracleText:         oracleText,
			StaticAbilities:    staticAbilities,
			TriggeredAbilities: []game.TriggeredAbility{provokeAbility()},
			ActivatedAbilities: activatedAbilities,
		},
	}
}

func newGoblinGrappler() *game.CardDef {
	return provokeCreatureDef(
		"Goblin Grappler",
		cost.Mana{cost.R},
		color.Red,
		[]types.Sub{types.Goblin},
		1, 1, provokeText, nil, nil,
	)
}

func newCrestedCraghorn() *game.CardDef {
	return provokeCreatureDef(
		"Crested Craghorn",
		cost.Mana{cost.O(4), cost.R},
		color.Red,
		[]types.Sub{types.Goat, types.Beast},
		4, 1, crestedCraghornText,
		[]game.StaticAbility{game.HasteStaticBody}, nil,
	)
}

func newBrontotherium() *game.CardDef {
	return provokeCreatureDef(
		"Brontotherium",
		cost.Mana{cost.O(4), cost.G, cost.G},
		color.Green,
		[]types.Sub{types.Beast},
		5, 3, brontotheriumText,
		[]game.StaticAbility{game.TrampleStaticBody}, nil,
	)
}

func newDeftbladeElite() *game.CardDef {
	return provokeCreatureDef(
		"Deftblade Elite",
		cost.Mana{cost.W},
		color.White,
		[]types.Sub{types.Human, types.Soldier},
		1, 1, deftbladeEliteText, nil,
		[]game.ActivatedAbility{{
			Text:           "{1}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.",
			ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.W}),
			ZoneOfFunction: zone.Battlefield,
			Content: game.Mode{Sequence: []game.Instruction{
				{Primitive: game.PreventDamage{Object: game.SourcePermanentReference(), All: true, CombatOnly: true}},
				{Primitive: game.PreventDamage{Object: game.SourcePermanentReference(), All: true, CombatOnly: true, BySource: true}},
			}}.Ability(),
		}},
	)
}

func newKrosanVorine() *game.CardDef {
	return provokeCreatureDef(
		"Krosan Vorine",
		cost.Mana{cost.O(3), cost.G},
		color.Green,
		[]types.Sub{types.Cat, types.Beast},
		3, 2, krosanVorineText,
		[]game.StaticAbility{{
			Text: "This creature can't be blocked by more than one creature.",
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantBeBlockedByMoreThanOne,
				AffectedSource: true,
				Duration:       game.DurationPermanent,
			}},
		}}, nil,
	)
}

func newLowlandTracker() *game.CardDef {
	return provokeCreatureDef(
		"Lowland Tracker",
		cost.Mana{cost.O(4), cost.W},
		color.White,
		[]types.Sub{types.Human, types.Soldier},
		2, 2, lowlandTrackerText,
		[]game.StaticAbility{game.FirstStrikeStaticBody}, nil,
	)
}
