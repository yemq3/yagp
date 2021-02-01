package main

import "log"

func main(){
	a := 5
	b := a
	log.Printf("address of a: %v", &a)
	log.Printf("address of b: %v", &b)
}
