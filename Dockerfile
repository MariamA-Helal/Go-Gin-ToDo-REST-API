# ==========================================
# Stage 1:Builder
# ==========================================

FROM golang:alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o todo-app ./cmd/api

# ==========================================
# Stage 2: مرحلة التشغيل (Runner)
# ==========================================

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/todo-app .

EXPOSE 8080

CMD ["./todo-app"]