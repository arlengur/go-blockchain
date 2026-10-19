package main

type SendMessage struct {
	From, To string
	Value    int
}

type BalanceMessage struct {
	Address string
}
