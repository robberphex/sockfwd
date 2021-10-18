package cmd

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sockfwd",
	Short: "Forward between sockets",
	RunE:  runAction,
}

func init() {
	rootCmd.Flags().BoolP("quiet", "q", false, "Quiet mode")
	rootCmd.Flags().StringP("source", "s", "", "Source address")
	rootCmd.Flags().StringP("destination", "d", "", "Destination address")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	SIGINT  = os.Interrupt
	SIGKILL = os.Kill
	//allows compilation under windows,
	//even though it cannot send USR signals
	SIGUSR1 = syscall.Signal(0xa)
	SIGUSR2 = syscall.Signal(0xc)
	SIGTERM = syscall.Signal(0xf)
)

func listen(addr string) (net.Listener, error) {
	//listen
	network := "tcp"
	if strings.HasPrefix(addr, "unix:") {
		network = "unix"
		addr = strings.TrimPrefix(addr, "unix:")
	}
	return net.Listen(network, addr)
}

func runAction(cmd *cobra.Command, args []string) error {
	source := cmd.Flag("source").Value.String()
	destination := cmd.Flag("destination").Value.String()
	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return err
	}
	l, err := listen(source)
	if err != nil {
		return err
	}
	//cleanup before shutdown
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c)
		for sig := range c {
			switch sig {
			case SIGINT, SIGTERM, SIGKILL:
				l.Close()
				//os.Remove(config.SocketAddr)
				logrus.Info("closed listener and removed socket")
				os.Exit(0)
			case SIGUSR1:
				mem := runtime.MemStats{}
				runtime.ReadMemStats(&mem)
				logrus.Info("stats:\n"+
					"  %s, uptime: %s\n"+
					"  goroutines: %d, mem-alloc: %d\n"+
					"  connections open: %d total: %d",
					runtime.Version(), time.Now().Sub(uptime),
					runtime.NumGoroutine(), mem.Alloc,
					atomic.LoadInt64(&current), atomic.LoadUint64(&total))
			case SIGUSR2:
				//toggle logging with USR2 signal
				// config.Quiet = !config.Quiet
				// logf("connection logging: %v", config.Quiet)
			}
		}
	}()
	//accept connections
	logrus.Info("listening on " + source + " and forwarding to " + destination)
	for {
		uconn, err := l.Accept()
		if err != nil {
			logrus.Info("accept failed: %s", err)
			continue
		}
		go fwd(uconn, destination, quiet)
	}
}

//detailed statistics
var uptime = time.Now()
var total uint64
var current int64

//pool of buffers (default to io.Copy buffer size)
var pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

func dial(destination string) (net.Conn, error) {
	network := "tcp"
	if strings.HasPrefix(destination, "unix:") {
		network = "unix"
		destination = strings.TrimPrefix(destination, "unix:")
	}
	return net.Dial(network, destination)
}

func fwd(uconn net.Conn, destination string, quiet bool) {
	tconn, err := dial(destination)
	if err != nil {
		log.Printf("tcp dial failed: %s", err)
		uconn.Close()
		return
	}
	//stats
	atomic.AddUint64(&total, 1)
	atomic.AddInt64(&current, 1)
	//optional log
	if !quiet {
		logrus.Info("connection #%d (%d open)", atomic.LoadUint64(&total), atomic.LoadInt64(&current))
	}
	//pipe!
	go func() {
		ubuff := pool.Get().([]byte)
		io.CopyBuffer(uconn, tconn, ubuff)
		pool.Put(ubuff)
		uconn.Close()
		//stats
		atomic.AddInt64(&current, -1)
	}()
	go func() {
		tbuff := pool.Get().([]byte)
		io.CopyBuffer(tconn, uconn, tbuff)
		pool.Put(tbuff)
		tconn.Close()
	}()
}
