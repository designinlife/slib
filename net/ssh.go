package net

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"

	"github.com/designinlife/slib/errors"
	"github.com/designinlife/slib/glog"
	"github.com/designinlife/slib/types"
)

type SSHTunnelEndpoint struct {
	Host string
	Port int
}

func (endpoint *SSHTunnelEndpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHTunnel struct {
	// 本地
	Local *SSHTunnelEndpoint
	// 隧道主机
	Server *SSHTunnelEndpoint
	// 目标主机
	Remote *SSHTunnelEndpoint
	// SSH 客户端配置
	Config *ssh.ClientConfig
	exit   chan bool
}

func (tunnel *SSHTunnel) Start(opened chan bool) error {
	listener, err := net.Listen("tcp", tunnel.Local.String())
	if err != nil {
		return errors.Wrapf(err, "SSHTunnel Start net Listen %s failed", tunnel.Local.String())
	}
	defer listener.Close()

	// 若 Local 端口为 0, 则重新读取端口号 ...
	if tunnel.Local.Port == 0 {
		addr := listener.Addr().(*net.TCPAddr)
		tunnel.Local.Port = addr.Port
	}

	glog.Debugf("[SSHTunnel] Listen: %s", tunnel.Local.String())

	opened <- true

	for {
		conn, err1 := listener.Accept()
		if err1 != nil {
			return errors.Wrap(err, "SSHTunnel Start net Accept failed")
		}
		go tunnel.forward(conn)

		select {
		case <-tunnel.exit:
			break
		}
	}
}

func (tunnel *SSHTunnel) Stop() error {
	close(tunnel.exit)

	return nil
}

func (tunnel *SSHTunnel) forward(localConn net.Conn) {
	serverConn, err := ssh.Dial("tcp", tunnel.Server.String(), tunnel.Config)
	if err != nil {
		// logger.Errorf("[SSHTunnel] Server dial error: %s", err)
		_, _ = fmt.Fprintf(os.Stderr, "[SSHTunnel] Server dial error: %s\n", err)
		return
	}

	remoteConn, err := serverConn.Dial("tcp", tunnel.Remote.String())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[SSHTunnel] Remote dial error: %s\n", err)
		return
	}

	copyConn := func(writer, reader net.Conn) {
		_, err1 := io.Copy(writer, reader)
		if err1 != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[SSHTunnel] io.Copy error: %s", err1)
		}
	}

	go copyConn(localConn, remoteConn)
	go copyConn(remoteConn, localConn)
}

// SSHClient SSH 客户端
type SSHClient struct {
	// RSA 私钥证书路径或全文内容
	PrivateKey string
	// 主机 Domain/IP 地址
	Host string
	// SSH 端口
	Port int
	// SSH 登录用户名
	User string
	// SSH 登录密码
	Password string
	// 静默方式: 不输出 Stdout 信息
	Quiet bool
	// 是否已连接？
	Connected bool
	// SSH 客户端实例
	Client *ssh.Client
	// 代理服务器地址 (支持 http,https,socks5,socks5h, 例如: http://127.0.0.1:3128, socks5://127.0.0.1:1080)
	Proxy string
	// 超时时间 (默认不超时)
	Timeout time.Duration
	// 开启 TTY 终端模式
	TTY bool
	// 缓冲区大小 (默认: 8192 bytes)
	ChunkSize uint16
	// SSH 隧道
	Tunnel *SSHTunnel
	// Logger 接口对象
	Logger types.Logger
}

type SSHClientOption func(*SSHClient)

// SSHOptionWithProxy 支持 http/https/socks5/socks5h 代理协议。
func SSHOptionWithProxy(proxyUrl string) SSHClientOption {
	return func(c *SSHClient) {
		c.Proxy = proxyUrl
	}
}

func SSHOptionWithTimeout(timeout time.Duration) SSHClientOption {
	return func(c *SSHClient) {
		c.Timeout = timeout
	}
}

func SSHOptionWithChunkSize(chunkSize uint16) SSHClientOption {
	return func(c *SSHClient) {
		c.ChunkSize = chunkSize
	}
}

func SSHOptionWithTunnel(tunnel *SSHTunnel) SSHClientOption {
	return func(c *SSHClient) {
		c.Tunnel = tunnel
	}
}

func SSHOptionWithPassword(password string) SSHClientOption {
	return func(c *SSHClient) {
		c.Password = password
	}
}

func SSHOptionWithLogger(logger types.Logger) SSHClientOption {
	return func(c *SSHClient) {
		c.Logger = logger
	}
}

func NewSSHClient(host string, port int, user string, privateKey string, quiet bool, opts ...SSHClientOption) *SSHClient {
	c := &SSHClient{
		Host:       host,
		Port:       port,
		User:       user,
		PrivateKey: privateKey,
		Quiet:      quiet,
		ChunkSize:  8192,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func newSSHClientWithProxy(proxyAddress, sshServerAddress string, sshConfig *ssh.ClientConfig) (*ssh.Client, error) {
	// dialer, err := proxy.SOCKS5("tcp", proxyAddress, nil, proxy.Direct)
	proxyUrl, err := url.Parse(proxyAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "newSSHClientWithProxy url.Parse %s failed", proxyAddress)
	}

	dialer, err := proxy.FromURL(proxyUrl, proxy.Direct)
	if err != nil {
		return nil, errors.Wrapf(err, "newSSHClientWithProxy proxy.FromURL %s failed", proxyAddress)
	}

	conn, err := dialer.Dial("tcp", sshServerAddress)
	if err != nil {
		return nil, errors.Wrapf(err, "newSSHClientWithProxy Dial %s failed", sshServerAddress)
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, sshServerAddress, sshConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "newSSHClientWithProxy NewClientConn %s failed", sshServerAddress)
	}

	return ssh.NewClient(c, chans, reqs), nil
}

func (s *SSHClient) Connect() error {
	if !s.Connected {
		if s.PrivateKey == "" && s.Password == "" {
			return errors.New("at least one of the plaintext password or RSA key must be set")
		}

		// var hostKey ssh.PublicKey
		var key []byte

		if strings.HasPrefix(s.PrivateKey, "~/") || strings.HasPrefix(s.PrivateKey, "/") {
			pkey, err := homedir.Expand(s.PrivateKey)
			if err != nil {
				return errors.Wrap(err, "Load private key path failed")
			}

			key, err = os.ReadFile(pkey)
			if err != nil {
				return errors.Wrapf(err, "Unable to read private key %s", s.PrivateKey)
			}
		} else {
			if s.PrivateKey != "" {
				if len(s.PrivateKey) < 256 {
					return errors.New("Invalid private key string")
				}

				key = []byte(s.PrivateKey)
			}
		}

		var authMethods []ssh.AuthMethod
		var err error
		var signer ssh.Signer

		if len(key) > 0 {
			signer, err = ssh.ParsePrivateKey(key)
			if err != nil {
				return errors.Wrapf(err, "Ubable to parse private key %s", s.PrivateKey)
			}

			authMethods = append(authMethods, ssh.PublicKeys(signer))
		} else {
			if s.Password != "" {
				authMethods = append(authMethods, ssh.Password(s.Password))
			}
		}

		config := &ssh.ClientConfig{
			User: s.User,
			Auth: authMethods,
			// HostKeyCallback: ssh.FixedHostKey(hostKey),
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Timeout:         s.Timeout,
		}

		// 检查 SSH 隧道配置 ...
		if s.Tunnel != nil {
			opened := make(chan bool)

			s.Tunnel.Config = config

			go s.Tunnel.Start(opened)

			<-opened

			// 若指定端口为0, 则重新读取本地端口号.
			if s.Port == 0 {
				s.Port = s.Tunnel.Local.Port
			}
		}

		var client *ssh.Client

		if s.Proxy != "" {
			client, err = newSSHClientWithProxy(s.Proxy, fmt.Sprintf("%s:%d", s.Host, s.Port), config)
		} else {
			client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port), config)
		}

		if err != nil {
			return errors.Wrapf(err, "Unable to connect %s:%d", s.Host, s.Port)
		}

		if !s.Quiet {
			glog.Infof("[SSH] Connected to host %s:%d", s.Host, s.Port)
		}

		s.Client = client
		s.Connected = true
	}

	return nil
}

func (s *SSHClient) Close() error {
	if s.Connected {
		if s.Tunnel != nil {
			s.Tunnel.Stop()
		}

		err := s.Client.Close()
		if err != nil {
			return errors.Wrap(err, "SSHClient Close failed")
		}
	}

	return nil
}

func (s *SSHClient) Run(command string) (int, error) {
	return s.RunWithWriter(command, nil)
}

func (s *SSHClient) RunWithWriter(command string, w io.Writer) (int, error) {
	err := s.Connect()
	if err != nil {
		return -1, errors.Wrap(err, "SSHClient RunWithWriter Connect failed")
	}

	session, err := s.Client.NewSession()
	if err != nil {
		return -2, errors.Wrap(err, "SSHClient RunWithWriter NewSession failed")
	}

	if s.TTY {
		// Set up terminal modes
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}

		err = session.RequestPty("xterm", 24, 80, modes)
		if err != nil {
			return -2, errors.Wrapf(err, "failed to set tty (%s:%d)", s.Host, s.Port)
		}
	}

	defer session.Close()

	stderr, _ := session.StderrPipe()
	stdout, _ := session.StdoutPipe()

	if err = session.Start(command); err != nil {
		return -3, errors.Wrap(err, "SSHClient RunWithWriter session Start failed")
	}

	var scanner *bufio.Scanner

	if w != nil {
		scanner = bufio.NewScanner(io.TeeReader(io.MultiReader(stdout, stderr), w))
	} else {
		scanner = bufio.NewScanner(io.MultiReader(stdout, stderr))
	}

	for scanner.Scan() {
		m := scanner.Text()

		if !s.Quiet {
			if s.Logger != nil {
				s.Logger.Info(m)
			} else {
				fmt.Fprintln(os.Stdout, m)
			}
		}
	}

	if err = session.Wait(); err != nil {
		var exiterr *ssh.ExitError
		if errors.As(err, &exiterr) {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			exitstatus := exiterr.ExitStatus()

			return exitstatus, errors.Wrapf(err, "SSHClient RunWithWriter Session Wait failed #%d", exitstatus)
		}
	}

	return 0, nil
}

func (s *SSHClient) Upload(src, dst string) error {
	err := s.Connect()
	if err != nil {
		return errors.Wrap(err, "SSHClient Upload Connect failed")
	}

	sftpClient, err := sftp.NewClient(s.Client)
	if err != nil {
		return errors.Wrap(err, "SSHClient Upload sftp.NewClient failed")
	}
	defer sftpClient.Close()

	dstFile, err := sftpClient.Create(dst)
	if err != nil {
		return errors.Wrapf(err, "SSHClient Upload sftp.Create %s failed", dst)
	}
	defer dstFile.Close()

	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "SSHClient Upload os.Open %s failed", src)
	}
	defer srcFile.Close()

	fileInfo, err := srcFile.Stat()
	if err != nil {
		return errors.Wrapf(err, "SSHClient Upload os.Stat %s failed", src)
	}

	totalByteCount := fileInfo.Size()
	readByteCount := 0

	buf := make([]byte, s.ChunkSize)

	isTty := isatty.IsTerminal(os.Stdout.Fd())

	for {
		n, err1 := srcFile.Read(buf)
		if err1 != nil {
			if err1 != io.EOF {
				return errors.Wrap(err1, "SSHClient Upload srcFile Read EOF")
			} else {
				readByteCount = readByteCount + n

				_, _ = dstFile.Write(buf[:n])

				if isTty && !s.Quiet {
					fmt.Printf("\r%.2f%%", float32(readByteCount)*100/float32(totalByteCount))
				}
				// logger.Debugf("readByteCount=%d, n=%d, %v", readByteCount, n, err)
				break
			}
		}

		readByteCount = readByteCount + n

		_, _ = dstFile.Write(buf[:n])

		if isTty && !s.Quiet {
			fmt.Printf("\r%.2f%%", float32(readByteCount)*100/float32(totalByteCount))
		}
	}

	if isTty && !s.Quiet {
		glog.Infof("Uploaded. (%s -> %s)", src, dst)
	}

	// s.Run(fmt.Sprintf("ls -lh %s", dst))

	return nil
}

func (s *SSHClient) Download(src, dst string) error {
	err := s.Connect()
	if err != nil {
		return errors.Wrap(err, "SSHClient Download Connect failed")
	}

	sftpClient, err := sftp.NewClient(s.Client)
	if err != nil {
		return errors.Wrap(err, "SSHClient Download sftp.NewClient failed")
	}
	defer sftpClient.Close()

	srcFile, err := sftpClient.Open(src)
	if err != nil {
		return errors.Wrapf(err, "SSHClient Download sftp.Open %s failed", src)
	}
	defer srcFile.Close()

	srcFileInfo, err := srcFile.Stat()
	if err != nil {
		return errors.Wrapf(err, "SSHClient Download sftp.Stat %s failed", src)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.Wrapf(err, "SSHClient Download os.Create %s failed", dst)
	}
	defer dstFile.Close()

	totalByteCount := srcFileInfo.Size()
	readByteCount := 0

	fmt.Println(totalByteCount)

	buf := make([]byte, s.ChunkSize)

	isTty := isatty.IsTerminal(os.Stdout.Fd())

	for {
		n, err1 := srcFile.Read(buf)
		if err1 != nil {
			if err1 != io.EOF {
				return errors.Wrap(err1, "SSHClient Download failed (EOF)")
			} else {
				readByteCount = readByteCount + n

				_, _ = dstFile.Write(buf[:n])

				if isTty && !s.Quiet {
					fmt.Printf("\r%.2f%%", float32(readByteCount)*100/float32(totalByteCount))
				}
				// logger.Debugf("readByteCount=%d, n=%d, %v", readByteCount, n, err)
				break
			}
		}

		readByteCount = readByteCount + n

		_, _ = dstFile.Write(buf[:n])

		if isTty && !s.Quiet {
			fmt.Printf("\r%.2f%%", float32(readByteCount)*100/float32(totalByteCount))
		}
	}

	// if _, err := srcFile.WriteTo(dstFile); err != nil {
	// 	return err
	// }

	if isTty && !s.Quiet {
		glog.Infof("Downloaded. (%s -> %s)", src, dst)
	}

	return nil
}
