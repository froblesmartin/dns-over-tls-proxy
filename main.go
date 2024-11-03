package main

import (
	"crypto/tls"
	"net"
	"os"

	"froblesmartin/dot-proxy/config"

	"github.com/miekg/dns"
	"github.com/spf13/viper"
	"golang.org/x/exp/slog"
)

func main() {
	config.InitializeConfig()

	var programLevel = new(slog.LevelVar)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	// Uncomment to enable DEBUG logs
	programLevel.Set(slog.LevelDebug)

	go udpListener()
	tcpListener()
}

func udpListener() {
	listener, err := net.ListenPacket("udp", ":53")
	if err != nil {
		panic("Error: " + err.Error())
	}
	slog.Info(
		"Listening for packets",
		"Network", "UDP",
		"Port", viper.GetString("ListeningPort"),
	)
	defer listener.Close()

	for {
		dnsQuery := make([]byte, dns.MaxMsgSize)
		dnsQueryLength, remoteAddr, err := listener.ReadFrom(dnsQuery)
		if err != nil {
			slog.Error("Error reading DNS query",
				"Error", err)
			return
		}
		go udpProcessQuery(listener, dnsQuery[:dnsQueryLength], remoteAddr)
	}
}

func udpProcessQuery(listener net.PacketConn, dnsQuery []byte, remoteAddr net.Addr) {
	slog.Info("Received UDP DNS Query")

	dnsQueryWithLengthHeader := append([]byte{0, byte(len(dnsQuery))}, dnsQuery...)

	responseReceived, err := sendTcpDnsQuery(dnsQueryWithLengthHeader)
	if err != nil {
		slog.Error("Error sending DNS Query to target DoT server",
			"Error", err)
		return
	}

	listener.WriteTo(responseReceived[2:], remoteAddr)
	slog.Info("DNS response sent to the UDP querier")
}

func tcpListener() {
	listener, err := net.Listen("tcp", ":"+viper.GetString("ListeningPort"))
	if err != nil {
		slog.Error("Error listening",
			"Network", "TCP",
			"Port", viper.GetString("ListeningPort"))
		panic(err)
	}
	slog.Info(
		"Listening for packets",
		"Network", "TCP",
		"Port", viper.GetString("ListeningPort"),
	)
	defer listener.Close()

	for {
		incomingConnection, err := listener.Accept()
		if err != nil {
			slog.Error("Error accepting incoming connection",
				"Error", err)
			continue
		}

		go tcpProcessQuery(incomingConnection)
	}
}

func tcpProcessQuery(incomingConnection net.Conn) {
	slog.Info("Incoming TCP connection established")
	dnsQuery := make([]byte, dns.MaxMsgSize)
	dnsQueryLength, err := incomingConnection.Read(dnsQuery)
	if err != nil {
		slog.Error("Error reading DNS query",
			"Error", err)
		return
	}

	responseReceived, err := sendTcpDnsQuery(dnsQuery[:dnsQueryLength])
	if err != nil {
		slog.Error("Error sending DNS Query to target DoT server",
			"Error", err)
		return
	}

	incomingConnection.Write(responseReceived)
	slog.Info("DNS response sent to the TCP querier")
	incomingConnection.Close()
}

func sendTcpDnsQuery(dnsQuery []byte) ([]byte, error) {
	checkAndPrintDnsInfo(dnsQuery[2:])

	dotServerAddr := viper.GetString("DoTServer") + ":" + viper.GetString("DoTPort")

	slog.Info("Starting connection with remote DoT server",
		"addr", dotServerAddr)
	connectionWithDotServer, err := tls.Dial("tcp", dotServerAddr, &tls.Config{})
	if err != nil {
		slog.Error("Error establishing TLS connection with DoT server",
			"Error", err)
		return nil, err
	}

	err = connectionWithDotServer.VerifyHostname(viper.GetString("DoTServer"))
	if err != nil {
		slog.Error("The DoT server certificate is not valid",
			"Error", err)
		return nil, err
	}

	_, err = connectionWithDotServer.Write(dnsQuery)
	if err != nil {
		slog.Error("Error sending request to the DoT server",
			"Error", err)
		return nil, err
	}

	responseReceived := make([]byte, dns.MaxMsgSize)
	responseLength, err := connectionWithDotServer.Read(responseReceived)
	if err != nil {
		slog.Error("Error reading DoT response",
			"Error", err)
		return nil, err
	}

	checkAndPrintDnsInfo(responseReceived[2:responseLength])

	slog.Info("DNS response received from the DoT server")
	connectionWithDotServer.Close()
	return responseReceived[:responseLength], nil
}

func checkAndPrintDnsInfo(dnsQuery []byte) {
	dnsQueryMessage := new(dns.Msg)
	err := dnsQueryMessage.Unpack(dnsQuery)
	slog.Debug("Valid DNS packet sanity check",
		"null is OK", dns.IsMsg(dnsQuery))
	if err != nil {
		slog.Error("Error unpacking the DNS message",
			"Error", err)
		return
	}
	slog.Debug("DNS message info",
		"all", dnsQueryMessage)
}
