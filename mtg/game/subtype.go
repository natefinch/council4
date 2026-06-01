package game

const (
	ArtifactSubtypeClue      = "Clue"
	ArtifactSubtypeEquipment = "Equipment"
)

const (
	CreatureSubtypeAngel       = "Angel"
	CreatureSubtypeBear        = "Bear"
	CreatureSubtypeBeast       = "Beast"
	CreatureSubtypeBird        = "Bird"
	CreatureSubtypeCleric      = "Cleric"
	CreatureSubtypeConstruct   = "Construct"
	CreatureSubtypeDruid       = "Druid"
	CreatureSubtypeGolem       = "Golem"
	CreatureSubtypeHuman       = "Human"
	CreatureSubtypeIncarnation = "Incarnation"
	CreatureSubtypeMutant      = "Mutant"
	CreatureSubtypeNinja       = "Ninja"
	CreatureSubtypeRobot       = "Robot"
	CreatureSubtypeShaman      = "Shaman"
	CreatureSubtypeSnake       = "Snake"
	CreatureSubtypeTurtle      = "Turtle"
	CreatureSubtypeZombie      = "Zombie"
)

const (
	EnchantmentSubtypeAura = "Aura"
)

const (
	LandSubtypeForest   = "Forest"
	LandSubtypeIsland   = "Island"
	LandSubtypeMountain = "Mountain"
	LandSubtypePlains   = "Plains"
	LandSubtypeSwamp    = "Swamp"
)

var subtypesByType = map[CardType]map[string]struct{}{
	TypeArtifact: subtypeSet(
		ArtifactSubtypeClue,
		ArtifactSubtypeEquipment,
	),
	TypeCreature: subtypeSet(
		CreatureSubtypeAngel,
		CreatureSubtypeBear,
		CreatureSubtypeBeast,
		CreatureSubtypeBird,
		CreatureSubtypeCleric,
		CreatureSubtypeConstruct,
		CreatureSubtypeDruid,
		CreatureSubtypeGolem,
		CreatureSubtypeHuman,
		CreatureSubtypeIncarnation,
		CreatureSubtypeMutant,
		CreatureSubtypeNinja,
		CreatureSubtypeRobot,
		CreatureSubtypeShaman,
		CreatureSubtypeSnake,
		CreatureSubtypeTurtle,
		CreatureSubtypeZombie,
	),
	TypeEnchantment: subtypeSet(
		EnchantmentSubtypeAura,
	),
	TypeLand: subtypeSet(
		LandSubtypeForest,
		LandSubtypeIsland,
		LandSubtypeMountain,
		LandSubtypePlains,
		LandSubtypeSwamp,
	),
}

func subtypeSet(subtypes ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(subtypes))
	for _, subtype := range subtypes {
		set[subtype] = struct{}{}
	}
	return set
}

// KnownSubtypeForType reports whether subtype is defined for cardType.
func KnownSubtypeForType(cardType CardType, subtype string) bool {
	if cardType == TypeKindred {
		cardType = TypeCreature
	}
	subtypes := subtypesByType[cardType]
	_, ok := subtypes[subtype]
	return ok
}
