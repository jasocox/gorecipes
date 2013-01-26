package recipe

import "fmt"

type Recipe struct {
  Name string
  Link string
}

func (r *Recipe) String() string {
  return fmt.Sprintf("Recipe: %s\n\tLink: %s\n", r.Name, r.Link)
}
