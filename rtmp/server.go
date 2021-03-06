package rtmp

import (
	"encoding/binary"
	"io"
	"log"
	"math/rand"
	"net"
	"time"
)

type Listener struct {
	l *net.TCPListener
}

func Listen(laddr string) (l *Listener, err error) {
	a, err := net.ResolveTCPAddr("tcp", laddr)
	if err != nil {
		return
	}
	tl, err := net.ListenTCP("tcp", a)
	if err != nil {
		return
	}
	l = &Listener{l: tl}
	return
}

func (l *Listener) Accept() (c *ServerConn, err error) {
	tc, err := l.l.AcceptTCP()
	if err != nil {
		return
	}
	c = newServerConn(tc)
	return
}

func (l *Listener) Close() error {
	return l.l.Close()
}

type ServerConn struct {
	conn
}

func newServerConn(tcpConn *net.TCPConn) *ServerConn {
	return &ServerConn{conn{
		c:            tcpConn,
		chunkStreams: make(map[uint32]*ChunkStream),
		inChunkSize:  DEFAULT_CHUNK_SIZE,
		outChunkSize: DEFAULT_CHUNK_SIZE}}
}

// server side handshake
func (c *ServerConn) Handshake() (err error) {
	b := make([]byte, HANKSHAKE_MESSAGE_LEN)

	_, err = c.c.Read(b[:1])
	if err != nil {
		return
	}
	log.Printf("Get C0=%d\n", b[0])

	_, err = c.c.Write(b[:1])
	if err != nil {
		return
	}
	log.Println("S0 sent")

	c.ts = time.Now()
	binary.BigEndian.PutUint32(b, 0)
	// a little random to the data
	binary.BigEndian.PutUint32(b[8+rand.Intn(HANKSHAKE_MESSAGE_LEN-8):], rand.Uint32())
	_, err = c.c.Write(b)
	if err != nil {
		return
	}
	log.Println("S1 sent")

	_, err = io.ReadFull(c.c, b)
	if err != nil {
		return
	}
	ct1 := binary.BigEndian.Uint32(b)
	log.Printf("Get C1 time=%d\n", ct1)
	binary.BigEndian.PutUint32(b[4:], c.Timestamp())
	// a little random to the data
	binary.BigEndian.PutUint32(b[8+rand.Intn(HANKSHAKE_MESSAGE_LEN-8):], rand.Uint32())
	_, err = c.c.Write(b)
	if err != nil {
		return
	}
	log.Println("S2 sent")

	_, err = io.ReadFull(c.c, b)
	if err != nil {
		return
	}
	log.Println("Get C2. handshake completed.")
	return
}
