package recipe

import "fmt"

type Recipe struct {
  Name string
  Link string
  ImageLink string
  Rating string
  ReadyTimeHours string
  ReadyTimeMins string
  CookTimeHours string
  CookTimeMins string
  Ingredients []string
  Amounts []string
  Directions []string
}

func ingredientsListWithAmounts(ingredients []string, amounts []string) string {
  retVal := "Ingredients:\n"

  for i := range ingredients {
    retVal += fmt.Sprintf("\t\t%s %s\n", amounts[i], ingredients[i])
  }

  return retVal
}

func (r *Recipe) String() string {
  return fmt.Sprintf("Recipe: %s\n" + "\tLink: %s\n" + "\tImageLink: %s\n" +
                     "\tRating: %s\n" + "\tReadyTime: %sh %sm\n" +
                     "\tCookTime: %sh %sm\n" + "\t%s\n" + "\tDirections: %s\n",
                     r.Name, r.Link, r.ImageLink, r.Rating, r.ReadyTimeHours,
                     r.ReadyTimeMins, r.CookTimeHours, r.CookTimeMins,
                     ingredientsListWithAmounts(r.Ingredients, r.Amounts),
                     r.Directions)
}
