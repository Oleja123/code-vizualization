package interpreter

type ExecSignal int

const (
	SignalNormal ExecSignal = iota
	SignalBreak
	SignalContinue
	SignalReturn
)

type ExecResult struct {
	Signal ExecSignal
	Value  *int
}

func BreakResult() ExecResult {
	return ExecResult{Signal: SignalBreak}
}

func ContinueResult() ExecResult {
	return ExecResult{Signal: SignalContinue}
}

func NormalResult() ExecResult {
	return ExecResult{Signal: SignalNormal}
}

func ReturnResult(val *int) ExecResult {
	return ExecResult{Signal: SignalReturn, Value: val}
}
