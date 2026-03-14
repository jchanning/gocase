I need to deploy this to Oracle Cloud on an Ampere A1 (ARM64) instance.

Please write a `Dockerfile` for this Go application.
1.  **Build Stage:** Use `golang:1.22-alpine` as the builder.
2.  **Compiling:** When running `go build`, ensure you set `CGO_ENABLED=0` and `GOARCH=arm64` so it compiles a static binary for ARM.
3.  **Final Stage:** Use a lightweight image like `alpine:latest` or `scratch`.
4.  **Assets:** Ensure the `views/` directory and `assets/` directory are copied into the final image so the app can find the templates.
5.  **Expose:** Port 8080.