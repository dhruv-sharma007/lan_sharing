FROM golang:alpine

WORKDIR /app

# Install dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary
RUN go build -o lanshare ./cmd/lanshare

# Run the application
CMD ["./lanshare"]