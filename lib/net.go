package lib

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/dunelang/dune"
)

func init() {
	dune.RegisterLib(Net, `

declare namespace net {
	export function inCIDR(cidr: string, ip: string): boolean;

    export function getIPAddress(): string

    export function getMacAddress(): string

    export type dialNetwork = "tcp" | "tcp4" | "tcp6" | "udp" | "udp4" | "udp6" | "ip" | "ip4" | "ip6" | "unix" | "unixgram" | "unixpacket"

	export type listenNetwork = "tcp" | "tcp4" | "tcp6" | "unix" | "unixpacket"
	
	export interface IP {
		string(): string
	}

    export interface Connection {
        read(b: byte[]): number
        write(b: byte[]): number
        setDeadline(t: time.Time): void
        setWriteDeadline(t: time.Time): void
        setReadDeadline(t: time.Time): void
        close(): void
    }

    export interface Listener {
        accept(): Connection
        close(): void
    }

    export function dial(network: dialNetwork, address: string): Connection
    export function dialTimeout(network: dialNetwork, address: string, d: time.Duration | number): Connection
    export function listen(network: listenNetwork, address: string): Listener

    export interface TCPListener {
        accept(): TCPConnection
        close(): void
	}
	
    export function dialTCP(network: dialNetwork, localAddr: TCPAddr, remoteAddr: TCPAddr): TCPConnection
	export function listenTCP(network: listenNetwork, address: TCPAddr): TCPListener

    export interface TCPConnection {
		localAddr: TCPAddr | Addr
		remoteAddr: TCPAddr | Addr
        read(b: byte[]): number
        write(b: byte[]): number
        setDeadline(t: time.Time): void
        setWriteDeadline(t: time.Time): void
        setReadDeadline(t: time.Time): void
        close(): void
	}
	
	export function resolveTCPAddr(network: dialNetwork, address: string): TCPAddr
	
    export interface TCPAddr {
		IP: IP
		port: number
		IPAddress(): string
        string(): string
    }

    export interface Addr {
		IPAddress(): string
        string(): string
    }
}

`)
}

var Net = []dune.NativeFunction{
	{
		Name:      "net.inCIDR",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			cidr := args[0].String()

			ip := net.ParseIP(args[1].String())

			_, ipnet, err := net.ParseCIDR(cidr)
			if err != nil {
				return dune.NullValue, err
			}

			v := ipnet.Contains(ip)

			return dune.NewBool(v), nil
		},
	},
	{
		Name:        "net.listen",
		Arguments:   2,
		Permissions: []string{"netListen"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			listener, err := newNetListener(args[0].String(), args[1].String(), vm)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(listener), nil
		},
	},
	{
		Name:        "net.listenTCP",
		Arguments:   2,
		Permissions: []string{"netListen"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.Object); err != nil {
				return dune.NullValue, err
			}

			addr, ok := args[1].ToObject().(*tcpAddr)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected param 2 to be TCPAddr, got %s", args[1].TypeName())
			}

			listener, err := newTCPListener(args[0].String(), addr.addr, vm)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(listener), nil
		},
	},
	{
		Name:      "net.resolveTCPAddr",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}

			addr, err := net.ResolveTCPAddr(args[0].String(), args[1].String())
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(&tcpAddr{addr: addr}), nil
		},
	},
	{
		Name:      "net.dialTCP",
		Arguments: 3,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateOrNilArgs(args, dune.String, dune.Object, dune.Object); err != nil {
				return dune.NullValue, err
			}

			network := args[0].String()

			var localAddr *net.TCPAddr

			lArg := args[1].ToObjectOrNil()
			if lArg != nil {
				ltcpAddr, ok := lArg.(*tcpAddr)
				if !ok {
					return dune.NullValue, fmt.Errorf("expected param 2 to be TCPAddr, got %s", args[1].TypeName())
				}
				localAddr = ltcpAddr.addr
			}

			remoteAddr, ok := args[2].ToObject().(*tcpAddr)
			if !ok {
				return dune.NullValue, fmt.Errorf("expected param 3 to be TCPAddr, got %s", args[1].TypeName())
			}

			conn, err := net.DialTCP(network, localAddr, remoteAddr.addr)
			if err != nil {
				return dune.NullValue, err
			}

			tc := newTCPConn(conn, vm)

			return dune.NewObject(tc), nil
		},
	},
	{
		Name:      "net.dial",
		Arguments: 2,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if err := ValidateArgs(args, dune.String, dune.String); err != nil {
				return dune.NullValue, err
			}
			conn, err := net.Dial(args[0].String(), args[1].String())
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(newNetConn(conn, vm)), nil
		},
	},
	{
		Name:      "net.dialTimeout",
		Arguments: 3,
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			if args[0].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected param 1 to be string, got %s", args[0].TypeName())
			}
			if args[1].Type != dune.String {
				return dune.NullValue, fmt.Errorf("expected param 2 to be string, got %s", args[1].TypeName())
			}

			d, err := ToDuration(args[2])
			if err != nil {
				return dune.NullValue, err
			}

			conn, err := net.DialTimeout(args[0].String(), args[1].String(), d)
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewObject(newNetConn(conn, vm)), nil
		},
	},
	{
		Name:        "net.getIPAddress",
		Arguments:   0,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			addrs, err := net.InterfaceAddrs()
			if err != nil {
				return dune.NullValue, err
			}

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						return dune.NewString(ipnet.IP.String()), nil
					}
				}
			}

			return dune.NullValue, fmt.Errorf("no IP address found")
		},
	},
	{
		Name:        "net.getMacAddress",
		Arguments:   0,
		Permissions: []string{"trusted"},
		Function: func(this dune.Value, args []dune.Value, vm *dune.VM) (dune.Value, error) {
			addrs, err := net.InterfaceAddrs()
			if err != nil {
				return dune.NullValue, err
			}

			var ip string

			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip = ipnet.IP.String()
						break
					}
				}
			}

			if ip == "" {
				return dune.NullValue, fmt.Errorf("no IP address found")
			}

			interfaces, err := net.Interfaces()
			if err != nil {
				return dune.NullValue, err
			}

			var hardwareName string

			for _, interf := range interfaces {
				if addrs, err := interf.Addrs(); err == nil {
					for _, addr := range addrs {
						// only interested in the name with current IP address
						if strings.Contains(addr.String(), ip) {
							hardwareName = interf.Name
							break
						}
					}
				}
			}

			if hardwareName == "" {
				return dune.NullValue, fmt.Errorf("no network hardware found")
			}

			netInterface, err := net.InterfaceByName(hardwareName)
			if err != nil {
				return dune.NullValue, err
			}

			macAddress := netInterface.HardwareAddr

			// verify if the MAC address can be parsed properly
			hwAddr, err := net.ParseMAC(macAddress.String())
			if err != nil {
				return dune.NullValue, err
			}

			return dune.NewString(hwAddr.String()), nil
		},
	},
}

func newNetConn(conn net.Conn, vm *dune.VM) netConn {
	f := netConn{conn: conn}
	vm.SetGlobalFinalizer(f)
	return f
}

type netConn struct {
	conn net.Conn
}

func (netConn) Type() string {
	return "net.Connection"
}

func (c netConn) Close() error {
	return c.conn.Close()
}

func (c netConn) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "read":
		return c.read
	case "write":
		return c.write
	case "setDeadline":
		return c.setDeadline
	case "setWriteDeadline":
		return c.setWriteDeadline
	case "setReadDeadline":
		return c.setReadDeadline
	case "close":
		return c.close
	}
	return nil
}

func (c netConn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c netConn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c netConn) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}
	b := args[0].ToBytes()
	n, err := c.conn.Read(b)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewInt(n), nil
}

func (c netConn) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var b []byte
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	a := args[0]
	switch a.Type {
	case dune.Array, dune.Bytes:
		b = a.ToBytes()
	case dune.Int:
		b = []byte{byte(a.ToInt())}
	default:
		return dune.NullValue, ErrInvalidType
	}

	n, err := c.conn.Write(b)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewInt(n), nil
}

func (c netConn) setDeadline(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t, ok := args[0].ToObjectOrNil().(TimeObj)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	c.conn.SetDeadline(time.Time(t))
	return dune.NullValue, nil
}

func (c netConn) setWriteDeadline(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t, ok := args[0].ToObjectOrNil().(TimeObj)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	c.conn.SetWriteDeadline(time.Time(t))
	return dune.NullValue, nil
}

func (c netConn) setReadDeadline(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t, ok := args[0].ToObjectOrNil().(TimeObj)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	c.conn.SetReadDeadline(time.Time(t))
	return dune.NullValue, nil
}

func (c netConn) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	c.conn.Close()
	return dune.NullValue, nil
}

type IP struct {
	ip net.IP
}

func (*IP) Type() string {
	return "net.IP"
}

func (ip *IP) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "string":
		return ip.string
	}
	return nil
}

func (ip *IP) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return dune.NewString(ip.ip.String()), nil
}

func (ip *IP) String() string {
	return ip.ip.String()
}

type Addr struct {
	addr net.Addr
}

func (*Addr) Type() string {
	return "net.Addr"
}

func (a *Addr) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "IPAddress":
		return a.ipAddress
	case "string":
		return a.string
	}
	return nil
}

func (a *Addr) ipAddress(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	s := a.addr.String()
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return dune.NullValue, fmt.Errorf("could not parse IP: %s", s)
	}
	return dune.NewString(parts[0]), nil
}

func (a *Addr) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return dune.NewString(a.addr.String()), nil
}

func (a *Addr) String() string {
	return a.addr.String()
}

type tcpAddr struct {
	addr *net.TCPAddr
}

func (*tcpAddr) Type() string {
	return "net.TCPAddr"
}

func (a *tcpAddr) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "IP":
		ip := &IP{ip: a.addr.IP}
		return dune.NewObject(ip), nil

	case "port":
		return dune.NewInt(a.addr.Port), nil
	}

	return dune.UndefinedValue, nil
}

func (a *tcpAddr) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "IPAddress":
		return a.ipAddress
	case "string":
		return a.string
	}
	return nil
}

func (a *tcpAddr) ipAddress(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	s := a.addr.String()
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return dune.NullValue, fmt.Errorf("could not parse IP: %s", s)
	}
	return dune.NewString(parts[0]), nil
}

func (a *tcpAddr) string(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	return dune.NewString(a.addr.String()), nil
}

func (a *tcpAddr) String() string {
	return a.addr.String()
}

func newTCPConn(conn *net.TCPConn, vm *dune.VM) *tcpConn {
	f := &tcpConn{conn: conn}
	vm.SetGlobalFinalizer(f)
	return f
}

type tcpConn struct {
	conn *net.TCPConn
}

func (*tcpConn) Type() string {
	return "net.TCPConnection"
}

func (c *tcpConn) Close() error {
	return c.conn.Close()
}

func (c *tcpConn) GetField(name string, vm *dune.VM) (dune.Value, error) {
	switch name {
	case "localAddr":
		a := c.conn.LocalAddr()
		ta, ok := a.(*net.TCPAddr)
		if ok {
			return dune.NewObject(&tcpAddr{addr: ta}), nil
		}
		return dune.NewObject(&Addr{addr: a}), nil

	case "remoteAddr":
		a := c.conn.RemoteAddr()
		ta, ok := a.(*net.TCPAddr)
		if ok {
			return dune.NewObject(&tcpAddr{addr: ta}), nil
		}
		return dune.NewObject(&Addr{addr: a}), nil
	}

	return dune.UndefinedValue, nil
}

func (c *tcpConn) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "read":
		return c.read
	case "write":
		return c.write
	case "setDeadline":
		return c.setDeadline
	case "setWriteDeadline":
		return c.setWriteDeadline
	case "setReadDeadline":
		return c.setReadDeadline
	case "close":
		return c.close
	}
	return nil
}

func (c *tcpConn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *tcpConn) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *tcpConn) read(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Bytes); err != nil {
		return dune.NullValue, err
	}
	b := args[0].ToBytes()
	n, err := c.conn.Read(b)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewInt(n), nil
}

func (c *tcpConn) write(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	var b []byte
	if len(args) != 1 {
		return dune.NullValue, fmt.Errorf("expected 1 argument, got %d", len(args))
	}
	a := args[0]
	switch a.Type {
	case dune.Array, dune.Bytes:
		b = a.ToBytes()
	case dune.Int:
		b = []byte{byte(a.ToInt())}
	default:
		return dune.NullValue, ErrInvalidType
	}

	n, err := c.conn.Write(b)
	if err != nil {
		return dune.NullValue, err
	}
	return dune.NewInt(n), nil
}

func (c *tcpConn) setDeadline(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t, ok := args[0].ToObjectOrNil().(TimeObj)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	c.conn.SetDeadline(time.Time(t))
	return dune.NullValue, nil
}

func (c *tcpConn) setWriteDeadline(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t, ok := args[0].ToObjectOrNil().(TimeObj)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	c.conn.SetWriteDeadline(time.Time(t))
	return dune.NullValue, nil
}

func (c *tcpConn) setReadDeadline(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args, dune.Object); err != nil {
		return dune.NullValue, err
	}

	t, ok := args[0].ToObjectOrNil().(TimeObj)
	if !ok {
		return dune.NullValue, ErrInvalidType
	}

	c.conn.SetReadDeadline(time.Time(t))
	return dune.NullValue, nil
}

func (c *tcpConn) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	c.conn.Close()
	return dune.NullValue, nil
}

func newNetListener(network, port string, vm *dune.VM) (*netListener, error) {
	ls, err := net.Listen(network, port)
	if err != nil {
		return nil, err
	}
	listener := &netListener{ls: ls}
	vm.SetGlobalFinalizer(listener)
	return listener, nil
}

type netListener struct {
	ls net.Listener
}

func (netListener) Type() string {
	return "net.Listener"
}

func (c *netListener) Close() error {
	return c.ls.Close()
}

func (c *netListener) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "accept":
		return c.accept
	case "close":
		return c.close
	}
	return nil
}

func (c *netListener) accept(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	conn, err := c.ls.Accept()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(newNetConn(conn, vm)), nil
}

func (c *netListener) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := c.ls.Close()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}

func newTCPListener(network string, localAddr *net.TCPAddr, vm *dune.VM) (*tcpListener, error) {
	ls, err := net.ListenTCP(network, localAddr)
	if err != nil {
		return nil, err
	}
	listener := &tcpListener{ls: ls}
	vm.SetGlobalFinalizer(listener)
	return listener, nil
}

type tcpListener struct {
	ls *net.TCPListener
}

func (tcpListener) Type() string {
	return "net.TCPListener"
}

func (c *tcpListener) Close() error {
	return c.ls.Close()
}

func (c *tcpListener) GetMethod(name string) dune.NativeMethod {
	switch name {
	case "accept":
		return c.accept
	case "close":
		return c.close
	}
	return nil
}

func (c *tcpListener) accept(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	conn, err := c.ls.AcceptTCP()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NewObject(newTCPConn(conn, vm)), nil
}

func (c *tcpListener) close(args []dune.Value, vm *dune.VM) (dune.Value, error) {
	if err := ValidateArgs(args); err != nil {
		return dune.NullValue, err
	}

	err := c.ls.Close()
	if err != nil {
		return dune.NullValue, err
	}

	return dune.NullValue, nil
}
