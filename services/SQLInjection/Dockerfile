FROM golang:1.14.3-alpine AS build
WORKDIR /src
RUN apk --no-cache add build-base git bzr mercurial gcc
RUN go get github.com/mattn/go-sqlite3
COPY . .
RUN go build -o /bin/gosql .
FROM alpine
WORKDIR /app
COPY --from=build /bin/gosql /app/
CMD ["./gosql"]