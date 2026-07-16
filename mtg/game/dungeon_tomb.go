package game

import "github.com/natefinch/council4/mtg/game/types"

// tombOfAnnihilation is the Tomb of Annihilation dungeon graph (CR 309).
//
//	Trapped Entry — Each player loses 1 life. (Leads to: Veils of Fear, Oubliette)
//	Veils of Fear — Each player loses 2 life unless they discard a card. (Leads to: Sandfall Cell)
//	Sandfall Cell — Each player loses 2 life unless they sacrifice a creature, artifact, or land of their choice. (Leads to: Cradle of the Death God)
//	Oubliette — Discard a card and sacrifice a creature, an artifact, and a land. (Leads to: Cradle of the Death God)
//	Cradle of the Death God — Create The Atropal, a legendary 4/4 black God Horror creature token with deathtouch.
var tombOfAnnihilation = &DungeonDef{
	ID:   DungeonTombOfAnnihilation,
	Name: "Tomb of Annihilation",
	Rooms: []RoomDef{
		{
			Name: "Trapped Entry",
			Ability: roomAbility(nil, Instruction{Primitive: LoseLife{
				Amount:      Fixed(1),
				PlayerGroup: AllPlayersReference(),
			}}),
			Next: []int{1, 3},
		},
		{
			Name: "Veils of Fear",
			Ability: roomAbility(nil, Instruction{Primitive: PunisherEachLoseLife{
				PlayerGroup:  AllPlayersReference(),
				Amount:       Fixed(2),
				AllowDiscard: true,
			}}),
			Next: []int{2},
		},
		{
			Name: "Sandfall Cell",
			Ability: roomAbility(nil, Instruction{Primitive: PunisherEachLoseLife{
				PlayerGroup:        AllPlayersReference(),
				Amount:             Fixed(2),
				AllowSacrifice:     true,
				SacrificeSelection: Selection{RequiredTypesAny: []types.Card{types.Creature, types.Artifact, types.Land}},
			}}),
			Next: []int{4},
		},
		{
			Name: "Oubliette",
			Ability: roomAbility(nil,
				Instruction{Primitive: Discard{Amount: Fixed(1), Player: ControllerReference()}},
				Instruction{Primitive: SacrificePermanents{Player: ControllerReference(), Amount: Fixed(1), Selection: Selection{RequiredTypesAny: []types.Card{types.Creature}}}},
				Instruction{Primitive: SacrificePermanents{Player: ControllerReference(), Amount: Fixed(1), Selection: Selection{RequiredTypesAny: []types.Card{types.Artifact}}}},
				Instruction{Primitive: SacrificePermanents{Player: ControllerReference(), Amount: Fixed(1), Selection: Selection{RequiredTypesAny: []types.Card{types.Land}}}},
			),
			Next: []int{4},
		},
		{
			Name: "Cradle of the Death God",
			Ability: roomAbility(nil, Instruction{Primitive: CreateToken{
				Amount: Fixed(1),
				Source: TokenDef(dungeonAtropalToken),
			}}),
		},
	},
}
