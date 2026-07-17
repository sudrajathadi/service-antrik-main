package chatbot

type Engine struct {
	evaluator *Evaluator
}

func NewEngine(evaluator *Evaluator) *Engine {
	return &Engine{evaluator: evaluator}
}

func (e *Engine) Reply(req ChatRequest) (ChatResponse, error) {
	tokens := Tokenize(req.Message)
	parsed := Parse(req.Message, tokens)
	intent, confidence := Translate(parsed)

	return e.evaluator.Evaluate(req, parsed, intent, confidence)
}
