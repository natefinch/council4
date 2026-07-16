package game

import (
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// This file defines the immutable dungeon graphs and their room abilities,
// composed from typed effect primitives. Room text is taken verbatim from the
// Scryfall Oracle text of each dungeon card. Each final room's ability ends with
// a CompleteDungeon instruction so the venturing player completes the dungeon as
// the final room's ability resolves (CR 309.7).

// --- Room-ability construction helpers ---

func controllerScry(n int) Instruction {
	return Instruction{Primitive: Scry{Amount: Fixed(n), Player: ControllerReference()}}
}

func controllerDraw(n int) Instruction {
	return Instruction{Primitive: Draw{Amount: Fixed(n), Player: ControllerReference()}}
}

func controllerGainLife(n int) Instruction {
	return Instruction{Primitive: GainLife{Amount: Fixed(n), Player: ControllerReference()}}
}

func createTreasureToken() Instruction {
	return Instruction{Primitive: CreateToken{Amount: Fixed(1), Source: TokenDef(dungeonTreasureToken)}}
}

func targetCreatureSpec() TargetSpec {
	return TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature",
		Allow:      TargetAllowPermanent,
		Selection:  opt.Val(Selection{RequiredTypesAny: []types.Card{types.Creature}}),
	}
}

func targetPlayerSpec() TargetSpec {
	return TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target player",
		Allow:      TargetAllowPlayer,
	}
}

// roomAbility builds a non-modal room ability from its target specs and
// instruction sequence. Dungeon completion is recorded by the runtime when a
// final room's ability leaves the stack (a stack-object marker), so a final
// room's ability is just its effect with no special completion instruction.
func roomAbility(targets []TargetSpec, seq ...Instruction) AbilityContent {
	return Mode{Targets: targets, Sequence: seq}.Ability()
}

// --- Registry ---

// dungeonRegistry maps each supported dungeon id to its immutable definition.
var dungeonRegistry = map[DungeonID]*DungeonDef{
	DungeonLostMineOfPhandelver:  lostMineOfPhandelver,
	DungeonTombOfAnnihilation:    tombOfAnnihilation,
	DungeonDungeonOfTheMadMage:   dungeonOfTheMadMage,
	DungeonUndercity:             undercity,
	DungeonBaldursGateWilderness: baldursGateWilderness,
}

// ordinaryDungeons lists the dungeons a player may choose the first time they
// "venture into the dungeon" (CR 309.5). Undercity is excluded: it can be
// entered only through "venture into Undercity".
var ordinaryDungeons = []DungeonID{
	DungeonLostMineOfPhandelver,
	DungeonTombOfAnnihilation,
	DungeonDungeonOfTheMadMage,
	DungeonBaldursGateWilderness,
}

// DungeonByID returns the immutable definition of the given dungeon, reporting
// whether it is a supported dungeon.
func DungeonByID(id DungeonID) (*DungeonDef, bool) {
	def, ok := dungeonRegistry[id]
	return def, ok
}

// OrdinaryDungeons returns the ids of the three dungeons a player may choose the
// first time they venture into the dungeon.
func OrdinaryDungeons() []DungeonID {
	return append([]DungeonID(nil), ordinaryDungeons...)
}
