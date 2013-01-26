package main

import (
  "fmt"
  "gorecipes/allrecipes"
)

func main() {
  fmt.Println("Starting..");

  arReader := allrecipes.NewReader()
  for {
    fmt.Printf("%s\n", <-arReader);
  }
}
