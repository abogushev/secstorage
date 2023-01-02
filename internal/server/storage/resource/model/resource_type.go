package model

type ResourceType uint

const (
	Undefined ResourceType = iota
	LoginPassword
	File
	BankCard
)
