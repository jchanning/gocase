Now let's build the HTTP server and the first view.

1.  Create a `internal/server/server.go` file.
2.  Define a `Server` struct that holds the `database.Service` we created earlier and a `*chi.Mux` router.
3.  Create a `NewServer(db *database.Service) *Server` function.
4.  Inside `NewServer`, set up the Chi router.
    - Add standard middleware: `middleware.Logger`, `middleware.Recoverer`.
    - Serve static files (css/js) from a `./assets` directory under the `/assets` route.
5.  Create a `views/layout.html` template. This should be the base HTML skeleton (<html>, <head>, <body>). Include the HTMX script tag from unpkg and TailwindCSS via CDN in the <head>.
6.  Create a `views/home.html` template that defines a "content" block to be rendered inside the layout.
7.  Add a route `GET /` that parses these templates and serves the home page.
8.  Update `cmd/server/main.go` to start this server on port 8080.