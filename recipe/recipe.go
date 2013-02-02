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
  Directions []string
}

func (r *Recipe) String() string {
  return fmt.Sprintf("Recipe: %s\n" + "\tLink: %s\n" + "\tImageLink: %s\n" +
                     "\tRating: %s\n" + "\tReadyTime: %sh %sm\n" +
                     "\tCookTime: %sh %sm\n" + "\tIngredients: %s\n" +
                     "\tDirections: %s\n", r.Name, r.Link, r.ImageLink,
                     r.Rating, r.ReadyTimeHours, r.ReadyTimeMins,
                     r.CookTimeHours, r.CookTimeMins, r.Ingredients,
                     r.Directions)
}
