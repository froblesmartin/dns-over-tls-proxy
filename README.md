# DNS to DNS-over-TLS proxy

This proxy can help with your applications that do not handle DNS-over-TLS. It will listen for regular DNS requests on port 53 both UDP and TCP and then translate that into a TCP DNS-over-TLS request to the DNS provider configured.

## Usage

If you want to run the proxy in your own setup:

```shell
go install
dot-proxy
```

If you want to run it in a container:

```shell
docker build . -t dot-proxy
docker run --rm -p 53:53/tcp -p 53:53/udp -it dot-proxy
```

> [!WARNING] If running it in MacOS, you need to disable (set to false) the `kernelForUDP` configuration of Docker Desktop as mentioned in the `Known issues of version 4.24.0`: https://docs.docker.com/desktop/release-notes/#4240

### Configuration

You can set the following environment variables to change the configuration of the proxy:

```shell
DOT_PROXY_DOTPORT=<defaults to 853>
DOT_PROXY_DOTSERVER=<defaults to one.one.one.one>
DOT_PROXY_LISTENINGPORT=<defaults to 53>
```

#### DNS-over-TLS providers

##### Cloudflare

[Official Cloudflare documentation about DNS-over-TLS](https://developers.cloudflare.com/1.1.1.1/encryption/dns-over-tls/)

Cloudflare supports DNS over TLS on standard port 853 and is compliant with RFC 7858. It can be reached on the IPv4 addresses `1.1.1.1`, `1.0.0.1`, and the IPv6 addresses `2606:4700:4700::1111` and `2606:4700:4700::1001`, or the DNS `one.one.one.one`.

##### Google

[Official Google documentation about DNS-over-TLS](https://developers.google.com/speed/public-dns/docs/dns-over-tls)

It can be used at `dns.google`.

### Testing the proxy

To test a TCP DNS Query any of these CLI tools can be used:

```shell
nslookup -vc example.com 127.0.0.1
host -T example.com 127.0.0.1
dig @127.0.0.1 example.com +tcp
```

If UDP wants to be tested, then:

```shell
nslookup example.com 127.0.0.1
host example.com 127.0.0.1
dig @127.0.0.1 example.com
```

### Security concerns

The following things need to be taken into account when using it:

- The connection between the original DNS Query application and this DoT Proxy is still unprotected. This means that someone/something with access to this network could still monitor these requests.

### Suggestions on how to integrate it in your distributed, microservices-oriented and containerized architecture

You can configure your distributed architecture to send all DNS Queries to this proxy. For example, if you are using Kubernetes, you can [configure CoreDNS to send the DNS Queries](https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/) to this proxy. Then, configure this proxy to contact the DNS-over-TLS provider with their static IPs to avoid a circular dependency, so that the DoT proxy does not need to resolve first its own DNS query.

## Contributors

Embrace open-source! :D

### Future features and improvements

- The proxy should assume that the TCP clients will initiate connection closing, and should delay closing its end of the connection until all outstanding client requests have been satisfied.
- If the proxy needs to close a dormant connection to reclaim resources, it should wait until the connection has been idle for a period on the order of two minutes.
- Allow retrying to other DoT servers if the connection fails with the main configured one.
- Write logic to retry queries to the DoT server with a retransmission interval of 2-5 seconds as per RFC1035
- Add tests and CI

### Interesting/related reads

- [About DNS Response Size](https://www.netmeister.org/blog/dns-size.html)

## Author

Francisco Robles Martín (froblesmartin)
