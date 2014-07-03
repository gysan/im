package tunnel

type TunnelLet interface {
	MessageReceived(tunnelData TunnelData)
}
