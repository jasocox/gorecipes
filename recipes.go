package main

import (
  "fmt"
  "gorecipes/allrecipes"
)

func main() {
  fmt.Println("Starting..");

  arReader := allrecipes.NewReader()
  for {
    fmt.Printf("\t\t%s\n", <-arReader);
  }
}
