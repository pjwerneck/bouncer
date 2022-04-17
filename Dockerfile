FROM golang:1.18 as builder
COPY . /go/src/github.com/pjwerneck/bouncer
WORKDIR /go/src/github.com/pjwerneck/bouncer
RUN go mod download
RUN CGO_ENABLED=0 go build -o /main


FROM scratch AS runner
WORKDIR /
COPY --from=builder /main /main
ENTRYPOINT ["/main"]
