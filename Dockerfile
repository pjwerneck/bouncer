FROM golang:1.19 as builder
# copy the source code
COPY . /go/src/github.com/pjwerneck/bouncer
WORKDIR /go/src/github.com/pjwerneck/bouncer
# download dependencies
RUN go mod download
# build the binary
RUN CGO_ENABLED=0 go build -o /main


FROM scratch AS runner
# copy the binary
WORKDIR /
COPY --from=builder /main /main
# run the binary
ENTRYPOINT ["/main"]
