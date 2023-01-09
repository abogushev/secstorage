package model

import (
	"fmt"
)

type BankCard struct {
	Number  string `json:"number,omitempty"`
	Until   string `json:"until"`
	Name    string `json:"name,omitempty"`
	Surname string `json:"surname,omitempty"`
}

func (b *BankCard) Print(description string) string {
	return fmt.Sprintf(`
number:		%v 
until: 		%v
name: 		%v
surname: 	%v
description:%v 
`,
		b.Number,
		b.Until,
		b.Name,
		b.Surname,
		description,
	)
}

func NewBankCard(number string, until string, name string, surname string) *BankCard {
	return &BankCard{Number: number, Until: until, Name: name, Surname: surname}
}
