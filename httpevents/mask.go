package httpevents


type LoggerMask uint32

const (
	LoggerMaskReqHeader   = LoggerMask(int64(1) << iota)
	LoggerMaskResHeader = LoggerMask(int64(1) << iota)
	LoggerMaskPath      = LoggerMask(int64(1) << iota)

	LoggerMaskAll = LoggerMask(0xffffffff)
)

type LoggerMaskBuilder struct{
	mask LoggerMask
}

func NewLoggerMaskBuilder() LoggerMaskBuilder {
	return LoggerMaskBuilder{
		mask: LoggerMaskAll,
	}
}

func (b LoggerMaskBuilder) Exclude(mask LoggerMask) *LoggerMaskBuilder {
	b.mask ^= mask

	return b
}

func (b LoggerMaskBuilder) Include(mask LoggerMask) *LoggerMaskBuilder {
	b.mask |= mask

	return b
}

func (b LoggerMaskBuilder) Build() LoggerMask {
	return b.mask
}