FROM golang:1.27.0-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG SERVICE
RUN CGO_ENABLED=0 go build -trimpath -o /service ./cmd/${SERVICE}

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=build /service /service
USER 65532:65532
ENTRYPOINT ["/service"]
