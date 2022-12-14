package build

import (
	"fmt"
	"math"
	"strings"

	"github.com/hntrl/lang/language/tokens"
)

type NilLiteral struct{}

func (nl NilLiteral) ClassName() string {
	return "<nil>"
}
func (nl NilLiteral) Constructors() ConstructorMap {
	return NewConstructorMap()
}

func (nl NilLiteral) Class() Class {
	return nl
}
func (nl NilLiteral) Value() interface{} {
	return nil
}
func (nl NilLiteral) Set(key string, new ValueObject) error {
	return fmt.Errorf("cannot set property %s of %s", key, nl.Class().ClassName())
}
func (nl NilLiteral) Get(key string) Object {
	return nil
}

type Boolean struct{}

func (bl Boolean) ClassName() string {
	return "Boolean"
}
func (bl Boolean) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddConstructor(Boolean{}, func(obj ValueObject) (ValueObject, error) {
		return obj.(BooleanLiteral), nil
	})
	return csMap
}

func (bl Boolean) ComparableRules() ComparatorRules {
	rules := NewComparatorRules()
	rules.AddComparator(Boolean{}, tokens.AND, func(a, b ValueObject) (ValueObject, error) {
		return BooleanLiteral(a.(BooleanLiteral) && b.(BooleanLiteral)), nil
	})
	rules.AddComparator(Boolean{}, tokens.OR, func(a, b ValueObject) (ValueObject, error) {
		return BooleanLiteral(a.(BooleanLiteral) || b.(BooleanLiteral)), nil
	})
	return rules
}

func (bl Boolean) Get(key string) Object {
	return nil
}

type BooleanLiteral bool

func (bl BooleanLiteral) Class() Class {
	return Boolean{}
}
func (bl BooleanLiteral) Value() interface{} {
	return bool(bl)
}
func (bl BooleanLiteral) Set(key string, new ValueObject) error {
	return fmt.Errorf("cannot set property %s of %s", key, bl.Class().ClassName())
}
func (bl BooleanLiteral) Get(key string) Object {
	return nil
}

type String struct{}

func (str String) ClassName() string {
	return "String"
}
func (str String) Constructors() ConstructorMap {
	numericConstructor := func(obj ValueObject) (ValueObject, error) {
		return StringLiteral(fmt.Sprintf("%v", obj.Value())), nil
	}
	csMap := NewConstructorMap()
	csMap.AddConstructor(String{}, func(obj ValueObject) (ValueObject, error) {
		return obj, nil
	})
	csMap.AddConstructor(Number{}, numericConstructor)
	csMap.AddConstructor(Double{}, numericConstructor)
	csMap.AddConstructor(Integer{}, numericConstructor)
	csMap.AddConstructor(Float{}, numericConstructor)
	csMap.AddConstructor(Boolean{}, numericConstructor)
	return csMap
}
func (str String) Get(key string) Object {
	return nil
}

type StringLiteral string

func (sl StringLiteral) Class() Class {
	return String{}
}
func (sl StringLiteral) Value() interface{} {
	return string(sl)
}
func (sl StringLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (sl StringLiteral) Get(key string) Object {
	methods := map[string]Function{
		"lower": NewFunction(FunctionOptions{
			Returns: String{},
			Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
				return StringLiteral(strings.ToLower(string(sl))), nil
			},
		}),
		"upper": NewFunction(FunctionOptions{
			Returns: String{},
			Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
				return StringLiteral(strings.ToUpper(string(sl))), nil
			},
		}),
	}
	return methods[key]
}

type Number struct{}

func (num Number) ClassName() string {
	return "Number"
}
func (num Number) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddConstructor(Number{}, func(obj ValueObject) (ValueObject, error) {
		val := obj.Value()
		if intNum, ok := val.(int64); ok {
			val = float64(intNum)
		}
		return NumberLiteral(val.(float64)), nil
	})
	csMap.AddConstructor(Double{}, func(obj ValueObject) (ValueObject, error) {
		return NumberLiteral(obj.(DoubleLiteral)), nil
	})
	csMap.AddConstructor(Integer{}, func(obj ValueObject) (ValueObject, error) {
		return NumberLiteral(obj.(IntegerLiteral)), nil
	})
	csMap.AddConstructor(Float{}, func(obj ValueObject) (ValueObject, error) {
		return NumberLiteral(obj.(FloatLiteral)), nil
	})
	return csMap
}
func (num Number) Get(key string) Object {
	return nil
}

func numComparePredicate(cb func(float64, float64) bool) func(ValueObject, ValueObject) (ValueObject, error) {
	numConstructor := (Number{}).Constructors().Get(Number{})
	return func(a, b ValueObject) (ValueObject, error) {
		na, err := numConstructor(a)
		if err != nil {
			return nil, err
		}
		nb, err := numConstructor(b)
		if err != nil {
			return nil, err
		}
		return BooleanLiteral(cb(float64(na.(NumberLiteral)), float64(nb.(NumberLiteral)))), nil
	}
}

func addNumCompareMap(rules ComparatorRules, class Class) {
	rules.AddComparator(class, tokens.EQUALS, numComparePredicate(func(a, b float64) bool {
		return a == b
	}))
	rules.AddComparator(class, tokens.NOT_EQUALS, numComparePredicate(func(a, b float64) bool {
		return a != b
	}))
	rules.AddComparator(class, tokens.LESS, numComparePredicate(func(a, b float64) bool {
		return a < b
	}))
	rules.AddComparator(class, tokens.GREATER, numComparePredicate(func(a, b float64) bool {
		return a > b
	}))
	rules.AddComparator(class, tokens.LESS_EQUAL, numComparePredicate(func(a, b float64) bool {
		return a <= b
	}))
	rules.AddComparator(class, tokens.GREATER_EQUAL, numComparePredicate(func(a, b float64) bool {
		return a >= b
	}))
}
func numCompareMap() ComparatorRules {
	rules := NewComparatorRules()
	addNumCompareMap(rules, Number{})
	addNumCompareMap(rules, Double{})
	addNumCompareMap(rules, Integer{})
	addNumCompareMap(rules, Float{})
	return rules
}

func numOperatorPredicate(cb func(float64, float64) (ValueObject, error)) OperatorFn {
	numConstructor := (Number{}).Constructors().Get(Number{})
	return func(a, b ValueObject) (ValueObject, error) {
		na, err := numConstructor(a)
		if err != nil {
			return nil, err
		}
		nb, err := numConstructor(b)
		if err != nil {
			return nil, err
		}
		return cb(float64(na.(NumberLiteral)), float64(nb.(NumberLiteral)))
	}
}
func addNumOperatorMap(rules OperatorRules, class Class, fn ConstructorFn) {
	rules.AddOperator(class, tokens.ADD, numOperatorPredicate(func(a, b float64) (ValueObject, error) {
		return fn(NumberLiteral(a + b))
	}))
	rules.AddOperator(class, tokens.SUB, numOperatorPredicate(func(a, b float64) (ValueObject, error) {
		return fn(NumberLiteral(a - b))
	}))
	rules.AddOperator(class, tokens.MUL, numOperatorPredicate(func(a, b float64) (ValueObject, error) {
		return fn(NumberLiteral(a * b))
	}))
	rules.AddOperator(class, tokens.PWR, numOperatorPredicate(func(a, b float64) (ValueObject, error) {
		return fn(NumberLiteral(math.Pow(a, b)))
	}))
	rules.AddOperator(class, tokens.QUO, numOperatorPredicate(func(a, b float64) (ValueObject, error) {
		return fn(NumberLiteral(a / b))
	}))
	rules.AddOperator(class, tokens.REM, numOperatorPredicate(func(a, b float64) (ValueObject, error) {
		return fn(NumberLiteral(math.Mod(a, b)))
	}))
}
func numOperatorMap(fn ConstructorFn) OperatorRules {
	rules := NewOperatorRules()
	addNumOperatorMap(rules, Number{}, fn)
	addNumOperatorMap(rules, Double{}, fn)
	addNumOperatorMap(rules, Integer{}, fn)
	addNumOperatorMap(rules, Float{}, fn)
	return rules
}

func (num Number) ComparableRules() ComparatorRules {
	return numCompareMap()
}
func (num Number) OperatorRules() OperatorRules {
	fn := num.Constructors().Get(Number{})
	return numOperatorMap(fn)
}

type NumberLiteral float64

func (nl NumberLiteral) Class() Class {
	return Number{}
}
func (nl NumberLiteral) Value() interface{} {
	return float64(nl)
}
func (nl NumberLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (nl NumberLiteral) Get(key string) Object {
	return nil
}

type Double struct{}

func (db Double) ClassName() string {
	return "Double"
}
func (db Double) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	numericConstructor := func(obj ValueObject) (ValueObject, error) {
		return DoubleLiteral(math.Ceil(obj.Value().(float64)*100) / 100), nil
	}
	csMap.AddConstructor(Number{}, numericConstructor)
	csMap.AddConstructor(Double{}, numericConstructor)
	csMap.AddConstructor(Integer{}, numericConstructor)
	csMap.AddConstructor(Float{}, numericConstructor)
	return csMap
}

func (db Double) ComparableRules() ComparatorRules {
	return numCompareMap()
}
func (db Double) OperatorRules() OperatorRules {
	fn := db.Constructors().Get(Number{})
	return numOperatorMap(fn)
}

func (db Double) Get(key string) Object {
	return nil
}

type DoubleLiteral float64

func (dl DoubleLiteral) Class() Class {
	return Double{}
}
func (dl DoubleLiteral) Value() interface{} {
	return float64(dl)
}
func (dl DoubleLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (dl DoubleLiteral) Get(key string) Object {
	return nil
}

type Float struct{}

func (f Float) ClassName() string {
	return "Float"
}
func (f Float) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddConstructor(Number{}, func(obj ValueObject) (ValueObject, error) {
		return FloatLiteral(obj.Value().(float64)), nil
	})
	csMap.AddConstructor(Double{}, func(obj ValueObject) (ValueObject, error) {
		return FloatLiteral(obj.(DoubleLiteral)), nil
	})
	csMap.AddConstructor(Integer{}, func(obj ValueObject) (ValueObject, error) {
		return FloatLiteral(obj.(IntegerLiteral)), nil
	})
	csMap.AddConstructor(Float{}, func(obj ValueObject) (ValueObject, error) {
		return FloatLiteral(obj.(FloatLiteral)), nil
	})
	return csMap
}

func (f Float) ComparableRules() ComparatorRules {
	return numCompareMap()
}
func (f Float) OperatorRules() OperatorRules {
	fn := f.Constructors().Get(Number{})
	return numOperatorMap(fn)
}

func (f Float) Get(key string) Object {
	return nil
}

type FloatLiteral float64

func (fl FloatLiteral) Class() Class {
	return Float{}
}
func (fl FloatLiteral) Value() interface{} {
	return float64(fl)
}
func (fl FloatLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (fl FloatLiteral) Get(key string) Object {
	return nil
}

type Integer struct{}

func (i Integer) ClassName() string {
	return "Integer"
}
func (i Integer) Constructors() ConstructorMap {
	csMap := NewConstructorMap()
	csMap.AddConstructor(Number{}, func(obj ValueObject) (ValueObject, error) {
		return IntegerLiteral(obj.Value().(float64)), nil
	})
	csMap.AddConstructor(Double{}, func(obj ValueObject) (ValueObject, error) {
		return IntegerLiteral(obj.(DoubleLiteral)), nil
	})
	csMap.AddConstructor(Integer{}, func(obj ValueObject) (ValueObject, error) {
		return IntegerLiteral(obj.(IntegerLiteral)), nil
	})
	csMap.AddConstructor(Float{}, func(obj ValueObject) (ValueObject, error) {
		return IntegerLiteral(obj.(FloatLiteral)), nil
	})
	return csMap
}

func (i Integer) ComparableRules() ComparatorRules {
	return numCompareMap()
}
func (i Integer) OperatorRules() OperatorRules {
	fn := i.Constructors().Get(Number{})
	return numOperatorMap(fn)
}

func (i Integer) Get(key string) Object {
	return nil
}

type IntegerLiteral int64

func (il IntegerLiteral) Class() Class {
	return Integer{}
}
func (il IntegerLiteral) Value() interface{} {
	return int64(il)
}
func (il IntegerLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (il IntegerLiteral) Get(key string) Object {
	return nil
}

type Date struct{}

func (d Date) ClassName() string {
	return "Date"
}
func (d Date) Constructors() ConstructorMap {
	return NewConstructorMap()
}

func (d Date) Get(key string) Object {
	switch key {
	case "now":
		return NewFunction(FunctionOptions{
			Returns: Date{},
			Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
				return DateLiteral{}, nil
			},
		})
	}
	return nil
}

type DateLiteral struct{}

func (dl DateLiteral) Class() Class {
	return Date{}
}
func (dl DateLiteral) Value() interface{} {
	return nil
}
func (dl DateLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (dl DateLiteral) Get(key string) Object {
	return nil
}

type DateTime struct{}

func (d DateTime) ClassName() string {
	return "DateTime"
}
func (d DateTime) Constructors() ConstructorMap {
	return NewConstructorMap()
}

func (d DateTime) Get(key string) Object {
	switch key {
	case "now":
		return NewFunction(FunctionOptions{
			Returns: DateTime{},
			Handler: func(args []ValueObject, proto ValueObject) (ValueObject, error) {
				return DateTimeLiteral{}, nil
			},
		})
	}
	return nil
}

type DateTimeLiteral struct{}

func (dl DateTimeLiteral) Class() Class {
	return DateTime{}
}
func (dl DateTimeLiteral) Value() interface{} {
	return nil
}
func (dl DateTimeLiteral) Set(key string, obj ValueObject) error {
	return nil
}
func (dl DateTimeLiteral) Get(key string) Object {
	return nil
}
