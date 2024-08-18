// Code generated by vogen DO NOT EDIT.
package sample

func NewAddress(number int, number2 int, number2p *int, city string, country string) (*Address, error) {

	tempVarByVogen0, err := NewAddressNumber(number)
	if err != nil {
		return nil, err
	}

	tempVarByVogen1, err := NewAddressNumber2(number2)
	if err != nil {
		return nil, err
	}

	tempVarByVogen2, err := NewAddressNumber2p(number2p)
	if err != nil {
		return nil, err
	}

	tempVarByVogen3, err := NewAddressCity(city)
	if err != nil {
		return nil, err
	}

	tempVarByVogen4, err := NewAddressCountry(country)
	if err != nil {
		return nil, err
	}

	return &Address{

		Number: tempVarByVogen0,

		Number2: tempVarByVogen1,

		Number2p: tempVarByVogen2,

		City: tempVarByVogen3,

		Country: tempVarByVogen4,
	}, nil

}

type rawAddress struct {
	Number int

	Number2 int

	Number2p *int

	City string

	Country string
}

func (d Address) RawValue() rawAddress {

	var tempVarByVogenNumber2p *int
	if d.Number2p != nil {
		tempVarByVogenNumber2p = (*int)((*AddressNumber)(d.Number2p))
	}

	return rawAddress{

		Number: d.Number.RawValue(),

		Number2: d.Number2.RawValue().RawValue(),

		Number2p: tempVarByVogenNumber2p,

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
