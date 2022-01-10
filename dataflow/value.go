package dataflow

type Value struct {
	DeviceName    string
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
		valueEssentialMap[i] = v.ConvertToEssential()
	}
	return
}

func (value Value) ConvertToEssential() ValueEssential {
	return ValueEssential{
		Value: value.Value,
		Unit:  value.Unit,
	}
}
