// Code generated by vogen DO NOT EDIT.
package sample

func NewAddress(number int, number2 int, city string, country string) (*Address, error) {

	tempVarByVogen0, err := NewAddressNumber(number)
	if err != nil {
		return nil, err
	}

	tempVarByVogen1Base, err := NewAddressNumber(number2)
	if err != nil {
		return nil, err
	}
	tempVarByVogen1, err := NewAddressNumber2(tempVarByVogen1Base)
	if err != nil {
		return nil, err
	}

	tempVarByVogen2, err := NewAddressCity(city)
	if err != nil {
		return nil, err
	}

	tempVarByVogen3, err := NewAddressCountry(country)
	if err != nil {
		return nil, err
	}

	return &Address{

		Number: tempVarByVogen0,

		Number2: tempVarByVogen1,

		City: tempVarByVogen2,

		Country: tempVarByVogen3,
	}, nil
}

type rawAddress struct {
	Number int

	Number2 int

	City string

	Country string
}

func (d Address) RawValue() rawAddress {
	return rawAddress{

		Number: d.Number.RawValue(),

		Number2: d.Number2.RawValue().RawValue(),

		City: d.City.RawValue(),

		Country: d.Country.RawValue(),
	}
}

func (d AddressNumber) RawValue() int {
	return int(d)
}

func (d AddressNumber2) RawValue() AddressNumber {
	return AddressNumber(d)
}

func (d AddressCity) RawValue() string {
	return string(d)
}

func (d AddressCountry) RawValue() string {
	return string(d)
}
