package proxy

import "io"

func singleTunnel(w io.Writer, r io.Reader, ch chan<- struct{}) {
	io.Copy(w, r)
	ch <- struct{}{}
}

func Tunnel(a, b io.ReadWriter) {
	doneCh := make(chan struct{}, 2)
	go singleTunnel(a, b, doneCh)
	go singleTunnel(b, a, doneCh)
	<-doneCh
}
