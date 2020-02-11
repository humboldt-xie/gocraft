package world

import (
	"flag"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/hashicorp/yamux"
)

var (
	serverAddr = flag.String("s", "", "server address")
	listenAddr = flag.String("l", "", "listen address")

	client *Client
)

type Session struct {
	ClientID   int32
	masterConn net.Conn
	*rpc.Client
}

type Server struct {
	*rpc.Server
	clientid int32
	sessions sync.Map
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()
	id := atomic.AddInt32(&s.clientid, 1)
	log.Printf("allocated %d for %s", id, conn.RemoteAddr())
	// send id to client, handshake done.

	ysess, err := yamux.Server(conn, nil)
	if err != nil {
		log.Print(err)
		return
	}

	clientConn, err := ysess.Open()
	if err != nil {
		log.Print(err)
		return
	}

	sess := &Session{
		ClientID:   id,
		masterConn: conn,
		Client:     rpc.NewClientWithCodec(jsonrpc.NewClientCodec(clientConn)),
	}
	defer sess.Client.Close()
	defer sess.masterConn.Close()

	divCall := sess.Go("Status.InitClient", &InitClientRequest{ClientID: id}, new(InitClientResponse), nil)
	replyCall := <-divCall.Done // will be equal to divCall

	if replyCall.Error != nil {
		log.Print(replyCall.Error)
		return
	}

	s.sessions.Store(id, sess)

	//s.playerCallback("online", id)

	//serveRpc(sess)
	sconn, err := ysess.Accept()
	if err != nil {
		log.Print(err)
		return
	}
	s.ServeCodec(jsonrpc.NewServerCodec(sconn))

	s.sessions.Delete(id)
	//s.playerCallback("offline", id)
	log.Printf("%s(%d) closed connection", conn.RemoteAddr(), id)
}

func (s *Server) Serve(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go s.handleConn(conn)
	}
}

type Client struct {
	*rpc.Client
	ClientID  int32
	rpcServer *rpc.Server
	waitInit  chan bool
}

func InitService() error {
	if *listenAddr == "" {
		return nil
	}
	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	server := &Server{
		Server: rpc.NewServer(),
	}
	server.RegisterName("Block", &BlockService{})
	server.RegisterName("Player", &PlayerService{})
	go server.Serve(l)
	return nil
}

func InitClient() error {
	if *serverAddr == "" {
		return nil
	}
	addr := *serverAddr
	if strings.Index(addr, ":") == -1 {
		addr += ":8421"
	}
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	client = &Client{
		rpcServer: rpc.NewServer(),
		waitInit:  make(chan bool, 1),
	}
	client.rpcServer.RegisterName("Block", &BlockService{})
	client.rpcServer.RegisterName("Player", &PlayerService{})
	client.rpcServer.RegisterName("Status", &StatusService{})

	sess, err := yamux.Client(conn, nil)
	if err != nil {
		log.Panic(err)
	}

	clientConn, err := sess.Open()
	if err != nil {
		log.Panic(err)
	}
	client.Client = rpc.NewClientWithCodec(jsonrpc.NewClientCodec(clientConn))

	clientService, err := sess.Accept()
	if err != nil {
		log.Panic(err)
	}

	go client.rpcServer.ServeCodec(jsonrpc.NewServerCodec(clientService))
	<-client.waitInit
	return nil
}

func ClientFetchChunk(id Vec3, f func(bid Vec3, w *Block)) {
	if client == nil {
		return
	}
	req := FetchChunkRequest{
		P:       id.X,
		Q:       id.Z,
		Version: store.GetChunkVersion(id),
	}
	rep := new(FetchChunkResponse)
	err := client.Call("Block.FetchChunk", req, rep)
	if err == rpc.ErrShutdown {
		return
	}
	if err != nil {
		log.Panic(err)
	}
	for _, b := range rep.Blocks {
		f(Vec3{b[0], b[1], b[2]}, NewBlock(b[3]))
	}
	if req.Version != rep.Version {
		store.UpdateChunkVersion(id, rep.Version)
	}
}

func ClientUpdateBlock(id Vec3, w *Block) {
	if client == nil {
		return
	}
	cid := id.Chunkid()
	req := &UpdateBlockRequest{
		Id:    client.ClientID,
		P:     cid.X,
		Q:     cid.Z,
		X:     id.X,
		Y:     id.Y,
		Z:     id.Z,
		Block: w,
	}
	rep := new(UpdateBlockResponse)
	err := client.Call("Block.UpdateBlock", req, rep)
	if err == rpc.ErrShutdown {
		return
	}
	if err != nil {
		log.Panic(err)
	}
	store.UpdateChunkVersion(id.Chunkid(), rep.Version)
}

func ClientUpdatePlayerState(state Position) {
	if client == nil {
		return
	}
	req := &UpdateStateRequest{
		Id: client.ClientID,
	}
	s := &req.State
	s.X, s.Y, s.Z, s.Rx, s.Ry = state.X(), state.Y(), state.Z(), state.Rx, state.Ry
	rep := new(UpdateStateResponse)
	err := client.Call("Player.UpdateState", req, rep)
	if err == rpc.ErrShutdown {
		return
	}
	if err != nil {
		log.Panic(err)
	}

	/*for id, player := range rep.Players {
		game.playerRender.UpdateOrAdd(id, player, false)
	}*/
}

type StatusService struct {
}
type InitClientRequest struct {
	ClientID int32
}
type InitClientResponse struct {
}

func (s *StatusService) InitClient(req *InitClientRequest, rep *InitClientResponse) error {
	log.Printf("init client %d\n", req.ClientID)
	client.ClientID = req.ClientID
	client.waitInit <- true
	return nil
}

type BlockService struct {
}

func (s *BlockService) FetchChunk(req *FetchChunkRequest, rep *FetchChunkResponse) error {
	id := Vec3{req.P, 0, req.Q}
	version := store.GetChunkVersion(id)
	rep.Version = version
	if req.Version == version {
		return nil
	}
	store.RangeBlocks(id, func(bid Vec3, w *Block) {
		rep.Blocks = append(rep.Blocks, [...]int{bid.X, bid.Y, bid.Z, w.Type})
	})
	return nil
}
func (s *BlockService) UpdateBlock(req *UpdateBlockRequest, rep *UpdateBlockResponse) error {
	log.Printf("rpc::UpdateBlock:%v", *req)
	//bid := Vec3{req.X, req.Y, req.Z}

	//game.world.UpdateBlock(bid, req.Block)
	//game.blockRender.DirtyBlock(bid)
	return nil
}

type PlayerService struct {
}

func (s *PlayerService) UpdateState(req *UpdateStateRequest, rep *UpdateStateResponse) error {
	/*game.playerRender.UpdateOrAdd(req.Id, req.State, false)
	rep.Players = make(map[int32]PlayerState)
	game.players.Range(func(k, v interface{}) bool {
		id := k.(int32)
		p := v.(*Player)
		if id == req.Id {
			return true
		}
		state := PlayerState{X: p.X(), Y: p.Y(), Z: p.Z(), Rx: p.Rx, Ry: p.Ry}
		rep.Players[id] = state
		return true
	})*/

	return nil
}

func (s *PlayerService) RemovePlayer(req *RemovePlayerRequest, rep *RemovePlayerResponse) error {
	//game.playerRender.Remove(req.Id)
	return nil
}
