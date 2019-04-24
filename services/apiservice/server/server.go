package server

import (
	"google.golang.org/grpc"

	albumService "github.com/farerpath/albumservice/proto"
	authService "github.com/farerpath/authservice/proto"
	sessionService "github.com/farerpath/sessionservice/proto"
)

var authClient authService.AuthClient
var albumClient albumService.AlbumClient
var sessionClient sessionService.SessionClient

func InitGrpcConn(conn1, conn2, conn3 *grpc.ClientConn) {
	authClient = authService.NewAuthClient(conn1)
	albumClient = albumService.NewAlbumClient(conn2)
	sessionClient = sessionService.NewSessionClient(conn3)
}
