# supabase-go

Unofficial [Supabase](https://supabase.io) client for Go. It is an amalgamation of all the libraries similar to the [official Supabase client](https://supabase.io/docs/reference/javascript/supabase-client).

## Installation
```
go get github.com/nedpals/supabase-go
```

## Usage

Replace the `<SUPABASE-URL>` and `<SUPABASE-URL>` placeholders with values from `https://supabase.com/dashboard/project/YOUR_PROJECT/settings/api`

### Authenticate
```go
package main 
import (
    supa "github.com/nedpals/supabase-go"
    "fmt"
    "context"
)

func main() {
  supabaseUrl := "<SUPABASE-URL>"
  supabaseKey := "<SUPABASE-KEY>"
  supabase := supa.CreateClient(supabaseUrl, supabaseKey)

  ctx := context.Background()
  user, err := supabase.Auth.SignUp(ctx, supa.UserCredentials{
    Email:    "example@example.com",
    Password: "password",
  })
  if err != nil {
    panic(err)
  }

  fmt.Println(user)
}
```

### Sign-In
```go
package main 
import (
    supa "github.com/nedpals/supabase-go"
    "fmt"
    "context"
)

func main() {
  supabaseUrl := "<SUPABASE-URL>"
  supabaseKey := "<SUPABASE-KEY>"
  supabase := supa.CreateClient(supabaseUrl, supabaseKey)

  ctx := context.Background()
  user, err := supabase.Auth.SignIn(ctx, supa.UserCredentials{
    Email:    "example@example.com",
    Password: "password",
  })
  if err != nil {
    panic(err)
  }

  fmt.Println(user)
}
```

### Insert
```go
package main 
import (
    supa "github.com/nedpals/supabase-go"
    "fmt"
)

type Country struct {
  ID      int    `json:"id"`
  Name    string `json:"name"`
  Capital string `json:"capital"`
}

func main() {
  supabaseUrl := "<SUPABASE-URL>"
  supabaseKey := "<SUPABASE-KEY>"
  supabase := supa.CreateClient(supabaseUrl, supabaseKey)

  row := Country{
    ID:      5,
    Name:    "Germany",
    Capital: "Berlin",
  }

  var results []Country
  err := supabase.DB.From("countries").Insert(row).Execute(&results)
  if err != nil {
    panic(err)
  }

  fmt.Println(results) // Inserted rows
}
```

### Select
```go
package main 
import (
    supa "github.com/nedpals/supabase-go"
    "fmt"
)

func main() {
  supabaseUrl := "<SUPABASE-URL>"
  supabaseKey := "<SUPABASE-KEY>"
  supabase := supa.CreateClient(supabaseUrl, supabaseKey)

  var results map[string]interface{}
  err := supabase.DB.From("countries").Select("*").Single().Execute(&results)
  if err != nil {
    panic(err)
  }

  fmt.Println(results) // Selected rows
}
```

### Update
```go
package main 
import (
    supa "github.com/nedpals/supabase-go"
    "fmt"
)

type Country struct {
  Name    string `json:"name"`
  Capital string `json:"capital"`
}

func main() {
  supabaseUrl := "<SUPABASE-URL>"
  supabaseKey := "<SUPABASE-KEY>"
  supabase := supa.CreateClient(supabaseUrl, supabaseKey)

  row := Country{
    Name:    "France",
    Capital: "Paris",
  }

  var results map[string]interface{}
  err := supabase.DB.From("countries").Update(row).Eq("id", "5").Execute(&results)
  if err != nil {
    panic(err)
  }

  fmt.Println(results) // Updated rows
}
```

### Delete
```go
package main 
import (
    supa "github.com/nedpals/supabase-go"
    "fmt"
)

func main() {
  supabaseUrl := "<SUPABASE-URL>"
  supabaseKey := "<SUPABASE-KEY>"
  supabase := supa.CreateClient(supabaseUrl, supabaseKey)

  var results map[string]interface{}
  err := supabase.DB.From("countries").Delete().Eq("name", "France").Execute(&results)
  if err != nil {
    panic(err)
  }

  fmt.Println(results) // Empty - nothing returned from delete
}
```

## Roadmap
- [x] Auth support (1)
- [x] DB support (2)
- [ ] Realtime
- [x] Storage
- [ ] Testing

(1) - Thin API wrapper. Does not rely on the GoTrue library for simplicity
(2) - Through `postgrest-go`

I just implemented features which I actually needed for my project for now. If you like to implement these features, feel free to submit a pull request as stated in the [Contributing](#contributing) section below.

## Design Goals
It tries to mimick as much as possible the official Javascript client library in terms of ease-of-use and in setup process.

# Contributing
## Submitting a pull request
- Fork it (https://github.com/nedpals/supabase-go/fork)
- Create your feature branch (git checkout -b my-new-feature)
- Commit your changes (git commit -am 'Add some feature')
- Push to the branch (git push origin my-new-feature)
- Create a new Pull Request

# Contributors
- [nedpals](https://github.com/nedpals) - creator and maintainer
