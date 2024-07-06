# Start with the official Golang image
FROM golang:1.22.1 as build

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application and ensure it's executable
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gobank

# Use a minimal base image to run the application
# FROM gcr.io/distroless/base-debian10

# Copy the compiled binary to the minimal base image
# COPY --from=build /app/api /app/api

# # Set the entrypoint
# ENTRYPOINT ["app/api"]

# Set the port the application listens on
ENV PORT 5252

# Expose the port
EXPOSE 5252

CMD ["/docker-gobank"]