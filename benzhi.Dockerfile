FROM golang:1.22-bookworm

WORKDIR /workspace

COPY backend/go.mod backend/go.sum ./backend/
RUN cd backend && go mod download

COPY backend ./backend

WORKDIR /workspace/backend
RUN go build ./...

CMD [go, test, ./...]
