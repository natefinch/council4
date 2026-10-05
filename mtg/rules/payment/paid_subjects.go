package payment

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
)

type costSubjectSelection struct {
	key      string
	kind     game.PaidCostKind
	objectID id.ID
	cardID   id.ID
	random   bool
	amount   int
}

func (p *additionalCostPlan) recordSacrificeSubjects(key string, permanents []*game.Permanent) {
	if key == "" {
		return
	}
	for _, permanent := range permanents {
		p.subjects = append(p.subjects, costSubjectSelection{
			key: key, kind: game.PaidCostSacrifice, objectID: permanent.ObjectID,
		})
	}
}

func (p *additionalCostPlan) recordDiscardSubjects(key string, cards []id.ID) {
	if key == "" {
		return
	}
	for _, cardID := range cards {
		p.subjects = append(p.subjects, costSubjectSelection{
			key: key, kind: game.PaidCostDiscard, cardID: cardID,
		})
	}
}

func (p *additionalCostPlan) sacrificeSubjectKey(objectID id.ID) string {
	for _, subject := range p.subjects {
		if subject.kind == game.PaidCostSacrifice && subject.objectID == objectID {
			return subject.key
		}
	}
	return ""
}

func (p *additionalCostPlan) discardSubjectKey(cardID id.ID) string {
	for _, subject := range p.subjects {
		if subject.kind == game.PaidCostDiscard && !subject.random && subject.cardID == cardID {
			return subject.key
		}
	}
	return ""
}

func (p *additionalCostPlan) randomDiscardSubjectKeys() []string {
	var keys []string
	for _, subject := range p.subjects {
		if subject.random {
			for range subject.amount {
				keys = append(keys, subject.key)
			}
		}
	}
	return keys
}
