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
	// Pipeline chatbot rule-based:
	// 1. Tokenize: pecah teks user menjadi token sederhana.
	// 2. Parse: ambil entity penting seperti dokter, rumah sakit, kota, tanggal, jam.
	// 3. Translate: tentukan intent berdasarkan rule.
	// 4. Evaluate: jalankan intent dan simpan state percakapan bila perlu.
	tokens := Tokenize(req.Message)
	parsed := Parse(req.Message, tokens)
	intent, confidence := Translate(parsed)

	return e.evaluator.Evaluate(req, parsed, intent, confidence)
}
