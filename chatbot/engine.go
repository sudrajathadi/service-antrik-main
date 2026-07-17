package chatbot

type EvaluatorRunner interface {
	Evaluate(req ChatRequest, parsed ParseResult, intent Intent, confidence float64) (ChatResponse, error)
}

type Engine struct {
	evaluator EvaluatorRunner
}

func NewEngine(evaluator EvaluatorRunner) *Engine {
	return &Engine{evaluator: evaluator}
}

func (e *Engine) Reply(req ChatRequest) (ChatResponse, error) {
	tokens := Tokenize(req.Message)
	parsed := Parse(req.Message, tokens)
	intent, confidence := Translate(parsed)

	return e.evaluator.Evaluate(req, parsed, intent, confidence)
}
