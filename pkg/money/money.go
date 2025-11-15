package money

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/shopspring/decimal"
)

type Money decimal.Decimal

func NewMoneyFromFloat(f float64) Money {
	return Money(decimal.NewFromFloat(f))
}

func NewMoneyFromString(s string) (Money, error) {
	d, err := decimal.NewFromString(s)
	return Money(d), err
}

func (m Money) Decimal() decimal.Decimal { return decimal.Decimal(m) }

func (m Money) Float64() float64 { return m.Decimal().InexactFloat64() }

func (m Money) String() string { return m.Decimal().String() }

func (m Money) Equal(other Money) bool { return m.Decimal().Equal(other.Decimal()) }

func (m *Money) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		if f, ferr := strconv.ParseFloat(s, 64); ferr == nil {
			d = decimal.NewFromFloat(f)
		} else {
			return fmt.Errorf("invalid money: %w", err)
		}
	}
	*m = Money(d)
	return nil
}

func (m Money) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())
}

func (m Money) Value() (driver.Value, error) {
	return m.String(), nil
}

func (m *Money) Scan(value interface{}) error {
	if value == nil {
		*m = Money(decimal.NewFromInt(0))
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into Money", value)
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return err
	}
	*m = Money(d)
	return nil
}
