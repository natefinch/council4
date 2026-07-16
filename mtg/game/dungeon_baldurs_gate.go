package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// baldursGateWilderness is the Baldur's Gate Wilderness dungeon (CR 309). It is a
// free-traversal dungeon: each venture enters any room the player has not yet
// visited, and the dungeon is completed once all nineteen rooms have been
// visited. Room text is taken verbatim from the dungeon card's Scryfall Oracle
// text.
var baldursGateWilderness = &DungeonDef{
	ID:            DungeonBaldursGateWilderness,
	Name:          "Baldur's Gate Wilderness",
	FreeTraversal: true,
	Rooms: []RoomDef{
		{
			Name: "Crash Landing", // Search your library for a basic land card, reveal it, put it into your hand, then shuffle.
			Ability: roomAbility(nil, Instruction{Primitive: Search{
				Player: ControllerReference(),
				Spec: SearchSpec{
					SourceZone:  zone.Library,
					Destination: zone.Hand,
					Filter:      Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
					Reveal:      true,
				},
				Amount: Fixed(1),
			}}),
		},
		{
			Name:    "Goblin Camp", // Create a Treasure token.
			Ability: roomAbility(nil, createTreasureToken()),
		},
		{
			Name: "Emerald Grove", // Create a 2/2 white Knight creature token.
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenDef(dungeonKnightToken),
			}}),
		},
		{
			Name:    "Auntie's Teahouse", // Scry 3.
			Ability: roomAbility(nil, controllerScry(3)),
		},
		{
			Name: "Defiled Temple", // You may sacrifice a permanent. If you do, draw a card.
			Ability: roomAbility(nil,
				Instruction{
					Primitive:     SacrificePermanents{Amount: Fixed(1), Player: ControllerReference()},
					Optional:      true,
					PublishResult: ResultKey("defiled-temple-sacrificed"),
				},
				Instruction{
					Primitive:  Draw{Amount: Fixed(1), Player: ControllerReference()},
					ResultGate: opt.Val(InstructionResultGate{Key: "defiled-temple-sacrificed", Succeeded: TriTrue}),
				},
			),
		},
		{
			Name: "Mountain Pass", // You may put a land card from your hand onto the battlefield.
			Ability: roomAbility(nil, Instruction{
				Primitive: ChooseFromZone{
					Player:      ControllerReference(),
					SourceZone:  zone.Hand,
					Filter:      Selection{RequiredTypes: []types.Card{types.Land}},
					Quantity:    Fixed(1),
					Destination: ChooseDestination{Zone: zone.Battlefield},
					Prompt:      "Choose a land card to put onto the battlefield",
				},
				Optional: true,
			}),
		},
		{
			Name: "Ebonlake Grotto", // Create two 1/1 blue Faerie Dragon creature tokens with flying.
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(2),
				Source: TokenDef(dungeonFaerieDragonToken),
			}}),
		},
		{
			Name:    "Grymforge", // For each opponent, goad up to one target creature that player controls.
			Ability: roomAbility(nil, Instruction{Primitive: GoadForEachOpponent{}}),
		},
		{
			Name: "Githyanki Crèche", // Distribute three +1/+1 counters among up to three target creatures you control.
			Ability: roomAbility([]TargetSpec{{
				MinTargets: 0,
				MaxTargets: 3,
				Constraint: "up to three target creatures you control",
				Allow:      TargetAllowPermanent,
				Selection:  opt.Val(Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: ControllerYou}),
			}}, Instruction{Primitive: AddCounter{
				Amount:      Fixed(3),
				Object:      AllTargetPermanentsReference(0),
				CounterKind: counter.PlusOnePlusOne,
				Distribute:  true,
			}}),
		},
		{
			Name:    "Last Light Inn", // Draw two cards.
			Ability: roomAbility(nil, controllerDraw(2)),
		},
		{
			Name: "Reithwin Tollhouse", // Roll 2d4 and create that many Treasure tokens.
			Ability: roomAbility(nil, Instruction{Primitive: RollDiceCreateTokens{
				Dice:   2,
				Sides:  4,
				Source: TokenDef(dungeonTreasureToken),
			}}),
		},
		{
			Name: "Moonrise Towers", // Instant and sorcery spells you cast this turn cost {3} less to cast.
			Ability: roomAbility(nil, Instruction{Primitive: ApplyRule{
				RuleEffects: []RuleEffect{{
					Kind:           RuleEffectCostModifier,
					AffectedPlayer: PlayerYou,
					CostModifier: CostModifier{
						Kind:             CostModifierSpell,
						CardSelection:    Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
						GenericReduction: 3,
					},
				}},
				Duration: DurationThisTurn,
			}}),
		},
		{
			Name: "Gauntlet of Shar", // Each opponent loses 5 life.
			Ability: roomAbility(nil, Instruction{Primitive: LoseLife{
				Amount:      Fixed(5),
				PlayerGroup: OpponentsReference(),
			}}),
		},
		{
			Name: "Balthazar's Lab", // Return up to two target creature cards from your graveyard to your hand.
			Ability: roomAbility([]TargetSpec{{
				MinTargets: 0,
				MaxTargets: 2,
				Constraint: "up to two target creature cards from your graveyard",
				Allow:      TargetAllowCard,
				TargetZone: zone.Graveyard,
				Selection:  opt.Val(Selection{RequiredTypes: []types.Card{types.Creature}, Controller: ControllerYou}),
			}},
				Instruction{Primitive: MoveCard{Card: CardReference{Kind: CardReferenceTarget}, FromZone: zone.Graveyard, Destination: zone.Hand}},
				Instruction{Primitive: MoveCard{Card: CardReference{Kind: CardReferenceTarget, TargetIndex: 1}, FromZone: zone.Graveyard, Destination: zone.Hand}},
			),
		},
		{
			Name:    "Circus of the Last Days", // Create a token that's a copy of one of your commanders, except it's not legendary.
			Ability: roomAbility(nil, Instruction{Primitive: CreateCommanderCopyToken{}}),
		},
		{
			Name: "Undercity Ruins", // Create three 4/1 black Skeleton creature tokens with menace.
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(3),
				Source: TokenDef(dungeonMenaceSkeletonToken),
			}}),
		},
		{
			Name: "Steel Watch Foundry", // You get an emblem with "Creatures you control get +2/+2 and have trample."
			Ability: roomAbility(nil, Instruction{Primitive: CreateEmblem{
				EmblemAbilities: []Ability{&StaticAbility{
					Text: "Creatures you control get +2/+2 and have trample.",
					ContinuousEffects: []ContinuousEffect{
						{
							Layer:          LayerPowerToughnessModify,
							Group:          ObjectControlledGroup(SourcePermanentReference(), Selection{RequiredTypes: []types.Card{types.Creature}}),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						{
							Layer:       LayerAbility,
							Group:       ObjectControlledGroup(SourcePermanentReference(), Selection{RequiredTypes: []types.Card{types.Creature}}),
							AddKeywords: []Keyword{Trample},
						},
					},
				}},
			}}),
		},
		{
			Name: "Ansur's Sanctum", // Reveal the top four cards of your library and put them into your hand. Each opponent loses life equal to those cards' total mana value.
			Ability: roomAbility(nil, Instruction{Primitive: RevealToHandDrainManaValue{
				Amount: Fixed(4),
			}}),
		},
		{
			Name: "Temple of Bhaal", // Creatures your opponents control get -5/-5 until end of turn.
			Ability: roomAbility(nil, Instruction{Primitive: ApplyContinuous{
				ContinuousEffects: []ContinuousEffect{{
					Layer:          LayerPowerToughnessModify,
					Group:          BattlefieldGroup(Selection{RequiredTypes: []types.Card{types.Creature}, Controller: ControllerOpponent}),
					PowerDelta:     -5,
					ToughnessDelta: -5,
				}},
				Duration: DurationUntilEndOfTurn,
			}}),
		},
	},
}
