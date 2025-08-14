package main

import (
	"fmt"
	"testing"
)

func TestM(t *testing.T) {
	f, err := ReadFileData("/Users/xukai/Downloads/cs-data/TGT Sales Data - Zip Code.xlsx")
	fmt.Println(f, err)

}
