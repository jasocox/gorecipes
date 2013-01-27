package recipe

import "fmt"

type Recipe struct {
  Name string
  Link string
  ImageLink string
  Rating int
  ReviewsLink string
  ReadyTime string
  CookTime string
  Ingredients []string
  Directions []string
}

func (r *Recipe) String() string {
  return fmt.Sprintf("Recipe: %s\n" + "\tLink: %s\n" + "\tImageLink: %s\n" +
                     "\tRating: %d\n" + "\tReviewsLink: %s\n" + "\tReadyTime: %s\n" +
                     "\tCookTime: %s\n" + "\tIngredients: %s\n" +
                     "\tDirections: %s\n", r.Name, r.Link, r.ImageLink,
                     r.Rating, r.ReviewsLink, r.ReadyTime, r.CookTime,
                     r.Ingredients, r.Directions)
}
