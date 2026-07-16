package game

import (
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// undercity is the Undercity dungeon graph (CR 309). Undercity can be entered
// only through the "venture into Undercity" action (the venture rules enforce
// the printed "You can't enter this dungeon unless you 'venture into
// Undercity.'"), so it is not among the ordinary dungeons a plain venture may
// choose.
//
//	Secret Entrance — Search your library for a basic land card, reveal it, put it into your hand, then shuffle. (Leads to: Forge, Lost Well)
//	Forge — Put two +1/+1 counters on target creature. (Leads to: Trap!, Arena)
//	Lost Well — Scry 2. (Leads to: Arena, Stash)
//	Trap! — Target player loses 5 life. (Leads to: Archives)
//	Arena — Goad target creature. (Leads to: Archives, Catacombs)
//	Stash — Create a Treasure token. (Leads to: Catacombs)
//	Archives — Draw a card. (Leads to: Throne of the Dead Three)
//	Catacombs — Create a 4/1 black Skeleton creature token with menace. (Leads to: Throne of the Dead Three)
//	Throne of the Dead Three — Reveal the top ten cards of your library. Put a creature card from among them onto the battlefield with three +1/+1 counters on it. It gains hexproof until your next turn. Then shuffle.
var undercity = &DungeonDef{
	ID:   DungeonUndercity,
	Name: "Undercity",
	Rooms: []RoomDef{
		{
			Name: "Secret Entrance",
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
			Next: []int{1, 2},
		},
		{
			Name: "Forge",
			Ability: roomAbility([]TargetSpec{targetCreatureSpec()}, Instruction{Primitive: AddCounter{
				Amount:      Fixed(2),
				Object:      TargetPermanentReference(0),
				CounterKind: counter.PlusOnePlusOne,
			}}),
			Next: []int{3, 4},
		},
		{
			Name:    "Lost Well",
			Ability: roomAbility(nil, controllerScry(2)),
			Next:    []int{4, 5},
		},
		{
			Name: "Trap!",
			Ability: roomAbility([]TargetSpec{targetPlayerSpec()}, Instruction{Primitive: LoseLife{
				Amount: Fixed(5),
				Player: TargetPlayerReference(0),
			}}),
			Next: []int{6},
		},
		{
			Name: "Arena",
			Ability: roomAbility([]TargetSpec{targetCreatureSpec()}, Instruction{Primitive: Goad{
				Object: TargetPermanentReference(0),
			}}),
			Next: []int{6, 7},
		},
		{
			Name:    "Stash",
			Ability: roomAbility(nil, createTreasureToken()),
			Next:    []int{7},
		},
		{
			Name:    "Archives",
			Ability: roomAbility(nil, controllerDraw(1)),
			Next:    []int{8},
		},
		{
			Name: "Catacombs",
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenDef(dungeonMenaceSkeletonToken),
			}}),
			Next: []int{8},
		},
		{
			Name: "Throne of the Dead Three",
			Ability: roomAbility(nil, Instruction{Primitive: RevealPutOntoBattlefield{
				Player:          ControllerReference(),
				Look:            Fixed(10),
				Selection:       Selection{RequiredTypesAny: []types.Card{types.Creature}},
				Counters:        Fixed(3),
				CounterKind:     counter.PlusOnePlusOne,
				GrantKeyword:    opt.Val(Hexproof),
				KeywordDuration: DurationUntilYourNextTurn,
				Shuffle:         true,
			}}),
		},
	},
}
