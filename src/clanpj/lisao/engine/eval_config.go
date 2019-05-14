package engine

import (
	"fmt"
)

type ConfigParam struct {
	Descr string
	Min int
	Max int
	Delta uint // The max increment/decrement range
	Get func() int
	Set func(val int)
}

var configParams = make([]ConfigParam, 0, 2048) // big enough for anyone!

func GetConfigParams() []ConfigParam {
	return configParams
}

func SetConfigParams(vals []int) {
	for i := 0; i < len(vals); i++ {
		configParams[i].Set(vals[i])
	}
}

func RegisterConfigParamInt(descr string, param *int, min int, max int, delta uint) {
	configParams = append(configParams, ConfigParam{
		Descr: descr,
		Min: min,
		Max: max,
		Delta: delta,
		Get: func() int { return *param },
		Set: func(val int) { *param = val }})
}

const defaultInt8Min = -128
const defaultInt8Max = 127
const defaultInt8Delta = 16

func RegisterConfigParamInt8Default(descr string, param *int8) {
	configParams = append(configParams, ConfigParam{
		Descr: descr,
		Min: defaultInt8Min,
		Max: defaultInt8Max,
		Delta: defaultInt8Delta,
		Get: func() int { return int(*param) },
		Set: func(val int) { *param = int8(val) }})
}

func RegisterConfigParamInt8ArrayDefault(descr string, params []int8) {
	for i := 0; i < len(params); i++ {
		RegisterConfigParamInt8Default(fmt.Sprintf("%s[%d]", descr, i), &params[i])
	}
}

const defaultEvalCpDelta = 16

func RegisterConfigParamEvalCpDefault(descr string, param *EvalCp) {
	configParams = append(configParams, ConfigParam{
		Descr: descr,
		Min: int(BlackCheckMateEval),
		Max: int(WhiteCheckMateEval),
		Delta: defaultEvalCpDelta,
		Get: func() int { return int(*param) },
		Set: func(val int) { *param = EvalCp(val) }})
}
