package rules

import (
	"math/rand/v2"
	"slices"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules/payment"
)

// maxManaPaymentProbeRuns bounds the payment-time mana ability activations
// manualManaMayPay tries before it stops searching.
const maxManaPaymentProbeRuns = 128

// manualManaMayPay reports whether the caster of request may activate
// payment-time mana abilities (CR 601.2g) so that the locked total prepared
// becomes payable (CR 601.2h). On clones of g it tries each sequence of the
// caster's legal manual mana abilities with each of their choices, checking
// after every activation whether the total can then be paid from the pool and
// the planner's automatic sources. When the search can't be completed, because
// it ran out of activations, met a choice it can't enumerate, or activated an
// ability whose outcome depends on information hidden from the caster, it
// reports true: the proposal is legal to begin and is reversed if it can't be
// paid (CR 733.1). Offering or withholding the cast must not reveal hidden
// information such as the identity or order of library cards (CR 401.2), so
// no answer is drawn from a state that saw it.
func (e *Engine) manualManaMayPay(g *game.Game, request payment.SpellRequest, prepared payment.PreparedSpellCost) bool {
	root := g.Clone()
	root.Execution = game.ExecutionDiagnostics{}
	seed := g.IDGen.Current()
	probe := &manaPaymentProbe{
		engine: &Engine{
			rng:                 rand.New(rand.NewPCG(seed, seed^searchSeedLo)),
			searchRNG:           rand.New(rand.NewPCG(seed^searchSeedHi, seed)),
			cardImplementations: e.cardImplementations,
			programs:            e.programs,
			DecisionLimit:       e.DecisionLimit,
		},
		player: request.PlayerID,
		payable: func(clone *game.Game) bool {
			payable, _ := preparedSpellCostFeasibility(clone, request, prepared, false, nil)
			return payable
		},
	}
	return probe.search(root) || probe.inconclusive
}

// manaPaymentProbe searches for payment-time mana ability activations after
// which payable holds (see manualManaMayPay). Its engine has its own random
// streams so probing never advances the live game's.
type manaPaymentProbe struct {
	engine       *Engine
	player       game.PlayerID
	payable      func(*game.Game) bool
	runs         int
	inconclusive bool
}

// search reports whether some sequence of activations from g makes the cost
// payable. Once the probe is inconclusive its answer is settled, so it stops.
func (p *manaPaymentProbe) search(g *game.Game) bool {
	for _, candidate := range p.engine.paymentManaAbilityActions(g, p.player) {
		activation, _ := candidate.ActivateAbilityPayload()
		var path []int
		for {
			if p.runs >= maxManaPaymentProbeRuns {
				p.inconclusive = true
			}
			if p.inconclusive {
				return false
			}
			p.runs++
			clone := g.Clone()
			chooser := &manaProbeChooser{player: p.player, path: path}
			hidden := hiddenStateOf(clone, p.player)
			activated := p.activate(clone, activation, chooser)
			if activated && hidden.seen(clone, p.player) {
				p.inconclusive = true
				return false
			}
			if activated && (p.payable(clone) || p.search(clone)) {
				return true
			}
			p.inconclusive = p.inconclusive || chooser.unenumerable
			next, ok := chooser.next()
			if !ok {
				break
			}
			path = next
		}
	}
	return false
}

// hiddenState records the information hidden from viewer before an activation:
// every library (CR 401.2), the hands viewer can't see and the face-down cards
// in exile, plus how many events had happened.
type hiddenState struct {
	libraries [game.NumPlayers]zone.Zone
	cards     [][]id.ID
	events    int
}

func hiddenStateOf(g *game.Game, viewer game.PlayerID) hiddenState {
	state := hiddenState{events: len(g.Events)}
	for owner := range state.libraries {
		state.libraries[owner] = g.Players[owner].Library.Clone()
	}
	state.cards = hiddenCards(g, viewer)
	return state
}

func hiddenCards(g *game.Game, viewer game.PlayerID) [][]id.ID {
	var cards [][]id.ID
	for _, player := range g.Players {
		if !g.CanSeePlayerInformation(viewer, player.ID) {
			cards = append(cards, player.Hand.All())
		}
		var faceDown []id.ID
		for _, cardID := range player.Exile.All() {
			if player.Exile.IsFaceDown(cardID) {
				faceDown = append(faceDown, cardID)
			}
		}
		cards = append(cards, faceDown)
	}
	return cards
}

// seen reports whether an activation in g since the state was recorded moved,
// reordered, shuffled, drew, revealed or searched information hidden from
// viewer, so its outcome may depend on that information.
func (h hiddenState) seen(g *game.Game, viewer game.PlayerID) bool {
	if h.events > len(g.Events) || !manaActivationIsReversible(g, h.libraries, h.events) ||
		!slices.EqualFunc(h.cards, hiddenCards(g, viewer), slices.Equal) {
		return true
	}
	for _, event := range g.Events[h.events:] {
		switch event.Kind {
		case game.EventCardRevealed, game.EventCardDrawn, game.EventScry, game.EventSurveil, game.EventLibrarySearched:
			return true
		}
	}
	return false
}

func (p *manaPaymentProbe) activate(g *game.Game, activation action.ActivateAbilityAction, chooser *manaProbeChooser) bool {
	g.ManaPaymentScopes = append(g.ManaPaymentScopes, p.player)
	defer func() { g.ManaPaymentScopes = g.ManaPaymentScopes[:len(g.ManaPaymentScopes)-1] }()
	var agents [game.NumPlayers]PlayerAgent
	for i := range agents {
		agents[i] = chooser
	}
	log := &TurnLog{}
	p.engine.setReplacementChoiceContext(g, agents, log)
	defer g.ClearChoiceContext()
	return p.engine.applyActivateAbilityWithChoices(g, p.player, activation, agents, log)
}

// manaProbeChooser answers the choices made while activating one mana ability
// in a probe: the caster's single-selection choices follow path, then take
// their first option, and next returns the path of the following combination.
// Any other choice takes the engine's fallback and marks the probe
// unenumerable.
type manaProbeChooser struct {
	player       game.PlayerID
	path         []int
	chosen       []int
	widths       []int
	unenumerable bool
}

func (*manaProbeChooser) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (c *manaProbeChooser) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Player != c.player || request.MaxChoices != 1 || request.MinChoices > 1 || len(request.Options) == 0 ||
		request.Kind == game.ChoiceNumber || request.Kind == game.ChoiceOrder {
		c.unenumerable = true
		return fallbackChoice(request)
	}
	width := len(request.Options)
	if request.MinChoices == 0 {
		width++
	}
	position := 0
	if len(c.chosen) < len(c.path) {
		position = c.path[len(c.chosen)]
	}
	c.chosen = append(c.chosen, position)
	c.widths = append(c.widths, width)
	if position == len(request.Options) {
		return []int{}
	}
	return []int{request.Options[position].Index}
}

func (c *manaProbeChooser) next() ([]int, bool) {
	for i := len(c.chosen) - 1; i >= 0; i-- {
		if c.chosen[i]+1 < c.widths[i] {
			return append(slices.Clone(c.chosen[:i]), c.chosen[i]+1), true
		}
	}
	return nil, false
}
