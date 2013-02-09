package main

import (
  "fmt"
  "gorecipes/allrecipes"
)

func main() {
  fmt.Println("Starting..")

  arReader, messageBox := allrecipes.NewRecipeReader()
  count := 0
  continueReading := true
  for continueReading {
    select {
    case r := <-arReader:
      count++
      fmt.Printf("%d: %s\n", count, r);
    case message := <-messageBox:
      if message == "done" {
        continueReading = false
      }
    }
  }

  fmt.Println("Done.")
}
