package sequence

type mode byte

const (
	_ mode = iota
	modeNormal
	modeFastgap
	modeBuffer
)

func (m mode) String() string {
	switch m {
	case modeNormal:
		return "normal"
	case modeFastgap:
		return "fastgap"
	case modeBuffer:
		return "buffer"
	default:
		return "unknown"
	}
}
