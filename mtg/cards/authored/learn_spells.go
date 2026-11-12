package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var IgneousInspiration = cardset.Entry{
	OracleID: "2e31af8f-5e94-4888-857d-911253eac2cd",
	Name:     "Igneous Inspiration",
	New:      newIgneousInspiration,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestIgneousInspirationScenario",
}

var EnthusiasticStudy = cardset.Entry{
	OracleID: "4233334d-bb19-41e4-a203-2b7bb0c48f52",
	Name:     "Enthusiastic Study",
	New:      newEnthusiasticStudy,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestEnthusiasticStudyScenario",
}

var ArcaneSubtraction = cardset.Entry{
	OracleID: "4b5806fa-430f-4b78-b874-63d01a8a9de5",
	Name:     "Arcane Subtraction",
	New:      newArcaneSubtraction,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestArcaneSubtractionScenario",
}

var GuidingVoice = cardset.Entry{
	OracleID: "5c7e7c9e-00d1-497e-bee7-278bbd82e234",
	Name:     "Guiding Voice",
	New:      newGuidingVoice,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestGuidingVoiceScenario",
}

var CramSession = cardset.Entry{
	OracleID: "6e5b8bb2-9dd8-4986-9d0a-659114be5a2a",
	Name:     "Cram Session",
	New:      newCramSession,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestCramSessionScenario",
}

var StudyBreak = cardset.Entry{
	OracleID: "7aae7c7b-0b31-4b48-a8d3-81c913f604cf",
	Name:     "Study Break",
	New:      newStudyBreak,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestStudyBreakScenario",
}

var PopQuiz = cardset.Entry{
	OracleID: "dda9ac6c-1ab2-4f77-b499-f03d587163d8",
	Name:     "Pop Quiz",
	New:      newPopQuiz,
	Verification: "mtg/rules/authored_learn_spells_test.go: " +
		"TestPopQuizScenario",
}

func newIgneousInspiration() *game.CardDef {
	const text = "Igneous Inspiration deals 3 damage to any target.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Igneous Inspiration", text, cost.Mana{cost.O(2), cost.R}, types.Sorcery,
		[]color.Color{color.Red}, []game.TargetSpec{learnAnyTarget()},
		game.Instruction{Primitive: game.Damage{
			Amount:    game.Fixed(3),
			Recipient: game.AnyTargetDamageRecipient(0),
		}},
	)
}

func newEnthusiasticStudy() *game.CardDef {
	const text = "Target creature gets +3/+1 and gains trample until end of turn.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Enthusiastic Study", text, cost.Mana{cost.O(2), cost.R}, types.Instant,
		[]color.Color{color.Red}, []game.TargetSpec{creatureTarget()},
		game.Instruction{Primitive: game.ModifyPT{
			Object:         game.TargetPermanentReference(0),
			PowerDelta:     game.Fixed(3),
			ToughnessDelta: game.Fixed(1),
			Duration:       game.DurationUntilEndOfTurn,
		}},
		game.Instruction{Primitive: game.ApplyContinuous{
			Object: opt.Val(game.TargetPermanentReference(0)),
			ContinuousEffects: []game.ContinuousEffect{{
				Layer:       game.LayerAbility,
				AddKeywords: []game.Keyword{game.Trample},
			}},
			Duration: game.DurationUntilEndOfTurn,
		}},
	)
}

func newArcaneSubtraction() *game.CardDef {
	const text = "Target creature gets -4/-0 until end of turn.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Arcane Subtraction", text, cost.Mana{cost.O(1), cost.U}, types.Instant,
		[]color.Color{color.Blue}, []game.TargetSpec{creatureTarget()},
		game.Instruction{Primitive: game.ModifyPT{
			Object:     game.TargetPermanentReference(0),
			PowerDelta: game.Fixed(-4),
			Duration:   game.DurationUntilEndOfTurn,
		}},
	)
}

func newGuidingVoice() *game.CardDef {
	const text = "Put a +1/+1 counter on target creature.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Guiding Voice", text, cost.Mana{cost.W}, types.Sorcery,
		[]color.Color{color.White}, []game.TargetSpec{creatureTarget()},
		game.Instruction{Primitive: game.AddCounter{
			Object:      game.TargetPermanentReference(0),
			CounterKind: counter.PlusOnePlusOne,
			Amount:      game.Fixed(1),
		}},
	)
}

func newCramSession() *game.CardDef {
	const text = "You gain 4 life.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Cram Session", text,
		cost.Mana{cost.O(1), cost.HybridMana(mana.B, mana.G)}, types.Sorcery,
		[]color.Color{color.Black, color.Green}, nil,
		game.Instruction{Primitive: game.GainLife{
			Amount: game.Fixed(4),
			Player: game.ControllerReference(),
		}},
	)
}

func newStudyBreak() *game.CardDef {
	const text = "Tap up to two target creatures.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Study Break", text, cost.Mana{cost.O(1), cost.W}, types.Instant,
		[]color.Color{color.White}, []game.TargetSpec{{
			MinTargets: 0,
			MaxTargets: 2,
			Constraint: "up to two target creatures",
			Allow:      game.TargetAllowPermanent,
			Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		}},
		game.Instruction{Primitive: game.Tap{
			Object: game.AllTargetPermanentsReference(0),
		}},
	)
}

func newPopQuiz() *game.CardDef {
	const text = "Draw a card.\n" +
		"Learn. (You may reveal a Lesson card you own from outside the game and put it into your hand, or discard a card to draw a card.)"
	return newLearnSpell(
		"Pop Quiz", text, cost.Mana{cost.O(2), cost.U}, types.Instant,
		[]color.Color{color.Blue}, nil,
		game.Instruction{Primitive: game.Draw{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
		}},
	)
}

func newLearnSpell(
	name, text string,
	manaCost cost.Mana,
	cardType types.Card,
	colors []color.Color,
	targets []game.TargetSpec,
	effects ...game.Instruction,
) *game.CardDef {
	sequence := append(effects, game.Instruction{Primitive: game.DiscardThenDraw{
		Player: game.ControllerReference(),
		Max:    1,
	}})
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(colors...),
		CardFace: game.CardFace{
			Name:       name,
			ManaCost:   opt.Val(manaCost),
			Types:      []types.Card{cardType},
			Colors:     colors,
			OracleText: text,
			SpellAbility: opt.Val(game.Mode{
				Text:     text,
				Targets:  targets,
				Sequence: sequence,
			}.Ability()),
		},
	}
}

func learnAnyTarget() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "any target",
		Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
		Selection: opt.Val(game.Selection{
			RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle},
		}),
	}
}
