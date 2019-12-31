FROM golang:alpine AS base

WORKDIR /app
COPY . .
RUN go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build

FROM alpine as runner

ENV ENV=dev
WORKDIR /app
COPY --from=base /app/s3-backup-scheduler /app/s3-backup-scheduler
CMD ./s3-backup-scheduler --config_file=config/${ENV}.yaml