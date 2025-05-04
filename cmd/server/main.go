package main

import (
	"fmt"

	cmd "github.com/yash3004/user_management_service/cmd"
)

func main(){
	//getting the configurations 
	cfg := cmd.GetConfigurations()
	fmt.Print(cfg)

	
}  




