FROM --platform=$BUILDPLATFORM golang:alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /tneb-scraper .

FROM alpine:latest AS runtime
WORKDIR /app

RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-eng tzdata
ENV TZ=Asia/Kolkata

COPY --from=go-builder /tneb-scraper /app/tneb-scraper

EXPOSE 8080
CMD ["/app/tneb-scraper"]
