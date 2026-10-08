// Package rules is the deterministic implementation of the ConversationAI
// port: keyword intents, numbered choices and fuzzy medicine matching. An LLM
// backed adapter can replace it without touching the core.
package rules

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/farmaenlace/farmi/internal/core/domain"
)

// Engine implements ports.ConversationAI.
type Engine struct{}

// New builds the rules engine.
func New() *Engine { return &Engine{} }

var choiceRe = regexp.MustCompile(`^(?:opcion\s*|la\s*|el\s*)?(\d{1,2})\b`)

// intentPhrases are checked in order: the more specific commands come first so
// that "cancelar pedido" never reads as a greeting or a "no".
var intentPhrases = []struct {
	intent  domain.Intent
	phrases []string
}{
	{domain.IntentRestart, []string{"reiniciar", "nueva compra", "empezar de nuevo", "comenzar de nuevo", "reset"}},
	{domain.IntentHelp, []string{"ayuda", "help", "?"}},
	{domain.IntentStatus, []string{"estado", "mi pedido", "status", "como va"}},
	{domain.IntentCancel, []string{"cancelar", "cancela", "anular"}},
	{domain.IntentContinue, []string{"continuar", "renovar", "reintentar", "seguir"}},
	{domain.IntentPayLink, []string{"link", "enlace", "pagar", "pago"}},
	{domain.IntentPickup, []string{"retiro", "retirar", "farmacia", "recoger", "recojo"}},
	{domain.IntentDelivery, []string{"domicilio", "entrega", "casa", "delivery", "envio"}},
	{domain.IntentYes, []string{"si", "confirmo", "confirmar", "ok", "dale", "claro", "de acuerdo", "acepto", "listo", "yes"}},
	{domain.IntentNo, []string{"no", "cambiar", "otra", "otras", "nop"}},
	{domain.IntentGreeting, []string{"hola", "buenas", "buenos dias", "hey", "hi", "hello", "ola"}},
}

// Interpret reads a free-text reply.
func (e *Engine) Interpret(_ context.Context, text string) domain.Interpretation {
	t := normalize(text)
	if t == "" {
		return domain.Interpretation{Intent: domain.IntentFreeText}
	}
	if m := choiceRe.FindStringSubmatch(t); m != nil {
		n, _ := strconv.Atoi(m[1])
		return domain.Interpretation{Intent: domain.IntentChoice, Choice: n}
	}
	for _, ip := range intentPhrases {
		for _, ph := range ip.phrases {
			if containsPhrase(t, ph) {
				return domain.Interpretation{Intent: ip.intent}
			}
		}
	}
	return domain.Interpretation{Intent: domain.IntentFreeText}
}

var zoneStopwords = map[string]bool{"zona": true, "de": true, "la": true, "el": true, "en": true, "quiero": true, "comprar": true, "por": true, "favor": true}

// MatchZone accepts the option number or words of the zone label.
func (e *Engine) MatchZone(_ context.Context, text string, zones []domain.Zone) (domain.Zone, bool) {
	t := normalize(text)
	if m := choiceRe.FindStringSubmatch(t); m != nil {
		n, _ := strconv.Atoi(m[1])
		if n >= 1 && n <= len(zones) {
			return zones[n-1], true
		}
		return domain.Zone{}, false
	}
	var tokens []string
	for _, tok := range strings.Fields(t) {
		if !zoneStopwords[tok] {
			tokens = append(tokens, tok)
		}
	}
	best, bestScore, tie := domain.Zone{}, 0, false
	for _, z := range zones {
		label := normalize(z.Label)
		score := 0
		for _, tok := range tokens {
			if containsPhrase(label, tok) {
				score++
			}
		}
		switch {
		case score > bestScore:
			best, bestScore, tie = z, score, false
		case score == bestScore && score > 0:
			tie = true
		}
	}
	if bestScore == 0 || tie {
		return domain.Zone{}, false
	}
	return best, true
}

// MatchMedicine maps an OCR line onto a catalog medicine by name and strength.
func (e *Engine) MatchMedicine(_ context.Context, item domain.PrescribedItem, medicines []domain.Medicine) (domain.Medicine, bool) {
	name := normalize(item.Medicine)
	strength := digits(item.Concentration)
	var byName []domain.Medicine
	for _, m := range medicines {
		mn := normalize(m.Name)
		if mn == name || strings.Contains(name, mn) || strings.Contains(mn, name) {
			byName = append(byName, m)
		}
	}
	for _, m := range byName {
		if strength == "" || digits(m.Concentration) == strength {
			return m, true
		}
	}
	return domain.Medicine{}, false
}

var accents = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n", "—", " ", "-", " ", "–", " ")

func normalize(s string) string {
	s = accents.Replace(strings.ToLower(strings.TrimSpace(s)))
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == ' ', r == '?':
			b.WriteRune(r)
		default:
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func containsPhrase(text, phrase string) bool {
	if phrase == "?" {
		return strings.Contains(text, "?")
	}
	return strings.Contains(" "+text+" ", " "+phrase+" ")
}

func digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
