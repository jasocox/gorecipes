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
  AmountsAndIngredients [][2]string
  Directions []string
}

func ingredientsListWithAmounts(amountsAndIngredients [][2]string) string {
  retVal := "Ingredients:\n"

  for i := range amountsAndIngredients {
    retVal += fmt.Sprintf("\t\t%s %s\n", amountsAndIngredients[i][0], amountsAndIngredients[i][1])
  }

  return retVal
}

func directions(directions []string) string {
  retVal := "Directions:\n"

  for i := range directions {
    retVal += fmt.Sprintf("\t\t%d) %s\n", i+1, directions[i])
  }

  return retVal
}

func (r *Recipe) String() string {
  return fmt.Sprintf("Recipe: %s\n" + "\tLink: %s\n" + "\tImageLink: %s\n" +
                     "\tRating: %s\n" + "\tReadyTime: %sh %sm\n" +
                     "\tCookTime: %sh %sm\n" + "\t%s\n" + "\t%s\n",
                     r.Name, r.Link, r.ImageLink, r.Rating, r.ReadyTimeHours,
                     r.ReadyTimeMins, r.CookTimeHours, r.CookTimeMins,
                     ingredientsListWithAmounts(r.AmountsAndIngredients),
                     directions(r.Directions))
}
