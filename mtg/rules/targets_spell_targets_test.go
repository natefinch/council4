package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func counterSpellThatTargetsSpec(requirements []game.SpellTargetRequirement) game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Constraint: "spell that targets",
		Predicate: game.TargetPredicate{
			StackObjectKinds: []game.StackObjectKind{game.StackSpell},
			SpellTargets:     requirements,
		},
	}
}

func addBattlefieldPermanent(g *game.Game, controller game.PlayerID, name string, cardTypes []types.Card) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  name,
		Types: cardTypes,
	}})
}

func TestStackSpellTargetCandidatesRespectSpellTargetTypeRequirement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addBattlefieldPermanent(g, game.Player1, "Bear", []types.Card{types.Creature})
	artifact := addBattlefieldPermanent(g, game.Player1, "Rock", []types.Card{types.Artifact})

	atCreature := addStackSpell(g, game.Player2, "Targets Creature", []types.Card{types.Instant})
	atCreature.Targets = []game.Target{game.PermanentTarget(creature.ObjectID)}
	atArtifact := addStackSpell(g, game.Player2, "Targets Artifact", []types.Card{types.Instant})
	atArtifact.Targets = []game.Target{game.PermanentTarget(artifact.ObjectID)}

	spec := counterSpellThatTargetsSpec([]game.SpellTargetRequirement{{
		Kind:          game.SpellTargetRequirementPermanent,
		RequiredTypes: []types.Card{types.Creature},
	}})
	source := counterTargetSpell(&spec)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	if !slices.Contains(candidates, game.StackObjectTarget(atCreature.ID)) {
		t.Fatalf("candidates = %+v, want spell targeting a creature %d", candidates, atCreature.ID)
	}
	if slices.Contains(candidates, game.StackObjectTarget(atArtifact.ID)) {
		t.Fatal("candidates included spell targeting an artifact, not a creature")
	}
}

func TestStackSpellTargetCandidatesRespectControllerRelation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	mine := addBattlefieldPermanent(g, game.Player1, "My Bear", []types.Card{types.Creature})
	theirs := addBattlefieldPermanent(g, game.Player2, "Their Bear", []types.Card{types.Creature})

	atMine := addStackSpell(g, game.Player2, "Targets My Permanent", []types.Card{types.Instant})
	atMine.Targets = []game.Target{game.PermanentTarget(mine.ObjectID)}
	atTheirs := addStackSpell(g, game.Player2, "Targets Their Permanent", []types.Card{types.Instant})
	atTheirs.Targets = []game.Target{game.PermanentTarget(theirs.ObjectID)}

	spec := counterSpellThatTargetsSpec([]game.SpellTargetRequirement{{
		Kind:       game.SpellTargetRequirementPermanent,
		Controller: game.ControllerYou,
	}})
	source := counterTargetSpell(&spec)

	// The counter's controller is Player1, so "you control" means Player1.
	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	if !slices.Contains(candidates, game.StackObjectTarget(atMine.ID)) {
		t.Fatalf("candidates = %+v, want spell targeting a permanent the counter's controller owns", candidates)
	}
	if slices.Contains(candidates, game.StackObjectTarget(atTheirs.ID)) {
		t.Fatal("candidates included spell targeting an opponent's permanent under a you-control relation")
	}
}

func TestStackSpellTargetCandidatesRespectPlayerRequirement(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creature := addBattlefieldPermanent(g, game.Player1, "Bear", []types.Card{types.Creature})

	atPlayer := addStackSpell(g, game.Player2, "Targets A Player", []types.Card{types.Instant})
	atPlayer.Targets = []game.Target{game.PlayerTarget(game.Player1)}
	atPermanent := addStackSpell(g, game.Player2, "Targets A Permanent", []types.Card{types.Instant})
	atPermanent.Targets = []game.Target{game.PermanentTarget(creature.ObjectID)}

	spec := counterSpellThatTargetsSpec([]game.SpellTargetRequirement{{
		Kind: game.SpellTargetRequirementPlayer,
	}})
	source := counterTargetSpell(&spec)

	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	if !slices.Contains(candidates, game.StackObjectTarget(atPlayer.ID)) {
		t.Fatalf("candidates = %+v, want spell targeting a player %d", candidates, atPlayer.ID)
	}
	if slices.Contains(candidates, game.StackObjectTarget(atPermanent.ID)) {
		t.Fatal("candidates included spell targeting a permanent under a player requirement")
	}
}

func TestStackSpellTargetCandidatesRespectYouPlayerRelation(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	atYou := addStackSpell(g, game.Player2, "Targets You", []types.Card{types.Instant})
	atYou.Targets = []game.Target{game.PlayerTarget(game.Player1)}
	atOpponent := addStackSpell(g, game.Player2, "Targets Opponent", []types.Card{types.Instant})
	atOpponent.Targets = []game.Target{game.PlayerTarget(game.Player2)}

	spec := counterSpellThatTargetsSpec([]game.SpellTargetRequirement{{
		Kind:   game.SpellTargetRequirementPlayer,
		Player: game.PlayerYou,
	}})
	source := counterTargetSpell(&spec)

	// "you" resolves to the counter's controller, Player1.
	candidates := targetCandidatesForSpecChosenBy(g, game.Player1, game.Player1, source, 0, &spec)

	if !slices.Contains(candidates, game.StackObjectTarget(atYou.ID)) {
		t.Fatalf("candidates = %+v, want spell targeting the counter's controller", candidates)
	}
	if slices.Contains(candidates, game.StackObjectTarget(atOpponent.ID)) {
		t.Fatal("candidates included spell targeting another player under a you relation")
	}
}
