package net

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/designinlife/slib/errors"

	"golang.org/x/crypto/ssh"
	"golang.org/x/net/proxy"

	// optional SFTP; keep so upload/download easier
	"github.com/pkg/sftp"
)

type RichSSHClientOption func(c *RichSSHClient)

type RichSSHClient struct {
	Host           string
	Port           int
	User           string
	Password       string
	PrivateKeyFile string
	PrivateKey     []byte // PEM bytes; if both Password and PrivateKey provided, use PrivateKey
	JumpSSHHost    string
	JumpSSHPort    int
	ProxyURL       string

	EnablePTY bool // default false

	// internals
	mu          sync.Mutex
	client      *ssh.Client
	jumpClient  *ssh.Client
	sftpClient  *sftp.Client
	closed      bool
	dialTimeout time.Duration
}

type RichSSHClientResponse struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

// NewRichSSHClient builds client with defaults and options.
func NewRichSSHClient(host string, port int, user string, opts ...RichSSHClientOption) *RichSSHClient {
	c := &RichSSHClient{
		Host:        host,
		Port:        port,
		User:        user,
		dialTimeout: 15 * time.Second,
	}
	for _, opt := range opts {
		opt(c)
	}
	// Do not auto connect. Connect on first Run/RunStream or user can call Connect
	return c
}

func WithPassword(pw string) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.Password = pw
	}
}
func WithPrivateKeyPEM(pem []byte) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.PrivateKey = pem
	}
}
func WithPrivateKeyFile(path string) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.PrivateKeyFile = path
	}
}
func WithJumpHost(host string, port int) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.JumpSSHHost = host
		c.JumpSSHPort = port
	}
}
func WithProxyURL(proxyURL string) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.ProxyURL = proxyURL
	}
}
func WithPTY(enable bool) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.EnablePTY = enable
	}
}
func WithDialTimeout(d time.Duration) RichSSHClientOption {
	return func(c *RichSSHClient) {
		c.dialTimeout = d
	}
}

// Connect establishes SSH connection (handles jump/proxy)
func (c *RichSSHClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("client closed")
	}
	if c.client != nil {
		return nil // already connected
	}

	if len(c.PrivateKey) == 0 && c.PrivateKeyFile != "" {
		bPrivKey, err := os.ReadFile(c.PrivateKeyFile)
		if err != nil {
			return errors.Wrapf(err, "Unable to read private key %s", c.PrivateKeyFile)
		}
		c.PrivateKey = bPrivKey
	}

	// prepare auth methods
	var auths []ssh.AuthMethod
	if len(c.PrivateKey) > 0 {
		signer, err := ssh.ParsePrivateKey(c.PrivateKey)
		if err != nil {
			return fmt.Errorf("parse private key: %w", err)
		}
		auths = append(auths, ssh.PublicKeys(signer))
	} else if c.Password != "" {
		auths = append(auths, ssh.Password(c.Password))
	}

	sshConfig := &ssh.ClientConfig{
		User:            c.User,
		Auth:            auths,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // production: replace
		Timeout:         c.dialTimeout,
	}

	// If jump host specified, connect jump first
	var baseConn net.Conn
	var err error

	// choose dial function depending on proxy
	dialFunc := c.netDialContextWithProxy

	targetAddr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	if c.JumpSSHHost != "" {
		// connect jump host (may still need proxy)
		jumpAddr := fmt.Sprintf("%s:%d", c.JumpSSHHost, c.JumpSSHPort)
		jumpConn, err := dialFunc(ctx, "tcp", jumpAddr)
		if err != nil {
			return fmt.Errorf("dial jump host: %w", err)
		}
		// upgrade to ssh client over jumpConn
		jumpClientConn, chans, reqs, err := ssh.NewClientConn(jumpConn, jumpAddr, sshConfig)
		if err != nil {
			jumpConn.Close()
			return fmt.Errorf("new client conn to jump host: %w", err)
		}
		c.jumpClient = ssh.NewClient(jumpClientConn, chans, reqs)

		// now from jumpClient, dial target
		conn, err := c.jumpClient.Dial("tcp", targetAddr)
		if err != nil {
			return fmt.Errorf("dial target from jump: %w", err)
		}
		baseConn = conn
	} else {
		// direct dial (via proxy if present)
		baseConn, err = dialFunc(ctx, "tcp", targetAddr)
		if err != nil {
			return fmt.Errorf("dial target: %w", err)
		}
	}

	// create SSH client over baseConn
	nc := baseConn
	clientConn, chans, reqs, err := ssh.NewClientConn(nc, targetAddr, sshConfig)
	if err != nil {
		nc.Close()
		return fmt.Errorf("ssh new client conn: %w", err)
	}
	c.client = ssh.NewClient(clientConn, chans, reqs)
	return nil
}

// netDialContextWithProxy supports socks5:// and http(s):// CONNECT, otherwise direct dial
func (c *RichSSHClient) netDialContextWithProxy(ctx context.Context, network, addr string) (net.Conn, error) {
	// if no proxy -> normal dial
	if c.ProxyURL == "" {
		d := &net.Dialer{Timeout: c.dialTimeout}
		return d.DialContext(ctx, network, addr)
	}
	u, err := url.Parse(c.ProxyURL)
	if err != nil {
		return nil, err
	}
	switch strings.ToLower(u.Scheme) {
	case "socks5", "socks5h", "socks":
		// use golang.org/x/net/proxy
		var auth *proxy.Auth
		if u.User != nil {
			pw, _ := u.User.Password()
			auth = &proxy.Auth{User: u.User.Username(), Password: pw}
		}
		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}
		conn, err := dialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}
		return conn, nil
	case "http", "https":
		// implement simple HTTP CONNECT
		d := &net.Dialer{Timeout: c.dialTimeout}
		proxyConn, err := d.DialContext(ctx, "tcp", u.Host)
		if err != nil {
			return nil, err
		}

		// send CONNECT
		req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", addr, addr)
		if u.User != nil {
			pw, _ := u.User.Password()
			auth := "Basic " + basicAuth(u.User.Username(), pw)
			req += "Proxy-Authorization: " + auth + "\r\n"
		}
		req += "\r\n"
		if _, err := proxyConn.Write([]byte(req)); err != nil {
			proxyConn.Close()
			return nil, err
		}
		// read response minimal
		buf := make([]byte, 1024)
		n, err := proxyConn.Read(buf)
		if err != nil {
			proxyConn.Close()
			return nil, err
		}
		resp := string(buf[:n])
		if !strings.Contains(resp, "200") {
			proxyConn.Close()
			return nil, fmt.Errorf("proxy connect failed: %s", strings.SplitN(resp, "\r\n", 2)[0])
		}
		return proxyConn, nil
	default:
		// unsupported scheme
		return nil, fmt.Errorf("unsupported proxy scheme %s", u.Scheme)
	}
}

func basicAuth(user, pass string) string {
	enc := user + ":" + pass
	return base64.StdEncoding.EncodeToString([]byte(enc))
}

// Close closes everything
func (c *RichSSHClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	if c.sftpClient != nil {
		_ = c.sftpClient.Close()
		c.sftpClient = nil
	}
	if c.client != nil {
		_ = c.client.Close()
		c.client = nil
	}
	if c.jumpClient != nil {
		_ = c.jumpClient.Close()
		c.jumpClient = nil
	}
}

// Run executes a command and returns output and exit code
func (c *RichSSHClient) Run(ctx context.Context, cmd string) (*RichSSHClientResponse, error) {
	if err := c.Connect(ctx); err != nil {
		return nil, err
	}
	sess, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer sess.Close()

	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf

	// request PTY if enabled
	if c.EnablePTY {
		modes := ssh.TerminalModes{
			ssh.ECHO:          1,
			ssh.TTY_OP_ISPEED: 14400,
			ssh.TTY_OP_OSPEED: 14400,
		}
		if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
			// if PTY request fails, continue without PTY or return? we return error.
			return nil, fmt.Errorf("request pty: %w", err)
		}
	}

	// start and wait, respect ctx
	errCh := make(chan error, 1)
	go func() {
		errCh <- sess.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		return nil, ctx.Err()
	case err := <-errCh:
		resp := &RichSSHClientResponse{Stdout: outBuf.Bytes(), Stderr: errBuf.Bytes()}
		if err == nil {
			resp.ExitCode = 0
			return resp, nil
		}
		// try to get exit status
		var ee *ssh.ExitError
		if errors.As(err, &ee) {
			resp.ExitCode = ee.ExitStatus()
			return resp, nil
		}
		return resp, err
	}
}

// RunStream runs command, piping stdout/stderr to the provided writers in real time.
// outWriter and errWriter can be nil. PTY respects EnablePTY.
func (c *RichSSHClient) RunStream(ctx context.Context, cmd string, outWriter, errWriter io.Writer) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	sess, err := c.client.NewSession()
	if err != nil {
		return err
	}
	// not deferring sess.Close() here because we need to ensure Wait done before Close
	// We'll close at the end.
	if c.EnablePTY {
		modes := ssh.TerminalModes{ssh.ECHO: 1}
		if err := sess.RequestPty("xterm", 80, 40, modes); err != nil {
			_ = sess.Close()
			return fmt.Errorf("request pty: %w", err)
		}
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		return err
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		sess.Close()
		return err
	}

	if err := sess.Start(cmd); err != nil {
		sess.Close()
		return err
	}

	copyErrCh := make(chan error, 2)
	go func() {
		if outWriter != nil {
			_, err := io.Copy(outWriter, stdout)
			copyErrCh <- err
		} else {
			_, _ = io.Copy(io.Discard, stdout)
			copyErrCh <- nil
		}
	}()
	go func() {
		if errWriter != nil {
			_, err := io.Copy(errWriter, stderr)
			copyErrCh <- err
		} else {
			_, _ = io.Copy(io.Discard, stderr)
			copyErrCh <- nil
		}
	}()

	done := make(chan error, 1)
	go func() {
		done <- sess.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		_ = sess.Close()
		return ctx.Err()
	case err := <-done:
		// ensure copies finished
		<-copyErrCh
		<-copyErrCh
		_ = sess.Close()
		return err
	}
}

// ensureSFTP initializes sftp client lazily
func (c *RichSSHClient) ensureSFTP() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sftpClient != nil {
		return nil
	}
	if c.client == nil {
		return errors.New("ssh client not connected")
	}
	sftpClient, err := sftp.NewClient(c.client)
	if err != nil {
		return err
	}
	c.sftpClient = sftpClient
	return nil
}

// UploadFile uploads localPath -> remotePath. If progressWriter != nil, it will be written with bytes transferred.
func (c *RichSSHClient) UploadFile(ctx context.Context, localPath, remotePath string, progressWriter io.Writer) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	if err := c.ensureSFTP(); err != nil {
		return err
	}

	srcFile, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	fi, err := srcFile.Stat()
	if err != nil {
		return err
	}
	remoteDir := filepath.Dir(remotePath)
	_ = c.sftpClient.MkdirAll(remoteDir)

	dstFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	var reader io.Reader = srcFile
	if progressWriter != nil {
		reader = &progressReader{r: srcFile, total: fi.Size(), sink: progressWriter}
	}
	_, err = io.Copy(dstFile, reader)
	return err
}

// DownloadFile downloads remotePath -> localPath. If progressWriter != nil, it will be written with bytes transferred.
func (c *RichSSHClient) DownloadFile(ctx context.Context, remotePath, localPath string, progressWriter io.Writer) error {
	if err := c.Connect(ctx); err != nil {
		return err
	}
	if err := c.ensureSFTP(); err != nil {
		return err
	}

	srcFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// create local directory
	_ = os.MkdirAll(filepath.Dir(localPath), 0o755)
	dstFile, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	var reader io.Reader = srcFile
	if progressWriter != nil {
		var size int64
		if fi, err := srcFile.Stat(); err == nil {
			size = fi.Size()
		}
		reader = &progressReader{r: srcFile, total: size, sink: progressWriter}
	}
	_, err = io.Copy(dstFile, reader)
	return err
}

type progressReader struct {
	r     io.Reader
	total int64
	read  int64
	sink  io.Writer
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.read += int64(n)
		if p.sink != nil {
			// write simple "read/total\n"
			if p.total > 0 {
				fmt.Fprintf(p.sink, "%d/%d\n", p.read, p.total)
			} else {
				fmt.Fprintf(p.sink, "%d\n", p.read)
			}
		}
	}
	return n, err
}

// small utility guard: to get exit status from Wait error - handled in Run

// Helper: minimal implementation
// NOTE: replace encodeBase64 placeholder with real base64 in production
