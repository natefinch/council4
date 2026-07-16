package game

import "github.com/natefinch/council4/mtg/game/counter"

// lostMineOfPhandelver is the Lost Mine of Phandelver dungeon graph (CR 309).
//
//	Cave Entrance — Scry 1. (Leads to: Goblin Lair, Mine Tunnels)
//	Goblin Lair — Create a 1/1 red Goblin creature token. (Leads to: Storeroom, Dark Pool)
//	Mine Tunnels — Create a Treasure token. (Leads to: Dark Pool, Fungi Cavern)
//	Storeroom — Put a +1/+1 counter on target creature. (Leads to: Temple of Dumathoin)
//	Dark Pool — Each opponent loses 1 life and you gain 1 life. (Leads to: Temple of Dumathoin)
//	Fungi Cavern — Target creature gets -4/-0 until your next turn. (Leads to: Temple of Dumathoin)
//	Temple of Dumathoin — Draw a card.
var lostMineOfPhandelver = &DungeonDef{
	ID:   DungeonLostMineOfPhandelver,
	Name: "Lost Mine of Phandelver",
	Rooms: []RoomDef{
		{
			Name:    "Cave Entrance",
			Ability: roomAbility(nil, controllerScry(1)),
			Next:    []int{1, 2},
		},
		{
			Name: "Goblin Lair",
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenDef(dungeonGoblinToken),
			}}),
			Next: []int{3, 4},
		},
		{
			Name:    "Mine Tunnels",
			Ability: roomAbility(nil, createTreasureToken()),
			Next:    []int{4, 5},
		},
		{
			Name: "Storeroom",
			Ability: roomAbility([]TargetSpec{targetCreatureSpec()}, Instruction{Primitive: AddCounter{
				Amount:      Fixed(1),
				Object:      TargetPermanentReference(0),
				CounterKind: counter.PlusOnePlusOne,
			}}),
			Next: []int{6},
		},
		{
			Name: "Dark Pool",
			Ability: roomAbility(nil,
				Instruction{Primitive: LoseLife{Amount: Fixed(1), PlayerGroup: OpponentsReference()}},
				controllerGainLife(1),
			),
			Next: []int{6},
		},
		{
			Name: "Fungi Cavern",
			Ability: roomAbility([]TargetSpec{targetCreatureSpec()}, Instruction{Primitive: ModifyPT{
				Object:         TargetPermanentReference(0),
				PowerDelta:     Fixed(-4),
				ToughnessDelta: Fixed(0),
				Duration:       DurationUntilYourNextTurn,
			}}),
			Next: []int{6},
		},
		{
			Name:    "Temple of Dumathoin",
			Ability: roomAbility(nil, controllerDraw(1)),
		},
	},
}
