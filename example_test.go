package wise_test

import (
	"fmt"

	"github.com/larsartmann/wise-go"
)

func ExampleNewCurrency() {
	currency, err := wise.NewCurrency("EUR")
	if err != nil {
		panic(err)
	}

	fmt.Println(currency)
	// Output: EUR
}

func ExampleMoney_String() {
	m := wise.Money{Cents: 1234, Currency: wise.Currency("EUR")}
	fmt.Println(m)
	// Output: EUR 12.34
}

func ExampleMoney_String_negative() {
	m := wise.Money{Cents: -5000, Currency: wise.Currency("USD")}
	fmt.Println(m)
	// Output: USD -50.00
}
