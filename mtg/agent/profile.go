package agent

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

// curveBuckets is the number of mana-curve buckets: one each for mana value 0
// through 6 and a final bucket for 7 or more.
const curveBuckets = 8

// Deck-analysis tuning. These thresholds are deliberately coarse; pre-analysis
// only needs to place a deck in a rough archetype and power band (see
// docs/research/COMMANDER-AGENT-PLAYBOOK.md §1), not value it precisely.
const (
	threatPower          = 4 // a creature this large counts as a threat
	commanderWinconPower = 5 // a commander this large is treated as a wincon

	tokenArchetypeMin = 6
	sacArchetypeMin   = 6
	controlCounterMin = 3
	controlMin        = 14
	controlThreatMax  = 10
	aggroThreatMin    = 18
	rampArchetypeMin  = 8
	rampThreatMin     = 6

	goldfishBase    = 10
	rampPerTurn     = 4 // this many ramp cards shave a turn off the kill
	tutorPerTurn    = 6 // this many tutors shave a turn (consistency/combo)
	fastCurveMV     = 2.5
	manyThreats     = 16
	minGoldfishTurn = 3
	maxGoldfishTurn = 13
)

// CardTag classifies the role a card plays in a deck. A card may carry several
// tags. Tags are derived structurally from a card's types and the effect
// primitives in its abilities, not from its oracle text.
type CardTag int

// Card tag values name the deck-building roles pre-analysis recognises.
const (
	TagThreat CardTag = iota
	TagRamp
	TagManaRock
	TagManaDork
	TagRemoval
	TagBoardWipe
	TagCounterspell
	TagDraw
	TagTutor
	TagInteraction
	TagToken
	TagSacrifice
)

// Archetype is a coarse classification of a deck's game plan.
type Archetype int

// Archetype values name the deck game plans pre-analysis recognises.
const (
	ArchetypeMidrange Archetype = iota
	ArchetypeAggro
	ArchetypeControl
	ArchetypeRamp
	ArchetypeTokens
	ArchetypeAristocrats
)

// CommanderRole is how central the commander is to the deck's plan.
type CommanderRole int

// Commander role values classify a commander's strategic role.
const (
	RoleUnknown CommanderRole = iota
	RoleWincon
	RoleValue
)

// PowerBracket is a coarse power band derived from the estimated goldfish kill
// turn (the turn the deck wins unopposed).
type PowerBracket int

// Power bracket values name the deck power bands.
const (
	BracketCasual PowerBracket = iota
	BracketMid
	BracketHigh
	BracketCEDH
)

// ManaCurve summarises the nonland mana curve of a deck.
type ManaCurve struct {
	// Buckets counts nonland cards by mana value: index i is mana value i for
	// i < 7, and index 7 counts mana value 7 or more.
	Buckets      [curveBuckets]int
	NonlandCount int
	AverageMV    float64
}

// CommanderProfile is the precomputed analysis of a deck's commander.
type CommanderProfile struct {
	Name          string
	ColorIdentity []color.Color
	ManaValue     int
	// CastTrajectory is the command-zone cost for the first four casts, escalating
	// by the +2 generic commander tax per prior cast (CR 903.8).
	CastTrajectory [4]int
	Role           CommanderRole
}

// DeckProfile is the once-per-match analysis of a deck. It is a pure function of
// the deck's card definitions, so a strategy can consult it deterministically.
type DeckProfile struct {
	Commander    CommanderProfile
	Colors       []color.Color
	Curve        ManaCurve
	TagCounts    map[CardTag]int
	Archetype    Archetype
	Bracket      PowerBracket
	GoldfishTurn int
}

// AnalyzeDeck computes a DeckProfile for a player's deck and commander. It is
// pure and deterministic: the same config always yields the same profile.
func AnalyzeDeck(config game.PlayerConfig) DeckProfile {
	profile := DeckProfile{TagCounts: make(map[CardTag]int)}
	colorsSeen := make(map[color.Color]bool)

	if config.Commander != nil {
		profile.Commander = commanderProfile(config.Commander)
		for _, c := range profile.Commander.ColorIdentity {
			colorsSeen[c] = true
		}
	}

	totalMV := 0
	for _, def := range config.Deck {
		if def == nil {
			continue
		}
		if !defHasType(def, types.Land) {
			profile.Curve.NonlandCount++
			profile.Curve.Buckets[curveBucket(def.ManaValue())]++
			totalMV += def.ManaValue()
		}
		for tag := range tagsForCard(def) {
			profile.TagCounts[tag]++
		}
		for _, c := range def.Colors {
			colorsSeen[c] = true
		}
	}
	if profile.Curve.NonlandCount > 0 {
		profile.Curve.AverageMV = float64(totalMV) / float64(profile.Curve.NonlandCount)
	}
	profile.Colors = orderedColors(colorsSeen)
	profile.Archetype = classifyArchetype(profile.TagCounts, profile.Curve)
	profile.GoldfishTurn = estimateGoldfishTurn(profile.TagCounts, profile.Curve)
	profile.Bracket = bracketForTurn(profile.GoldfishTurn)
	return profile
}

func commanderProfile(def *game.CardDef) CommanderProfile {
	mv := def.ManaValue()
	role := RoleUnknown
	switch {
	case defHasType(def, types.Creature) && (defPower(def) >= commanderWinconPower || defHasEvasion(def)):
		role = RoleWincon
	default:
		tags := tagsForCard(def)
		if tags[TagDraw] || tags[TagRamp] {
			role = RoleValue
		}
	}
	return CommanderProfile{
		Name:           def.Name,
		ColorIdentity:  def.ColorIdentity.Colors(),
		ManaValue:      mv,
		CastTrajectory: [4]int{mv, mv + 2, mv + 4, mv + 6},
		Role:           role,
	}
}

// tagsForCard derives the deck-building roles a card fills from its types and
// the effect primitives in its abilities.
func tagsForCard(def *game.CardDef) map[CardTag]bool {
	tags := make(map[CardTag]bool)
	isLand := defHasType(def, types.Land)
	isCreature := defHasType(def, types.Creature)
	instantSpeed := defHasType(def, types.Instant) || def.HasKeyword(game.Flash)

	if hasIntrinsicManaAbility(def) && !isLand {
		tags[TagRamp] = true
		if isCreature {
			tags[TagManaDork] = true
		} else if defHasType(def, types.Artifact) {
			tags[TagManaRock] = true
		}
	}

	removalSpot := false
	for _, mode := range cardModes(def) {
		kinds := modePrimitiveKinds(mode)
		if kinds[game.PrimitiveAddMana] && !isLand {
			tags[TagRamp] = true
		}
		if kinds[game.PrimitiveDraw] || kinds[game.PrimitiveInvestigate] {
			tags[TagDraw] = true
		}
		if kinds[game.PrimitiveSearch] {
			tags[TagTutor] = true
		}
		if kinds[game.PrimitiveCounterObject] {
			tags[TagCounterspell] = true
		}
		if kinds[game.PrimitiveCreateToken] {
			tags[TagToken] = true
		}
		if kinds[game.PrimitiveSacrifice] || kinds[game.PrimitiveSacrificePermanents] {
			tags[TagSacrifice] = true
		}
		if modeHasRemoval(kinds) {
			// A removal effect with its own target is spot removal; one that
			// hits a group with no target slot is a board wipe.
			if len(mode.Targets) > 0 {
				tags[TagRemoval] = true
				removalSpot = true
			} else {
				tags[TagBoardWipe] = true
			}
		}
	}

	if isCreature && !isLand && (defPower(def) >= threatPower || defHasEvasion(def)) {
		tags[TagThreat] = true
	}
	if instantSpeed && (removalSpot || tags[TagCounterspell]) {
		tags[TagInteraction] = true
	}
	return tags
}

func modeHasRemoval(kinds map[game.PrimitiveKind]bool) bool {
	return kinds[game.PrimitiveDestroy] ||
		kinds[game.PrimitiveExile] ||
		kinds[game.PrimitiveDamage] ||
		kinds[game.PrimitiveBounce] ||
		kinds[game.PrimitiveFight]
}

func classifyArchetype(tags map[CardTag]int, curve ManaCurve) Archetype {
	control := tags[TagRemoval] + tags[TagBoardWipe] + tags[TagCounterspell] + tags[TagDraw]
	switch {
	case tags[TagToken] >= tokenArchetypeMin && tags[TagToken] >= tags[TagSacrifice]:
		return ArchetypeTokens
	case tags[TagSacrifice] >= sacArchetypeMin:
		return ArchetypeAristocrats
	case tags[TagCounterspell] >= controlCounterMin && control >= controlMin && tags[TagThreat] <= controlThreatMax:
		return ArchetypeControl
	case tags[TagThreat] >= aggroThreatMin && curve.AverageMV <= fastCurveMV:
		return ArchetypeAggro
	case tags[TagRamp] >= rampArchetypeMin && tags[TagThreat] >= rampThreatMin:
		return ArchetypeRamp
	default:
		return ArchetypeMidrange
	}
}

// estimateGoldfishTurn estimates the turn the deck wins unopposed: a battlecruiser
// base shortened by ramp, a fast curve, a dense threat base, and tutors.
func estimateGoldfishTurn(tags map[CardTag]int, curve ManaCurve) int {
	turn := goldfishBase
	turn -= tags[TagRamp] / rampPerTurn
	turn -= tags[TagTutor] / tutorPerTurn
	if curve.NonlandCount > 0 && curve.AverageMV <= fastCurveMV {
		turn -= 2
	}
	if tags[TagThreat] >= manyThreats {
		turn--
	}
	return clamp(turn, minGoldfishTurn, maxGoldfishTurn)
}

func bracketForTurn(turn int) PowerBracket {
	switch {
	case turn <= 4:
		return BracketCEDH
	case turn <= 7:
		return BracketHigh
	case turn <= 10:
		return BracketMid
	default:
		return BracketCasual
	}
}

func clamp(value, low, high int) int {
	return max(low, min(value, high))
}
