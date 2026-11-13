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
// becomes payable with the Phyrexian payments announced (CR 601.2b, 601.2h).
// On clones of g it tries each sequence of the caster's legal manual mana
// abilities with each of their choices, checking after every activation
// whether the total can then be paid from the pool and the planner's automatic
// sources. When the search can't be completed, because
// it ran out of activations, met a choice it can't enumerate, or activated an
// ability whose outcome depends on information hidden from the caster, it
// reports true: the proposal is legal to begin and is reversed if it can't be
// paid (CR 733.1). Offering or withholding the cast must not reveal hidden
// information such as the identity or order of library cards (CR 401.2), so
// no answer is drawn from a state that saw it.
//
// Those abilities are activated only after the card has moved to the stack
// (CR 601.2a), so when g still holds the proposed card in its source zone, as
// when a cast is offered or a total locked before the move, the search starts
// from a clone with the card staged on the stack: a mana ability can't
// discard, exile or otherwise pay its cost with the spell being cast. A copy
// cast, or a state where the card already moved, is searched as it is.
func (e *Engine) manualManaMayPay(g *game.Game, request payment.SpellRequest, prepared payment.PreparedSpellCost, announced map[int]bool) bool {
	root := g.Clone()
	root.Execution = game.ExecutionDiagnostics{}
	seed := g.IDGen.Current()
	if request.CardID != 0 {
		stageProposedSpell(root, request.PlayerID, request.CardID, request.SourceZone, proposedSpellFace(root, request), request.XValue)
	}
	return e.manualManaMayMakePayable(root, seed, request.PlayerID, func(clone *game.Game) bool {
		payable, _ := preparedSpellCostFeasibility(clone, request, prepared, nil, announced)
		return payable
	})
}

// manualActivationManaMayPay reports whether player may activate payment-time
// mana abilities while activating an ability (CR 602.2b, 601.2g) so that its
// cost becomes payable. payable checks the cost in a clone of g, resolving the
// ability's source there by its identity. The search and its handling of
// hidden information are manualManaMayPay's. The ability's source stays where
// it is while its cost is paid, so the search starts from g as it is.
func (e *Engine) manualActivationManaMayPay(g *game.Game, player game.PlayerID, payable func(*game.Game) bool) bool {
	root := g.Clone()
	root.Execution = game.ExecutionDiagnostics{}
	return e.manualManaMayMakePayable(root, g.IDGen.Current(), player, payable)
}

// manualManaMayMakePayable searches root, a clone of the live game whose ID
// generator was at seed, for payment-time mana ability activations by player
// after which payable holds, reporting true when the search is inconclusive.
func (e *Engine) manualManaMayMakePayable(root *game.Game, seed uint64, player game.PlayerID, payable func(*game.Game) bool) bool {
	probe := &manaPaymentProbe{
		engine: &Engine{
			rng:                 rand.New(rand.NewPCG(seed, seed^searchSeedLo)),
			searchRNG:           rand.New(rand.NewPCG(seed^searchSeedHi, seed)),
			cardImplementations: e.cardImplementations,
			programs:            e.programs,
			DecisionLimit:       e.DecisionLimit,
		},
		player:  player,
		payable: payable,
	}
	return probe.search(root) || probe.inconclusive
}

// proposedSpellFace is the face of the proposed card whose characteristics
// request casts, or its front face when none of its faces carries that name. A
// prototyped cast shares its front face's name and is recognized by its
// prototype mana cost (CR 719.3b).
func proposedSpellFace(g *game.Game, request payment.SpellRequest) game.FaceIndex {
	card, ok := g.GetCardInstance(request.CardID)
	if !ok || card.Def == nil || request.Card == nil {
		return game.FaceFront
	}
	if face, ok := card.Def.FaceView(game.FacePrototype); ok && face.Name == request.Card.Name &&
		request.Card.ManaCost.Exists && slices.Equal(request.Card.ManaCost.Val, face.ManaCost.Val) &&
		!slices.Equal(request.Card.ManaCost.Val, card.Def.ManaCost.Val) {
		return game.FacePrototype
	}
	faces := card.Def.FaceIndexes()
	if card.Def.CanFuse() {
		faces = append(faces, game.FaceFused)
	}
	for _, index := range faces {
		if face, ok := card.Def.FaceView(index); ok && face.Name == request.Card.Name {
			return index
		}
	}
	return game.FaceFront
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
