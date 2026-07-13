# --- build stage ---------------------------------------------------------
# Full Go toolchain, discarded after this stage — none of it ships in the
# final image.
FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/api /api

EXPOSE 8080

ENTRYPOINT ["/api"]