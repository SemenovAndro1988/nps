package proxy

import (
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"sync"

	"ehang.io/nps/bridge"
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/conn"
	"ehang.io/nps/lib/file"
	"github.com/astaxie/beego/logs"
)

// SharedSocks5Server exposes a single SOCKS5 endpoint that routes traffic
// through different bots based on the SOCKS5 username/password presented
// by the connecting application.
//
// Every connected/registered client is treated as a "bot". A user obtains
// its credentials from the admin panel and connects to this single port
// like any normal SOCKS5 proxy:
//
//	socks5://<bot user>:<bot pass>@<server ip>:<socks5_port>
//
// The server then forwards the request to the matching bot.
type SharedSocks5Server struct {
	BaseServer
	listener net.Listener
	bridge   *bridge.Bridge
	port     int
	ip       string
}

// NewSharedSocks5Server creates a new shared SOCKS5 server on the given
// ip and port. Pass an empty ip to bind on 0.0.0.0.
func NewSharedSocks5Server(b *bridge.Bridge, ip string, port int) *SharedSocks5Server {
	s := &SharedSocks5Server{
		bridge: b,
		ip:     ip,
		port:   port,
	}
	s.BaseServer = BaseServer{bridge: b, Mutex: sync.Mutex{}}
	return s
}

func (s *SharedSocks5Server) Start() error {
	addr := s.ip + ":" + strconv.Itoa(s.port)
	if s.ip == "" {
		addr = "0.0.0.0:" + strconv.Itoa(s.port)
	}
	logs.Info("shared socks5 server started on %s", addr)
	return conn.NewTcpListenerAndProcess(addr, func(c net.Conn) {
		s.handleConn(c)
	}, &s.listener)
}

func (s *SharedSocks5Server) Close() error {
	if s.listener == nil {
		return nil
	}
	return s.listener.Close()
}

func (s *SharedSocks5Server) handleConn(c net.Conn) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(c, buf); err != nil {
		c.Close()
		return
	}
	if buf[0] != 5 {
		logs.Warn("shared socks5: only socks5 is supported, request from %s", c.RemoteAddr())
		c.Close()
		return
	}
	nMethods := buf[1]
	methods := make([]byte, nMethods)
	if n, err := io.ReadFull(c, methods); n != int(nMethods) || err != nil {
		c.Close()
		return
	}
	// We always require username / password authentication so we can
	// figure out which bot to route the request through.
	if _, err := c.Write([]byte{5, UserPassAuth}); err != nil {
		c.Close()
		return
	}
	client, err := s.authClient(c)
	if err != nil {
		logs.Warn("shared socks5 auth failed from %s: %s", c.RemoteAddr(), err.Error())
		c.Close()
		return
	}

	// Route the SOCKS5 request through the matched client.
	if err := s.proxyForClient(c, client); err != nil {
		logs.Warn("shared socks5 proxy error for client %d: %s", client.Id, err.Error())
		c.Close()
		return
	}
}

// authClient performs SOCKS5 USER/PASS authentication and returns the
// corresponding client.
func (s *SharedSocks5Server) authClient(c net.Conn) (*file.Client, error) {
	header := []byte{0, 0}
	if _, err := io.ReadAtLeast(c, header, 2); err != nil {
		return nil, err
	}
	if header[0] != userAuthVersion {
		return nil, errors.New("authentication method is not supported")
	}
	userLen := int(header[1])
	user := make([]byte, userLen)
	if _, err := io.ReadAtLeast(c, user, userLen); err != nil {
		return nil, err
	}
	if _, err := c.Read(header[:1]); err != nil {
		return nil, errors.New("failed to read password length")
	}
	passLen := int(header[0])
	pass := make([]byte, passLen)
	if _, err := io.ReadAtLeast(c, pass, passLen); err != nil {
		return nil, err
	}
	username := string(user)
	password := string(pass)

	client := lookupClientByCredentials(username, password)
	if client == nil {
		_, _ = c.Write([]byte{userAuthVersion, authFailure})
		return nil, errors.New("no bot matches the supplied credentials")
	}
	if !client.Status {
		_, _ = c.Write([]byte{userAuthVersion, authFailure})
		return nil, errors.New("bot is disabled")
	}
	if _, ok := s.bridge.Client.Load(client.Id); !ok {
		_, _ = c.Write([]byte{userAuthVersion, authFailure})
		return nil, errors.New("bot is offline")
	}
	if _, err := c.Write([]byte{userAuthVersion, authSuccess}); err != nil {
		return nil, err
	}
	return client, nil
}

// proxyForClient implements the SOCKS5 connect / udp associate phase by
// delegating the connection to the bot identified by client.
func (s *SharedSocks5Server) proxyForClient(c net.Conn, client *file.Client) error {
	header := make([]byte, 3)
	if _, err := io.ReadFull(c, header); err != nil {
		return err
	}
	if header[0] != 5 {
		return errors.New("expected socks5 in request phase")
	}
	cmd := header[1]
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(c, addrType); err != nil {
		return err
	}
	var host string
	switch addrType[0] {
	case ipV4:
		ipv4 := make(net.IP, net.IPv4len)
		if _, err := io.ReadFull(c, ipv4); err != nil {
			return err
		}
		host = ipv4.String()
	case ipV6:
		ipv6 := make(net.IP, net.IPv6len)
		if _, err := io.ReadFull(c, ipv6); err != nil {
			return err
		}
		host = ipv6.String()
	case domainName:
		var dl uint8
		if err := binary.Read(c, binary.BigEndian, &dl); err != nil {
			return err
		}
		dom := make([]byte, dl)
		if _, err := io.ReadFull(c, dom); err != nil {
			return err
		}
		host = string(dom)
	default:
		s.sendReply(c, addrTypeNotSupported)
		return errors.New("unsupported address type")
	}
	var port uint16
	if err := binary.Read(c, binary.BigEndian, &port); err != nil {
		return err
	}
	if cmd != connectMethod {
		// We only support TCP CONNECT in the shared SOCKS5 entry. UDP
		// associate, BIND and other commands are not exposed.
		s.sendReply(c, commandNotSupported)
		return errors.New("only socks5 connect is supported")
	}

	addr := net.JoinHostPort(host, strconv.Itoa(int(port)))

	cnf := client.Cnf
	if cnf == nil {
		cnf = &file.Config{}
	}

	// CopyWaitGroup needs a non-nil Flow pointer for accounting; we
	// use a throw-away one so the periodic dealClientData rebuild
	// (which recomputes client.Flow from tunnels/hosts) does not
	// fight with us. Per-bot SOCKS5 byte counters are not surfaced
	// in the panel today; rate-limiting still works because we pass
	// the real client.Rate.
	flow := new(file.Flow)
	link := conn.NewLink(common.CONN_TCP, addr, cnf.Crypt, cnf.Compress, c.RemoteAddr().String(), false)
	target, err := s.bridge.SendLinkInfo(client.Id, link, nil)
	if err != nil {
		s.sendReply(c, hostUnreachable)
		return err
	}
	s.sendReply(c, succeeded)
	conn.CopyWaitGroup(target, c, link.Crypt, link.Compress, client.Rate, flow, true, nil)
	return nil
}

// sendReply mirrors the helper used by the per-tunnel SOCKS5 server.
func (s *SharedSocks5Server) sendReply(c net.Conn, rep uint8) {
	reply := []byte{5, rep, 0, 1}
	localAddr := c.LocalAddr().String()
	localHost, localPort, _ := net.SplitHostPort(localAddr)
	ipBytes := net.ParseIP(localHost).To4()
	if ipBytes == nil {
		ipBytes = []byte{0, 0, 0, 0}
	}
	nPort, _ := strconv.Atoi(localPort)
	reply = append(reply, ipBytes...)
	portBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(portBytes, uint16(nPort))
	reply = append(reply, portBytes...)
	_, _ = c.Write(reply)
}

// lookupClientByCredentials returns the client whose SOCKS5 user/password
// match the supplied credentials, or nil.
func lookupClientByCredentials(username, password string) *file.Client {
	if username == "" || password == "" {
		return nil
	}
	var found *file.Client
	file.GetDb().JsonDb.Clients.Range(func(key, value interface{}) bool {
		c := value.(*file.Client)
		if c.Cnf == nil {
			return true
		}
		if c.Cnf.U == username && c.Cnf.P == password {
			found = c
			return false
		}
		return true
	})
	return found
}
