package game

import "github.com/natefinch/council4/opt"

// dungeonOfTheMadMage is the Dungeon of the Mad Mage dungeon graph (CR 309).
//
//	Yawning Portal — You gain 1 life. (Leads to: Dungeon Level)
//	Dungeon Level — Scry 1. (Leads to: Goblin Bazaar, Twisted Caverns)
//	Goblin Bazaar — Create a Treasure token. (Leads to: Lost Level)
//	Twisted Caverns — Target creature can't attack until your next turn. (Leads to: Lost Level)
//	Lost Level — Scry 2. (Leads to: Runestone Caverns, Muiral's Graveyard)
//	Runestone Caverns — Exile the top two cards of your library. You may play them. (Leads to: Deep Mines)
//	Muiral's Graveyard — Create two 1/1 black Skeleton creature tokens. (Leads to: Deep Mines)
//	Deep Mines — Scry 3. (Leads to: Mad Wizard's Lair)
//	Mad Wizard's Lair — Draw three cards and reveal them. You may cast one of them without paying its mana cost.
var dungeonOfTheMadMage = &DungeonDef{
	ID:   DungeonDungeonOfTheMadMage,
	Name: "Dungeon of the Mad Mage",
	Rooms: []RoomDef{
		{
			Name:    "Yawning Portal",
			Ability: roomAbility(nil, controllerGainLife(1)),
			Next:    []int{1},
		},
		{
			Name:    "Dungeon Level",
			Ability: roomAbility(nil, controllerScry(1)),
			Next:    []int{2, 3},
		},
		{
			Name:    "Goblin Bazaar",
			Ability: roomAbility(nil, createTreasureToken()),
			Next:    []int{4},
		},
		{
			Name: "Twisted Caverns",
			Ability: roomAbility([]TargetSpec{targetCreatureSpec()}, Instruction{Primitive: ApplyRule{
				Object:      opt.Val(TargetPermanentReference(0)),
				RuleEffects: []RuleEffect{{Kind: RuleEffectCantAttack}},
				Duration:    DurationUntilYourNextTurn,
			}}),
			Next: []int{4},
		},
		{
			Name:    "Lost Level",
			Ability: roomAbility(nil, controllerScry(2)),
			Next:    []int{5, 6},
		},
		{
			Name: "Runestone Caverns",
			Ability: roomAbility(nil, Instruction{Primitive: ImpulseExile{
				Player:   ControllerReference(),
				Amount:   Fixed(2),
				Duration: DurationPermanent,
			}}),
			Next: []int{7},
		},
		{
			Name: "Muiral's Graveyard",
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(2),
				Source: TokenDef(dungeonSkeletonToken),
			}}),
			Next: []int{7},
		},
		{
			Name:    "Deep Mines",
			Ability: roomAbility(nil, controllerScry(3)),
			Next:    []int{8},
		},
		{
			Name: "Mad Wizard's Lair",
			Ability: roomAbility(nil,
				Instruction{Primitive: Draw{
					Amount:        Fixed(3),
					Player:        ControllerReference(),
					PublishLinked: LinkedKey("mad-wizards-lair-drawn"),
				}},
				Instruction{Primitive: CastLinkedCardForFree{
					Player: ControllerReference(),
					LinkID: LinkedKey("mad-wizards-lair-drawn"),
				}},
			),
		},
	},
}
