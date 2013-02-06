package main

import (
  "fmt"
  "gorecipes/allrecipes"
)

func main() {
  fmt.Println("Starting..");

  arReader := allrecipes.NewRecipeReader()
  count := 0
  for {
    count++
    fmt.Printf("%d: %s\n", count, <-arReader);
  }
}
