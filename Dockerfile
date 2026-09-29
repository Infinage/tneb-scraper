FROM --platform=$BUILDPLATFORM golang:alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /tneb-scraper .

FROM alpine:latest as runtime
WORKDIR /app

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    font-liberation font-noto

COPY --from=go-builder /tneb-scraper /app/tneb-scraper
COPY bills.html /app/bills.html

EXPOSE 8080
CMD ["/app/tneb-scraper"]
