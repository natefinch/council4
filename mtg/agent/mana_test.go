package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

func landCardDef(name string, produces mana.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(produces)},
	}}
}

func taplandCardDef(name string, produces mana.Color) *game.CardDef {
	def := landCardDef(name, produces)
	def.ReplacementAbilities = []game.ReplacementAbility{
		game.EntersTappedReplacement(name + " enters the battlefield tapped."),
	}
	return def
}

func manaRockDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:          name,
		Types:         []types.Card{types.Artifact},
		ManaCost:      opt.Val(genericCost(manaValue)),
		ManaAbilities: []game.ManaAbility{game.TapManaAbility(mana.C)},
	}}
}

func plainArtifactDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Artifact},
		ManaCost: opt.Val(genericCost(manaValue)),
	}}
}

func rampSpellDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(genericCost(manaValue)),
		SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Search{
				Player: game.ControllerReference(),
				Spec: game.SearchSpec{
					SourceZone:   zone.Library,
					Destination:  zone.Battlefield,
					Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}},
					EntersTapped: true,
				},
			},
		}}}.Ability()),
	}}
}

func plainSorceryDef(name string, manaValue int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Sorcery},
		ManaCost: opt.Val(genericCost(manaValue)),
	}}
}

func coloredCreatureDef(name string, power, toughness int, c color.Color) *game.CardDef {
	def := creatureCardDef(name, power, toughness)
	def.Colors = []color.Color{c}
	return def
}

// creatureWithCost is a creature with a generic mana cost so its mana value, and
// thus the mana the agent would spend casting it, is non-zero.
func creatureWithCost(name string, power, toughness, manaValue int) *game.CardDef {
	def := creatureCardDef(name, power, toughness)
	def.ManaCost = opt.Val(genericCost(manaValue))
	return def
}

func instantDef(name string, manaValue int, c color.Color) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{types.Instant},
		Colors:   []color.Color{c},
		ManaCost: opt.Val(genericCost(manaValue)),
	}}
}

// genericCost builds a mana cost of the given mana value as generic mana, enough
// for ManaValue() to report the cost without pinning a colour.
func genericCost(amount int) cost.Mana {
	if amount <= 0 {
		return cost.Mana{}
	}
	return cost.Mana{cost.O(amount)}
}

// TestColorScrewAvoidancePrefersMissingColorLand checks the agent plays the land
// that fixes a colour its hand needs but cannot yet produce, rather than adding
// a source of a colour it already has.
func TestColorScrewAvoidancePrefersMissingColorLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Hand wants green and blue; the battlefield already makes blue.
	addObservedHandCard(g, game.Player1, coloredCreatureDef("Green Beast", 3, 3, color.Green))
	addObservedHandCard(g, game.Player1, instantDef("Blue Trick", 1, color.Blue))
	addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))

	forestID := addObservedHandCard(g, game.Player1, landCardDef("Forest", mana.G))
	islandID := addObservedHandCard(g, game.Player1, landCardDef("Island", mana.U))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	forest := strategy.ScoreAction(obs, action.PlayLand(forestID))
	island := strategy.ScoreAction(obs, action.PlayLand(islandID))
	if forest <= island {
		t.Fatalf("forest score %v should beat already-covered island %v", forest, island)
	}
}

// TestColorFixPrefersDualOverSingleNeed checks a land fixing two missing needed
// colours outranks a land fixing only one.
func TestColorFixPrefersDualOverSingleNeed(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, coloredCreatureDef("Green Beast", 3, 3, color.Green))
	addObservedHandCard(g, game.Player1, instantDef("Blue Trick", 1, color.Blue))

	forestID := addObservedHandCard(g, game.Player1, landCardDef("Forest", mana.G))
	dual := &game.CardDef{CardFace: game.CardFace{
		Name:  "Tropical Island",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{
			game.TapManaAbility(mana.G),
			game.TapManaAbility(mana.U),
		},
	}}
	dualID := addObservedHandCard(g, game.Player1, dual)
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	if dualScore, forestScore := strategy.ScoreAction(obs, action.PlayLand(dualID)), strategy.ScoreAction(obs, action.PlayLand(forestID)); dualScore <= forestScore {
		t.Fatalf("dual score %v should beat single-need forest %v", dualScore, forestScore)
	}
}

// TestHoldUpKeepsManaForCounter checks that, holding a cheap instant, the agent
// prefers a creature that leaves enough mana open for the instant over a more
// expensive creature that would tap it too low.
func TestHoldUpKeepsManaForCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, instantDef("Counterspell", 2, color.Blue))
	cheapID := addObservedHandCard(g, game.Player1, creatureWithCost("Cheap Bear", 2, 2, 2))
	expensiveID := addObservedHandCard(g, game.Player1, creatureWithCost("Big Beast", 4, 4, 4))
	for range 5 {
		addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	}
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	cheap := strategy.ScoreAction(obs, action.CastSpell(cheapID, nil, 0, nil))
	expensive := strategy.ScoreAction(obs, action.CastSpell(expensiveID, nil, 0, nil))
	if cheap <= expensive {
		t.Fatalf("holding a counter, cheap creature %v should beat tap-out creature %v", cheap, expensive)
	}
}

// TestNoHoldUpWithoutInstant checks the hold-up penalty applies only when the
// agent actually holds a reactive instant: without one it deploys its biggest
// threat.
func TestNoHoldUpWithoutInstant(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cheapID := addObservedHandCard(g, game.Player1, creatureWithCost("Cheap Bear", 2, 2, 2))
	expensiveID := addObservedHandCard(g, game.Player1, creatureWithCost("Big Beast", 4, 4, 4))
	for range 5 {
		addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	}
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	cheap := strategy.ScoreAction(obs, action.CastSpell(cheapID, nil, 0, nil))
	expensive := strategy.ScoreAction(obs, action.CastSpell(expensiveID, nil, 0, nil))
	if expensive <= cheap {
		t.Fatalf("without a counter, big beast %v should beat cheap bear %v", expensive, cheap)
	}
}

// TestHoldUpSkippedWhenManaTooLow checks that when the agent cannot keep the
// instant up regardless of its play, it does not penalise developing.
func TestHoldUpSkippedWhenManaTooLow(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedHandCard(g, game.Player1, instantDef("Counterspell", 2, color.Blue))
	creatureID := addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 2, 2, 2))
	// Only one source: the agent can never hold up the 2-mana counter, so it
	// should just develop without a hold-up penalty.
	addObservedPermanent(g, game.Player1, landCardDef("Island", mana.U))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	withInstant := strategy.ScoreAction(obs, action.CastSpell(creatureID, nil, 0, nil))
	wantNoPenalty := scoreCastBase + 2*scoreCastPerMana + scoreCreature
	if withInstant != wantNoPenalty {
		t.Fatalf("score %v should equal un-penalised %v when hold-up is impossible", withInstant, wantNoPenalty)
	}
}

// TestRampBonusRewardsManaSourceOverPlainSpell checks the agent prefers casting a
// mana rock to an otherwise equivalent artifact with no mana ability, so it
// develops mana.
func TestRampBonusRewardsManaSourceOverPlainSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(g, game.Player1, landCardDef("Plains", mana.W))
	rockID := addObservedHandCard(g, game.Player1, manaRockDef("Mind Stone", 2))
	plainID := addObservedHandCard(g, game.Player1, plainArtifactDef("Trinket", 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	rock := strategy.ScoreAction(obs, action.CastSpell(rockID, nil, 0, nil))
	plain := strategy.ScoreAction(obs, action.CastSpell(plainID, nil, 0, nil))
	if rock <= plain {
		t.Fatalf("mana rock %v should outscore an equivalent plain artifact %v", rock, plain)
	}
}

// TestRampBonusRewardsLandFetchSpell checks a land-ramp sorcery outscores an
// equivalent sorcery that does not ramp.
func TestRampBonusRewardsLandFetchSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(g, game.Player1, landCardDef("Forest", mana.G))
	rampID := addObservedHandCard(g, game.Player1, rampSpellDef("Rampant Growth", 2))
	plainID := addObservedHandCard(g, game.Player1, plainSorceryDef("Divination", 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	ramp := strategy.ScoreAction(obs, action.CastSpell(rampID, nil, 0, nil))
	plain := strategy.ScoreAction(obs, action.CastSpell(plainID, nil, 0, nil))
	if ramp <= plain {
		t.Fatalf("land-ramp spell %v should outscore an equivalent plain sorcery %v", ramp, plain)
	}
}

// TestRampBonusFadesWhenManaDeveloped checks the ramp incentive decays to zero
// once the agent already controls plenty of mana, so late-game ramp is not
// preferred over other plays.
func TestRampBonusFadesWhenManaDeveloped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 6 {
		addObservedPermanent(g, game.Player1, landCardDef("Forest", mana.G))
	}
	rockID := addObservedHandCard(g, game.Player1, manaRockDef("Mind Stone", 2))
	plainID := addObservedHandCard(g, game.Player1, plainArtifactDef("Trinket", 2))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	rock := strategy.ScoreAction(obs, action.CastSpell(rockID, nil, 0, nil))
	plain := strategy.ScoreAction(obs, action.CastSpell(plainID, nil, 0, nil))
	if rock != plain {
		t.Fatalf("with a developed mana base, ramp bonus should be zero: rock %v, plain %v", rock, plain)
	}
}

// TestTaplandPlayedWhenManaNotNeeded checks the agent prefers dropping a tapland
// on a turn it has no play the extra untapped mana would enable, saving its
// untapped lands for later.
func TestTaplandPlayedWhenManaNotNeeded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	// Only an expensive card in hand: one more mana would not enable a play.
	addObservedHandCard(g, game.Player1, creatureWithCost("Big Beast", 6, 6, 6))
	tapID := addObservedHandCard(g, game.Player1, taplandCardDef("Tapland", mana.G))
	untapID := addObservedHandCard(g, game.Player1, landCardDef("Forest", mana.G))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	tap := strategy.ScoreAction(obs, action.PlayLand(tapID))
	untap := strategy.ScoreAction(obs, action.PlayLand(untapID))
	if tap <= untap {
		t.Fatalf("tapland %v should be preferred over untapped land %v when mana is not needed", tap, untap)
	}
}

// TestUntappedLandPreferredWhenManaNeeded checks the agent prefers an untapped
// land when one more mana would enable a play this turn.
func TestUntappedLandPreferredWhenManaNeeded(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addObservedPermanent(g, game.Player1, landCardDef("Forest", mana.G))
	// A two-drop the agent could cast this turn with exactly one more mana.
	addObservedHandCard(g, game.Player1, creatureWithCost("Bear", 2, 2, 2))
	tapID := addObservedHandCard(g, game.Player1, taplandCardDef("Tapland", mana.G))
	untapID := addObservedHandCard(g, game.Player1, landCardDef("Mountain", mana.G))
	obs := rules.NewObservation(g, game.Player1)
	strategy := GenericStrategy{}

	tap := strategy.ScoreAction(obs, action.PlayLand(tapID))
	untap := strategy.ScoreAction(obs, action.PlayLand(untapID))
	if untap <= tap {
		t.Fatalf("untapped land %v should be preferred over tapland %v when mana is needed", untap, tap)
	}
}
