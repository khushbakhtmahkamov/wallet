package main

import (
	"fmt"
	"github.com/khushbakhtmahkamov/wallet/pkg/wallet"
)

func main()  {
	svc:=&wallet.Service{}
	account, err :=svc.RegisterAccount("+992928393813")
	if err != nil {
		fmt.Println(err)
		//return
	}
	fmt.Println(account)
}