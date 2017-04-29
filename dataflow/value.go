package dataflow

type Value struct {
	Device        *Device
	Name          string
	Value         float64
	Unit          string
	RoundDecimals int
}

type ValueMap map[string]Value

type ValueEssential struct {
	Value float64
	Unit  string
}

type ValueEssentialMap map[string]ValueEssential

func (valueMap ValueMap) ConvertToEssential() (valueEssentialMap ValueEssentialMap) {
	valueEssentialMap = make(ValueEssentialMap, len(valueMap))
	for i, v := range valueMap {
		valueEssentialMap[i] = ValueEssential{
			Value: v.Value,
			Unit:  v.Unit,
		}
	}
	return
}
