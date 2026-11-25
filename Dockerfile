# Etap 1: kompilacja statycznego pliku binarnego
FROM golang:1.24 AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/foxbat ./cmd/foxbat

# Etap 2: minimalny obraz uruchomieniowy bez powłoki, uruchamiany jako użytkownik bez uprawnień
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/foxbat /app/foxbat
COPY web/templates /app/web/templates
ENV SERVER_PORT=":9080"
EXPOSE 9080
USER nonroot:nonroot
ENTRYPOINT ["/app/foxbat"]
