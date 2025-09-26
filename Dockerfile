# Multi-stage build for AIS demo

FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
# Copy only source code needed to build; exclude model caches
COPY internal/ais ./internal/ais
COPY cmd ./cmd
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o /out/aisdemo ./cmd/aisdemo

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/aisdemo /app/aisdemo
# COPY internal/ollama-models /root/.ollama
EXPOSE 8890
ENV OLLAMA_URL=http://ollama:11434
ENV OLLAMA_MODEL=llama3
ENV AIS_SECRET=dev-secret-change-me
ENTRYPOINT ["/app/aisdemo"]


