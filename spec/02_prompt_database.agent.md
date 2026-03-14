Based on the project context, please write the code for the database setup.

1.  Create a file `internal/database/database.go`.
2.  Create a `Service` struct that holds the `*pgxpool.Pool`.
3.  Write a function `NewService(connString string) (*Service, error)` that initializes the connection pool.
4.  Ensure the connection is pinged to verify it works before returning.
5.  Include a `Close()` method to close the pool.
6.  Please also provide a snippet for `cmd/server/main.go` that loads the connection string from an environment variable named `DATABASE_URL`, initializes this service, and ensures it closes when the app exits.