FROM golang:1.20.11 as builder

RUN mkdir -p /app
WORKDIR /app

COPY . /app/

RUN make compile

# FROM scratch

# COPY --from=builder /app/main /app/main
# COPY --from=builder /app/go.mod /app/go.mod
# COPY --from=builder /app/go.sum /app/go.sum

ENTRYPOINT ["/app/main"]